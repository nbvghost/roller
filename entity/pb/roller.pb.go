// Code generated by protoc-gen-go. DO NOT EDIT.
// source: roller.proto

package pb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

//操作信息
type ActionState struct {
	Action               string   `protobuf:"bytes,2,opt,name=Action,proto3" json:"Action,omitempty"`
	Message              string   `protobuf:"bytes,3,opt,name=Message,proto3" json:"Message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ActionState) Reset()         { *m = ActionState{} }
func (m *ActionState) String() string { return proto.CompactTextString(m) }
func (*ActionState) ProtoMessage()    {}
func (*ActionState) Descriptor() ([]byte, []int) {
	return fileDescriptor_ae2ea131337aa6ca, []int{0}
}

func (m *ActionState) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ActionState.Unmarshal(m, b)
}
func (m *ActionState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ActionState.Marshal(b, m, deterministic)
}
func (m *ActionState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ActionState.Merge(m, src)
}
func (m *ActionState) XXX_Size() int {
	return xxx_messageInfo_ActionState.Size(m)
}
func (m *ActionState) XXX_DiscardUnknown() {
	xxx_messageInfo_ActionState.DiscardUnknown(m)
}

var xxx_messageInfo_ActionState proto.InternalMessageInfo

func (m *ActionState) GetAction() string {
	if m != nil {
		return m.Action
	}
	return ""
}

func (m *ActionState) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*ActionState)(nil), "pb.ActionState")
}

func init() { proto.RegisterFile("roller.proto", fileDescriptor_ae2ea131337aa6ca) }

var fileDescriptor_ae2ea131337aa6ca = []byte{
	// 89 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x48, 0xcb, 0xd2, 0x2b,
	0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x52, 0xb2, 0xe7, 0xe2, 0x76, 0x4c, 0x2e, 0xc9,
	0xcc, 0xcf, 0x0b, 0x2e, 0x49, 0x2c, 0x49, 0x15, 0x12, 0xe3, 0x62, 0x83, 0x70, 0x25, 0x98, 0x14,
	0x18, 0x35, 0x38, 0x83, 0xa0, 0x3c, 0x21, 0x09, 0x2e, 0x76, 0xdf, 0xd4, 0xe2, 0xe2, 0xc4, 0xf4,
	0x54, 0x09, 0x66, 0xb0, 0x04, 0x8c, 0x9b, 0xc4, 0x06, 0x36, 0xcb, 0x18, 0x10, 0x00, 0x00, 0xff,
	0xff, 0xa5, 0x0e, 0x18, 0x17, 0x57, 0x00, 0x00, 0x00,
}
