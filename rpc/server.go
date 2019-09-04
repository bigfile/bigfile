//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

// Package rpc is used to implement rpc protocol
package rpc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/bigfile/bigfile/databases"
	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/log"
	"github.com/bigfile/bigfile/service"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var (
	isTesting    = false
	testDbConn   *gorm.DB
	testRootPath *string
	// ErrGetIPFailed represent that get ip failed
	ErrGetIPFailed = errors.New("getClientIP failed")

	// ErrAppSecret represent appUID and appSecret doesn't match
	ErrAppSecret = errors.New("appUID and appSecret doesn't match")

	// ErrTokenNotMatchApp represent that the token doesn't belong to this app
	ErrTokenNotMatchApp = errors.New("the token doesn't belong to this app")

	// ErrTokenSecretWrong the secret of token is wrong
	ErrTokenSecretWrong = errors.New("the secret of token is wrong")

	// ErrDirShouldNotHasContent represent that create a directory with content
	ErrDirShouldNotHasContent = errors.New("the directory should not has content")
)

// Server is used to create a rpc server
type Server struct{}

// getClientIP is used to get client ip from the context
func (s *Server) getClientIP(ctx context.Context) (string, error) {
	var (
		pr      *peer.Peer
		ok      bool
		ipV4    net.IP
		tcpAddr *net.TCPAddr
	)
	if pr, ok = peer.FromContext(ctx); !ok {
		return "", ErrGetIPFailed
	}
	if tcpAddr, ok = pr.Addr.(*net.TCPAddr); ok && tcpAddr != nil {
		if tcpAddr.IP.IsLoopback() {
			return "127.0.0.1", nil
		}
		ipV4 = tcpAddr.IP.To4()
		if ipV4 == nil {
			return tcpAddr.IP.To16().String(), nil
		}
		return ipV4.String(), nil
	}
	return pr.Addr.String(), nil
}

// fetchAPP is used to generate *models.APP by app?UID and APPSecret
func fetchAPP(appUID, APPSecret string, db *gorm.DB) (app *models.App, err error) {
	app = &models.App{}
	err = db.Where("uid = ? and secret = ?", appUID, APPSecret).First(app).Error
	if gorm.IsRecordNotFoundError(err) {
		err = ErrAppSecret
	}
	return
}

// generateRequestRecord is used to generate request record
func (s *Server) generateRequestRecord(ctx context.Context, service string, request interface{}, db *gorm.DB) (record *models.Request, err error) {
	var (
		ip          string
		requestBody string
		requestMD   string
	)
	if ip, err = s.getClientIP(ctx); err != nil {
		return
	}

	if requestBody, err = jsoniter.MarshalToString(request); err != nil {
		return record, err
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if requestMD, err = jsoniter.MarshalToString(md); err != nil {
			return record, err
		}
	}

	record = &models.Request{
		Protocol:      "rpc",
		IP:            &ip,
		Service:       &service,
		RequestBody:   requestBody,
		RequestHeader: requestMD,
	}
	err = db.Create(record).Error
	return
}

func (s *Server) tokenResp(token *models.Token) (t *Token) {
	t = &Token{Token: token.UID, Path: token.Path, AvailableTimes: int32(token.AvailableTimes)}
	if token.IP != nil {
		t.Ip = &wrappers.StringValue{Value: *token.IP}
	}
	if token.ReadOnly == 1 {
		t.ReadOnly = true
	}
	if token.Secret != nil {
		t.Secret = &wrappers.StringValue{Value: *token.Secret}
	}
	if token.ExpiredAt != nil {
		ts, _ := ptypes.TimestampProto(*token.ExpiredAt)
		t.ExpiredAt = ts
	}
	if token.DeletedAt != nil {
		ts, _ := ptypes.TimestampProto(*token.DeletedAt)
		t.DeletedAt = ts
	}
	return
}

func (s *Server) fileResp(file *models.File, db *gorm.DB) (f *File, err error) {
	var path string
	if path, err = file.Path(db.Unscoped()); err != nil {
		return nil, err
	}
	f = &File{
		Uid:  file.UID,
		Path: path,
		Size: uint64(file.Size),
	}
	if file.Hidden == 1 {
		f.Hidden = true
	}
	if file.IsDir == 1 {
		f.IsDir = true
	} else {
		if file.Object.ID == 0 {
			db.Unscoped().Preload("Object").First(file)
		}
		f.Hash = &wrappers.StringValue{Value: file.Object.Hash}
		f.Ext = &wrappers.StringValue{Value: file.Ext}
	}
	if file.DeletedAt != nil {
		if f.DeletedAt, err = ptypes.TimestampProto(*file.DeletedAt); err != nil {
			return f, err
		}
	}
	return f, err
}

