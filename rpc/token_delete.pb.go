// Code generated by protoc-gen-go. DO NOT EDIT.
// source: token_delete.proto

package rpc

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// TokenDeleteRequest describe the input format for deleting token
type TokenDeleteRequest struct {
	AppUid               string   `protobuf:"bytes,1,opt,name=app_uid,json=appUid,proto3" json:"app_uid,omitempty"`
	AppSecret            string   `protobuf:"bytes,2,opt,name=app_secret,json=appSecret,proto3" json:"app_secret,omitempty"`
	Token                string   `protobuf:"bytes,3,opt,name=token,proto3" json:"token,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TokenDeleteRequest) Reset()         { *m = TokenDeleteRequest{} }
func (m *TokenDeleteRequest) String() string { return proto.CompactTextString(m) }
func (*TokenDeleteRequest) ProtoMessage()    {}
func (*TokenDeleteRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_b9cf4ff266416cfa, []int{0}
}

func (m *TokenDeleteRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TokenDeleteRequest.Unmarshal(m, b)
}
func (m *TokenDeleteRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TokenDeleteRequest.Marshal(b, m, deterministic)
}
func (m *TokenDeleteRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TokenDeleteRequest.Merge(m, src)
}
func (m *TokenDeleteRequest) XXX_Size() int {
	return xxx_messageInfo_TokenDeleteRequest.Size(m)
}
func (m *TokenDeleteRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TokenDeleteRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TokenDeleteRequest proto.InternalMessageInfo

func (m *TokenDeleteRequest) GetAppUid() string {
	if m != nil {
		return m.AppUid
	}
	return ""
}

func (m *TokenDeleteRequest) GetAppSecret() string {
	if m != nil {
		return m.AppSecret
	}
	return ""
}

func (m *TokenDeleteRequest) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

// TokenDeleteResponse represent the response of deleting token
type TokenDeleteResponse struct {
	RequestId            uint64   `protobuf:"varint,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	Token                *Token   `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TokenDeleteResponse) Reset()         { *m = TokenDeleteResponse{} }
func (m *TokenDeleteResponse) String() string { return proto.CompactTextString(m) }
func (*TokenDeleteResponse) ProtoMessage()    {}
func (*TokenDeleteResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_b9cf4ff266416cfa, []int{1}
}

func (m *TokenDeleteResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TokenDeleteResponse.Unmarshal(m, b)
}
func (m *TokenDeleteResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TokenDeleteResponse.Marshal(b, m, deterministic)
}
func (m *TokenDeleteResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TokenDeleteResponse.Merge(m, src)
}
func (m *TokenDeleteResponse) XXX_Size() int {
	return xxx_messageInfo_TokenDeleteResponse.Size(m)
}
func (m *TokenDeleteResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_TokenDeleteResponse.DiscardUnknown(m)
}

var xxx_messageInfo_TokenDeleteResponse proto.InternalMessageInfo

func (m *TokenDeleteResponse) GetRequestId() uint64 {
	if m != nil {
		return m.RequestId
	}
	return 0
}

func (m *TokenDeleteResponse) GetToken() *Token {
	if m != nil {
		return m.Token
	}
	return nil
}

func init() {
	proto.RegisterType((*TokenDeleteRequest)(nil), "bigfile.token_delete.TokenDeleteRequest")
	proto.RegisterType((*TokenDeleteResponse)(nil), "bigfile.token_delete.TokenDeleteResponse")
}

func init() { proto.RegisterFile("token_delete.proto", fileDescriptor_b9cf4ff266416cfa) }

var fileDescriptor_b9cf4ff266416cfa = []byte{
	// 287 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0xcf, 0x4a, 0xc3, 0x40,
	0x10, 0xc6, 0x4d, 0xaa, 0x95, 0x4c, 0x2e, 0xb2, 0x06, 0x0c, 0x85, 0x8a, 0xe4, 0x54, 0x3d, 0xac,
	0x50, 0xdf, 0x20, 0x88, 0xe0, 0x2d, 0xc4, 0x7a, 0xf1, 0x12, 0xf3, 0x67, 0x1a, 0x17, 0xd3, 0xee,
	0x98, 0xdd, 0x1c, 0x7c, 0x1d, 0x8f, 0x3e, 0xa1, 0x47, 0xc9, 0x66, 0x2d, 0x29, 0x7a, 0xf0, 0xb4,
	0x7c, 0xf3, 0xcd, 0x7c, 0xbf, 0x9d, 0x01, 0xa6, 0xe5, 0x2b, 0x6e, 0xb3, 0x0a, 0x1b, 0xd4, 0xc8,
	0xa9, 0x95, 0x5a, 0xb2, 0xa0, 0x10, 0xf5, 0x5a, 0x34, 0xc8, 0xc7, 0xde, 0xcc, 0x37, 0x6a, 0x68,
	0x89, 0x0a, 0x60, 0xab, 0x5e, 0xde, 0x1a, 0x2f, 0xc5, 0xb7, 0x0e, 0x95, 0x66, 0x67, 0x70, 0x9c,
	0x13, 0x65, 0x9d, 0xa8, 0x42, 0xe7, 0xc2, 0x59, 0x78, 0xe9, 0x34, 0x27, 0x7a, 0x14, 0x15, 0x9b,
	0x03, 0xf4, 0x86, 0xc2, 0xb2, 0x45, 0x1d, 0xba, 0xc6, 0xf3, 0x72, 0xa2, 0x07, 0x53, 0x60, 0x01,
	0x1c, 0x99, 0xf0, 0x70, 0x62, 0x9c, 0x41, 0x44, 0xcf, 0x70, 0xba, 0xc7, 0x50, 0x24, 0xb7, 0x0a,
	0xfb, 0xac, 0x76, 0xe0, 0x65, 0x96, 0x73, 0x98, 0x7a, 0xb6, 0x72, 0x5f, 0xb1, 0xab, 0x9f, 0xac,
	0x9e, 0xe2, 0x2f, 0x03, 0xbe, 0xb7, 0x0c, 0x37, 0x89, 0x96, 0xb0, 0x54, 0xe0, 0x8f, 0x08, 0xac,
	0x82, 0x61, 0x47, 0x2b, 0x17, 0xfc, 0xaf, 0x3b, 0xf0, 0xdf, 0x7b, 0xcf, 0x2e, 0xff, 0xd1, 0x39,
	0xfc, 0x3e, 0x3a, 0x88, 0x15, 0x04, 0xa5, 0xdc, 0xec, 0x26, 0xcc, 0x3d, 0x8b, 0x6e, 0x1d, 0x9f,
	0x8c, 0xda, 0x93, 0xbe, 0x98, 0x38, 0x4f, 0xe7, 0xb5, 0xd0, 0x2f, 0x5d, 0xc1, 0x4b, 0xb9, 0xb9,
	0xb6, 0x03, 0xbb, 0xb7, 0xa5, 0xf2, 0xcb, 0x71, 0x3e, 0xdc, 0x49, 0x9c, 0xa4, 0x9f, 0xee, 0x3c,
	0xb6, 0x79, 0x89, 0xcd, 0xe3, 0xb1, 0xa8, 0xef, 0x44, 0x83, 0xab, 0x77, 0x42, 0x55, 0x4c, 0x0d,
	0xe6, 0xe6, 0x3b, 0x00, 0x00, 0xff, 0xff, 0xfa, 0x12, 0xfd, 0xc5, 0xef, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// TokenDeleteClient is the client API for TokenDelete service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TokenDeleteClient interface {
	TokenDelete(ctx context.Context, in *TokenDeleteRequest, opts ...grpc.CallOption) (*TokenDeleteResponse, error)
}

type tokenDeleteClient struct {
	cc *grpc.ClientConn
}

func NewTokenDeleteClient(cc *grpc.ClientConn) TokenDeleteClient {
	return &tokenDeleteClient{cc}
}

func (c *tokenDeleteClient) TokenDelete(ctx context.Context, in *TokenDeleteRequest, opts ...grpc.CallOption) (*TokenDeleteResponse, error) {
	out := new(TokenDeleteResponse)
	err := c.cc.Invoke(ctx, "/bigfile.token_delete.TokenDelete/tokenDelete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TokenDeleteServer is the server API for TokenDelete service.
type TokenDeleteServer interface {
	TokenDelete(context.Context, *TokenDeleteRequest) (*TokenDeleteResponse, error)
}

// UnimplementedTokenDeleteServer can be embedded to have forward compatible implementations.
type UnimplementedTokenDeleteServer struct {
}

func (*UnimplementedTokenDeleteServer) TokenDelete(ctx context.Context, req *TokenDeleteRequest) (*TokenDeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TokenDelete not implemented")
}

func RegisterTokenDeleteServer(s *grpc.Server, srv TokenDeleteServer) {
	s.RegisterService(&_TokenDelete_serviceDesc, srv)
}

func _TokenDelete_TokenDelete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TokenDeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TokenDeleteServer).TokenDelete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bigfile.token_delete.TokenDelete/TokenDelete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TokenDeleteServer).TokenDelete(ctx, req.(*TokenDeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TokenDelete_serviceDesc = grpc.ServiceDesc{
	ServiceName: "bigfile.token_delete.TokenDelete",
	HandlerType: (*TokenDeleteServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "tokenDelete",
			Handler:    _TokenDelete_TokenDelete_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "token_delete.proto",
}
