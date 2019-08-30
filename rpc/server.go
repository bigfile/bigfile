//  Copyright 2019 The bigfile Authors. All rights reserved.
//  Use of this source code is governed by a MIT-style
//  license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"errors"
	"net"
	"reflect"
	"time"

	"github.com/bigfile/bigfile/databases"
	"github.com/bigfile/bigfile/databases/models"
	"github.com/bigfile/bigfile/log"
	"github.com/bigfile/bigfile/service"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var (
	isTesting  = false
	testDbConn *gorm.DB
	// ErrGetIPFailed represent that get ip failed
	ErrGetIPFailed = errors.New("[getClientIP] invoke FromContext() failed")

	// ErrAppSecret represent appUID and appSecret doesn't match
	ErrAppSecret = errors.New("appUID and appSecret doesn't match")

	// ErrTokenNotMatchApp represent
	ErrTokenNotMatchApp = errors.New("")
)

// Server is used to create a rpc server
type Server struct{}

// getClientIP is used to get client ip from the context
func (s *Server) getClientIP(ctx context.Context) (string, error) {
	var (
		pr      *peer.Peer
		ok      bool
		ipV4    string
		tcpAddr *net.TCPAddr
	)
	if pr, ok = peer.FromContext(ctx); !ok {
		return "", ErrGetIPFailed
	}
	if tcpAddr, ok = pr.Addr.(*net.TCPAddr); ok {
		if tcpAddr.IP.IsLoopback() {
			return "127.0.0.1", nil
		}
		ipV4 = tcpAddr.IP.To4().String()
		if len(ipV4) == 0 {
			return tcpAddr.IP.To16().String(), nil
		}
		return ipV4, nil
	}
	return "", ErrGetIPFailed
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

func (s *Server) updateRequestRecord(ctx context.Context, request *models.Request, resp interface{}, err error, db *gorm.DB) {
	var responseBody string

	if err != nil {
		responseBody = err.Error()
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
	if record, err = s.generateRequestRecord(ctx, "TokenCreate", req, db); err != nil {
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
	if record, err = s.generateRequestRecord(ctx, "TokenCreate", req, db); err != nil {
		return resp, err
	}
	resp = &TokenDeleteResponse{RequestId: record.ID}
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
	if err = db.Delete(token).Error; err != nil {
		return
	}
	db.Delete(token)
	db.Unscoped().First(token)
	resp.Token = s.tokenResp(token)
	return
}