func (s *Server) updateRequestRecord(ctx context.Context, request *models.Request, resp interface{}, err error, db *gorm.DB) {
	var responseBody string

	request.ResponseCode = int(codes.OK)
	if err != nil {
		responseBody = err.Error()
		request.ResponseCode = int(codes.InvalidArgument)
	} else {
		if responseBody, err = jsoniter.MarshalToString(resp); err != nil {
			log.MustNewLogger(nil).Error(err)
			return
		}
	}
	request.ResponseBody = responseBody
	if err = db.Save(request).Error; err != nil {
		log.MustNewLogger(nil).Error(err)
	}
}

func getDbConn() (db *gorm.DB) {
	if isTesting {
		db = testDbConn
	} else {
		db = databases.MustNewConnection(nil)
	}
	return
}

// TokenCreate is used to crate token
func (s *Server) TokenCreate(ctx context.Context, req *TokenCreateRequest) (resp *TokenCreateResponse, err error) {
	var (
		ip             *string
		db             = getDbConn()
		app            *models.App
		path           = "/"
		record         *models.Request
		secret         *string
		readOnly       int8
		expiredAt      *time.Time
		tokenCreateSrv *service.TokenCreate
		tokenCreateVal interface{}
		availableTimes = -1
	)
	defer func() {
		if err != nil {
			err = status.Error(codes.InvalidArgument, err.Error())
		}
	}()
	if record, err = s.generateRequestRecord(ctx, "TokenCreate", req, db); err != nil {
		return resp, err
	}
	resp = &TokenCreateResponse{RequestId: record.ID}
	defer func() { s.updateRequestRecord(ctx, record, resp, err, db) }()
	if app, err = fetchAPP(req.AppUid, req.AppSecret, db); err != nil {
		return
	}
	record.AppID = &app.ID
	if s := req.GetSecret(); s != nil {
		sv := s.GetValue()
		secret = &sv
	}
	if p := req.GetPath(); p != nil {
		path = p.GetValue()
	}
	if i := req.GetIp(); i != nil {
		ipv := i.GetValue()
		ip = &ipv
	}
	if r := req.ReadOnly; r != nil && r.GetValue() {
		readOnly = 1
	}
	if e := req.GetExpiredAt(); e != nil {
		seconds := req.ExpiredAt.GetSeconds()
		nsec := req.ExpiredAt.GetNanos()
		if seconds != 0 {
			exp := time.Unix(seconds, int64(nsec))
			expiredAt = &exp
		}
	}
	if a := req.GetAvailableTimes(); a != nil {
		availableTimes = int(a.GetValue())
	}

	tokenCreateSrv = &service.TokenCreate{
		BaseService: service.BaseService{
			DB: db,
		},
		IP:             ip,
		App:            app,
		Path:           path,
		Secret:         secret,
		ReadOnly:       readOnly,
		ExpiredAt:      expiredAt,
		AvailableTimes: availableTimes,
	}
	if err = tokenCreateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		return
	}

	if tokenCreateVal, err = tokenCreateSrv.Execute(ctx); err != nil {
		return
	}
	resp.Token = s.tokenResp(tokenCreateVal.(*models.Token))
	return resp, nil
}

