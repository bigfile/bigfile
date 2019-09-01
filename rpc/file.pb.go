// Code generated by protoc-gen-go. DO NOT EDIT.
// source: file.proto

package rpc

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
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

// File represent a file type
type File struct {
	Uid                  string                `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Path                 string                `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Size                 uint64                `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	IsDir                bool                  `protobuf:"varint,4,opt,name=is_dir,json=isDir,proto3" json:"is_dir,omitempty"`
	Hidden               bool                  `protobuf:"varint,5,opt,name=hidden,proto3" json:"hidden,omitempty"`
	Hash                 *wrappers.StringValue `protobuf:"bytes,6,opt,name=hash,proto3" json:"hash,omitempty"`
	Ext                  *wrappers.StringValue `protobuf:"bytes,7,opt,name=ext,proto3" json:"ext,omitempty"`
	DeletedAt            *timestamp.Timestamp  `protobuf:"bytes,8,opt,name=deleted_at,json=deletedAt,proto3" json:"deleted_at,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *File) Reset()         { *m = File{} }
func (m *File) String() string { return proto.CompactTextString(m) }
func (*File) ProtoMessage()    {}
func (*File) Descriptor() ([]byte, []int) {
	return fileDescriptor_9188e3b7e55e1162, []int{0}
}

func (m *File) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_File.Unmarshal(m, b)
}
func (m *File) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_File.Marshal(b, m, deterministic)
}
func (m *File) XXX_Merge(src proto.Message) {
	xxx_messageInfo_File.Merge(m, src)
}
func (m *File) XXX_Size() int {
	return xxx_messageInfo_File.Size(m)
}
func (m *File) XXX_DiscardUnknown() {
	xxx_messageInfo_File.DiscardUnknown(m)
}

var xxx_messageInfo_File proto.InternalMessageInfo

func (m *File) GetUid() string {
	if m != nil {
		return m.Uid
	}
	return ""
}

func (m *File) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *File) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *File) GetIsDir() bool {
	if m != nil {
		return m.IsDir
	}
	return false
}

func (m *File) GetHidden() bool {
	if m != nil {
		return m.Hidden
	}
	return false
}

func (m *File) GetHash() *wrappers.StringValue {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *File) GetExt() *wrappers.StringValue {
	if m != nil {
		return m.Ext
	}
	return nil
}

func (m *File) GetDeletedAt() *timestamp.Timestamp {
	if m != nil {
		return m.DeletedAt
	}
	return nil
}

func init() {
	proto.RegisterType((*File)(nil), "bigfile.file.File")
}

func init() { proto.RegisterFile("file.proto", fileDescriptor_9188e3b7e55e1162) }

var fileDescriptor_9188e3b7e55e1162 = []byte{
	// 325 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0xb1, 0x4e, 0xeb, 0x30,
	0x14, 0x86, 0xe5, 0x24, 0xcd, 0x6d, 0x7d, 0xef, 0x70, 0x65, 0xdd, 0x8b, 0xac, 0x0a, 0xb5, 0x11,
	0x53, 0x26, 0x17, 0xc1, 0xc4, 0x48, 0x84, 0x58, 0x58, 0xa2, 0x50, 0x81, 0xc4, 0x52, 0x25, 0xcd,
	0x69, 0x62, 0xc9, 0x89, 0x2d, 0xdb, 0x51, 0x81, 0x67, 0xe0, 0x29, 0x18, 0x79, 0x42, 0x46, 0x14,
	0x37, 0x61, 0x80, 0x85, 0x25, 0x39, 0xe7, 0x3f, 0xdf, 0xef, 0x73, 0xf4, 0x63, 0xbc, 0xe3, 0x02,
	0x98, 0xd2, 0xd2, 0x4a, 0xf2, 0xa7, 0xe0, 0x95, 0x6b, 0xfb, 0xcf, 0x7c, 0x51, 0x49, 0x59, 0x09,
	0x58, 0xb9, 0x59, 0xd1, 0xed, 0x56, 0x7b, 0x9d, 0x2b, 0x05, 0xda, 0x1c, 0xe8, 0xf9, 0xf2, 0xeb,
	0xdc, 0xf2, 0x06, 0x8c, 0xcd, 0x1b, 0x75, 0x00, 0x4e, 0x5e, 0x3c, 0x1c, 0x5c, 0x73, 0x01, 0xe4,
	0x2f, 0xf6, 0x3b, 0x5e, 0x52, 0x14, 0xa1, 0x78, 0x96, 0xf5, 0x25, 0x21, 0x38, 0x50, 0xb9, 0xad,
	0xa9, 0xe7, 0x24, 0x57, 0xf7, 0x9a, 0xe1, 0xcf, 0x40, 0xfd, 0x08, 0xc5, 0x41, 0xe6, 0x6a, 0xf2,
	0x1f, 0x87, 0xdc, 0x6c, 0x4a, 0xae, 0x69, 0x10, 0xa1, 0x78, 0x9a, 0x4d, 0xb8, 0xb9, 0xe2, 0x9a,
	0x1c, 0xe1, 0xb0, 0xe6, 0x65, 0x09, 0x2d, 0x9d, 0x38, 0x79, 0xe8, 0xc8, 0x29, 0x0e, 0xea, 0xdc,
	0xd4, 0x34, 0x8c, 0x50, 0xfc, 0xfb, 0xec, 0x98, 0x1d, 0x2e, 0x64, 0xe3, 0x85, 0xec, 0xd6, 0x6a,
	0xde, 0x56, 0x77, 0xb9, 0xe8, 0x20, 0x73, 0x24, 0x61, 0xd8, 0x87, 0x47, 0x4b, 0x7f, 0xfd, 0xc0,
	0xd0, 0x83, 0xe4, 0x02, 0xe3, 0x12, 0x04, 0x58, 0x28, 0x37, 0xb9, 0xa5, 0x53, 0x67, 0x9b, 0x7f,
	0xb3, 0xad, 0xc7, 0x24, 0xb2, 0xd9, 0x40, 0x5f, 0xda, 0xa4, 0xc5, 0xff, 0xb6, 0xb2, 0x61, 0x63,
	0xc6, 0x23, 0x9c, 0xcc, 0xfa, 0x8c, 0xd2, 0xbe, 0x4b, 0xd1, 0xc3, 0xa2, 0xe2, 0xb6, 0xee, 0x0a,
	0xb6, 0x95, 0xcd, 0x6a, 0x20, 0x3f, 0xff, 0x5a, 0x6d, 0xdf, 0x11, 0x7a, 0xf5, 0xfc, 0x24, 0xcd,
	0xde, 0xbc, 0x65, 0x32, 0x3c, 0x94, 0x8e, 0x5b, 0xef, 0x41, 0x88, 0x9b, 0x56, 0xee, 0xdb, 0xf5,
	0x93, 0x02, 0x53, 0x84, 0x6e, 0xc3, 0xf9, 0x47, 0x00, 0x00, 0x00, 0xff, 0xff, 0x4d, 0x65, 0xc7,
	0xfa, 0xe2, 0x01, 0x00, 0x00,
}