// TokenUpdate is used to update some token
func (s *Server) TokenUpdate(ctx context.Context, req *TokenUpdateRequest) (resp *TokenUpdateResponse, err error) {
	var (
		ip             *string
		db             = getDbConn()
		app            *models.App
		path           *string
		secret         *string
		token          *models.Token
		record         *models.Request
		readOnly       *int8
		expiredAt      *time.Time
		tokenUpdateSrv *service.TokenUpdate
		tokenUpdateVal interface{}
		availableTimes *int
	)
	defer func() {
		if err != nil {
			err = status.Error(codes.InvalidArgument, err.Error())
		}
	}()
	if record, err = s.generateRequestRecord(ctx, "TokenUpdate", req, db); err != nil {
		return resp, err
	}
	resp = &TokenUpdateResponse{RequestId: record.ID}
	defer func() { s.updateRequestRecord(ctx, record, resp, err, db) }()
	if app, err = fetchAPP(req.AppUid, req.AppSecret, db); err != nil {
		return
	}
	record.AppID = &app.ID
	if token, err = models.FindTokenByUID(req.Token, db); err != nil {
		return
	}
	if token.AppID != app.ID {
		return resp, ErrTokenNotMatchApp
	}
	if s := req.GetSecret(); s != nil {
		sv := s.GetValue()
		secret = &sv
	}
	if p := req.GetPath(); p != nil {
		pv := p.GetValue()
		path = &pv
	}
	if i := req.GetIp(); i != nil {
		ipv := i.GetValue()
		ip = &ipv
	}
	if r := req.ReadOnly; r != nil && r.GetValue() {
		ro := int8(1)
		readOnly = &ro
	}
	if e := req.GetExpiredAt(); e != nil {
		seconds := req.ExpiredAt.GetSeconds()
		nsec := req.ExpiredAt.GetNanos()
		if seconds != 0 {
			exp := time.Unix(seconds, int64(nsec))
			expiredAt = &exp
		}
	}
	if a := req.GetAvailableTimes(); a != nil {
		av := int(a.GetValue())
		availableTimes = &av
	}

	tokenUpdateSrv = &service.TokenUpdate{
		BaseService:    service.BaseService{DB: db},
		Token:          token.UID,
		IP:             ip,
		Path:           path,
		Secret:         secret,
		ReadOnly:       readOnly,
		ExpiredAt:      expiredAt,
		AvailableTimes: availableTimes,
	}
	if err = tokenUpdateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		return
	}
	if tokenUpdateVal, err = tokenUpdateSrv.Execute(ctx); err != nil {
		return
	}
	resp.Token = s.tokenResp(tokenUpdateVal.(*models.Token))
	return resp, nil
}

// TokenDelete is used to delete some tokens
func (s *Server) TokenDelete(ctx context.Context, req *TokenDeleteRequest) (resp *TokenDeleteResponse, err error) {
	var (
		db     = getDbConn()
		app    *models.App
		token  *models.Token
		record *models.Request
	)
	defer func() {
		if err != nil {
			err = status.Error(codes.InvalidArgument, err.Error())
		}
	}()
	if record, err = s.generateRequestRecord(ctx, "TokenDelete", req, db); err != nil {
		return resp, err
	}
	resp = &TokenDeleteResponse{RequestId: record.ID}
	if app, err = fetchAPP(req.AppUid, req.AppSecret, db); err != nil {
		return
	}
	record.AppID = &app.ID
	if token, err = models.FindTokenByUID(req.Token, db); err != nil {
		return
	}
	if token.AppID != app.ID {
		return resp, ErrTokenNotMatchApp
	}
	if err = db.Delete(token).Error; err != nil {
		return
	}
	db.Delete(token)
	db.Unscoped().First(token)
	resp.Token = s.tokenResp(token)
	return
}

func (s *Server) fetchToken(t string, secret *wrappers.StringValue, db *gorm.DB) (token *models.Token, err error) {
	if token, err = models.FindTokenByUID(t, db); err != nil {
		return nil, err
	}
	if token.Secret != nil {
		if secret == nil || secret.GetValue() != *token.Secret {
			return nil, ErrTokenSecretWrong
		}
	}
	return token, nil
}

// FileCreate is used to upload file in a stream
func (s *Server) FileCreate(stream FileCreate_FileCreateServer) (err error) {
	var (
		db              = getDbConn()
		ctx             = stream.Context()
		req             *FileCreateRequest
		resp            *FileCreateResponse
		token           *models.Token
		record          *models.Request
		previousToken   string
		tokenHasChecked bool
		fileCreateSrv   *service.FileCreate
		fileCreateVal   interface{}
	)
	defer func() {
		if err != nil {
			err = status.Error(codes.InvalidArgument, err.Error())
		}
	}()
	for {
		handler := func() (err error) {
			if req, err = stream.Recv(); err != nil {
				return
			}
			content := req.Content
			req.Content = nil
			if record, err = s.generateRequestRecord(ctx, "FileCreate", req, db); err != nil {
				return
			}
			req.Content = content
			resp = &FileCreateResponse{RequestId: record.ID}
			if !tokenHasChecked || previousToken != req.Token {
				if token, err = s.fetchToken(req.Token, req.Secret, db); err != nil {
					return
				}
			}
			record.AppID = &token.App.ID
			record.Token = &token.UID
			fileCreateSrv = &service.FileCreate{
				BaseService: service.BaseService{DB: db, RootPath: testRootPath},
				Token:       token,
				Path:        req.Path,
				IP:          record.IP,
			}
			if req.GetCreateDir() {
				if req.GetContent() != nil {
					return ErrDirShouldNotHasContent
				}
			} else {
				var content []byte
				if req.Content != nil {
					content = req.Content.GetValue()
				}
				fileCreateSrv.Reader = bytes.NewReader(content)
			}
			if req.GetOverwrite() {
				fileCreateSrv.Overwrite = 1
			}
			if req.GetAppend() {
				fileCreateSrv.Append = 1
			}
			if req.GetRename() {
				fileCreateSrv.Rename = 1
			}
			if req.Hidden != nil && req.Hidden.GetValue() {
				fileCreateSrv.Hidden = 1
			}
			if err = fileCreateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
				return
			}
			if fileCreateVal, err = fileCreateSrv.Execute(ctx); err != nil {
				return
			}
			if resp.File, err = s.fileResp(fileCreateVal.(*models.File), db); err != nil {
				return
			}
			return stream.Send(resp)
		}
		if err = handler(); err != nil {
			return
		}
	}
}

// FileUpdate is used to update a file
func (s *Server) FileUpdate(ctx context.Context, req *FileUpdateRequest) (resp *FileUpdateResponse, err error) {
	var (
		db            = getDbConn()
		file          *models.File
		token         *models.Token
		record        *models.Request
		fileUpdateSrv *service.FileUpdate
		fileUpdateVal interface{}
	)
	defer func() {
		if err != nil {
			err = status.Error(codes.InvalidArgument, err.Error())
		}
	}()
	if record, err = s.generateRequestRecord(ctx, "FileUpdate", req, db); err != nil {
		return
	}
	resp = &FileUpdateResponse{RequestId: record.ID}
	if token, err = s.fetchToken(req.Token, req.Secret, db); err != nil {
		return
	}
	record.AppID = &token.App.ID
	record.Token = &token.UID
	if file, err = models.FindFileByUID(req.FileUid, false, db); err != nil {
		return
	}
	var hidden int8
	if req.GetHidden() != nil && req.GetHidden().GetValue() {
		hidden = 1
	}
	fileUpdateSrv = &service.FileUpdate{
		BaseService: service.BaseService{DB: db},
		Token:       token,
		File:        file,
		IP:          record.IP,
		Hidden:      &hidden,
		Path:        &req.Path,
	}
	if err = fileUpdateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		return
	}
	if fileUpdateVal, err = fileUpdateSrv.Execute(ctx); err != nil {
		return
	}
	if resp.File, err = s.fileResp(fileUpdateVal.(*models.File), db); err != nil {
		return
	}
	return
}

// FileDelete is used yo delete a file or a directory
func (s *Server) FileDelete(ctx context.Context, req *FileDeleteRequest) (resp *FileDeleteResponse, err error) {
	var (
		db            = getDbConn()
		file          *models.File
		token         *models.Token
		record        *models.Request
		fileDeleteSrv *service.FileDelete
		fileDeleteVal interface{}
	)
	defer func() {
		if err != nil {
			err = status.Error(codes.InvalidArgument, err.Error())
		}
	}()
	if record, err = s.generateRequestRecord(ctx, "FileDelete", req, db); err != nil {
		return
	}
	resp = &FileDeleteResponse{RequestId: record.ID}
	if token, err = s.fetchToken(req.Token, req.Secret, db); err != nil {
		return
	}
	record.AppID = &token.App.ID
	record.Token = &token.UID
	if file, err = models.FindFileByUID(req.FileUid, false, db); err != nil {
		return
	}

	fileDeleteSrv = &service.FileDelete{
		BaseService: service.BaseService{DB: db},
		Token:       token,
		File:        file,
		IP:          record.IP,
		Force:       &req.ForceDeleteIfDir,
	}

	if err = fileDeleteSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		return
	}

	if fileDeleteVal, err = fileDeleteSrv.Execute(ctx); err != nil {
		return
	}
	if resp.File, err = s.fileResp(fileDeleteVal.(*models.File), db); err != nil {
		return
	}
	return
}

// FileRead is used to download file
func (s *Server) FileRead(req *FileReadRequest, resp FileRead_FileReadServer) (err error) {
	var (
		db          = getDbConn()
		ctx         = resp.Context()
		file        *models.File
		token       *models.Token
		record      *models.Request
		fileReader  io.Reader
		fileReadSrv *service.FileRead
		fileReadVal interface{}
	)
	defer func() {
		if err != nil {
			err = status.Error(codes.InvalidArgument, err.Error())
		}
	}()
	if record, err = s.generateRequestRecord(ctx, "FileRead", req, db); err != nil {
		return
	}
	if token, err = s.fetchToken(req.Token, req.Secret, db); err != nil {
		return
	}
	record.AppID = &token.App.ID
	record.Token = &token.UID
	if file, err = models.FindFileByUID(req.FileUid, false, db); err != nil {
		return
	}
	if err = db.Model(record).Updates(map[string]interface{}{"appId": record.AppID, "token": record.Token}).Error; err != nil {
		return
	}
	fileReadSrv = &service.FileRead{
		BaseService: service.BaseService{DB: db, RootPath: testRootPath},
		Token:       token,
		File:        file,
		IP:          record.IP,
	}
	if err = fileReadSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		return
	}
	if fileReadVal, err = fileReadSrv.Execute(ctx); err != nil {
		return
	}
	fileReader = fileReadVal.(io.Reader)

	if err = resp.SendHeader(metadata.New(map[string]string{
		"name": file.Name,
		"size": strconv.Itoa(file.Size),
		"hash": file.Object.Hash,
	})); err != nil {
		return
	}

	for {
		var chunk = make([]byte, models.ChunkSize)
		var readCount int
		readCount, err = fileReader.Read(chunk)
		if readCount > 0 {
			if err = resp.Send(&FileReadResponse{Content: chunk[:readCount]}); err != nil {
				return
			}
		}
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
	}
}

// DirectoryList is used to list a directory
func (s *Server) DirectoryList(ctx context.Context, req *DirectoryListRequest) (resp *DirectoryListResponse, err error) {
	var (
		db          = getDbConn()
		token       *models.Token
		record      *models.Request
		dirListSrv  *service.DirectoryList
		dirListVal  interface{}
		dirListResp *service.DirectoryListResponse
	)
	defer func() {
		if err != nil {
			err = status.Error(codes.InvalidArgument, err.Error())
		}
	}()
	if record, err = s.generateRequestRecord(ctx, "DirectoryList", req, db); err != nil {
		return
	}
	resp = &DirectoryListResponse{RequestId: record.ID}
	if token, err = s.fetchToken(req.Token, req.Secret, db); err != nil {
		return
	}
	record.AppID = &token.App.ID
	record.Token = &token.UID

	dirListSrv = &service.DirectoryList{
		BaseService: service.BaseService{DB: db},
		Token:       token,
		IP:          record.IP,
		SubDir:      "/",
		Limit:       10,
		Offset:      0,
		Sort:        "-type",
	}

	if req.GetSubDir() != nil {
		dirListSrv.SubDir = req.GetSubDir().GetValue()
	}
	if req.GetLimit() != nil {
		dirListSrv.Limit = int(req.GetLimit().GetValue())
	}
	if req.GetOffset() != nil {
		dirListSrv.Offset = int(req.GetOffset().GetValue())
	}

	sort := strings.Replace(req.Sort.String(), "Desc", "-", -1)
	sort = strings.Replace(sort, "Asc", "", -1)
	sort = strings.ToLower(sort)
	dirListSrv.Sort = sort

	if err = dirListSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		return
	}
	if dirListVal, err = dirListSrv.Execute(ctx); err != nil {
		return
	}
	dirListResp = dirListVal.(*service.DirectoryListResponse)
	resp.Total = int32(dirListResp.Total)
	resp.Pages = int32(dirListResp.Pages)
	resp.Files = make([]*File, len(dirListResp.Files))
	for i, f := range dirListResp.Files {
		if resp.Files[i], err = s.fileResp(&f, db); err != nil {
			return
		}
	}
	return
}
