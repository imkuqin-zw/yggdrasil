// Code generated by protoc-gen-go. DO NOT EDIT.
// source: plugininfo.proto

package v1

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

type PluginAPIResultType int32

const (
	PluginAPIResultType_UnknownResult PluginAPIResultType = 0
	PluginAPIResultType_APISuccess    PluginAPIResultType = 1
	PluginAPIResultType_APIFail       PluginAPIResultType = 2
)

var PluginAPIResultType_name = map[int32]string{
	0: "UnknownResult",
	1: "APISuccess",
	2: "APIFail",
}

var PluginAPIResultType_value = map[string]int32{
	"UnknownResult": 0,
	"APISuccess":    1,
	"APIFail":       2,
}

func (x PluginAPIResultType) String() string {
	return proto.EnumName(PluginAPIResultType_name, int32(x))
}

func (PluginAPIResultType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_2a4baee40ed718cc, []int{0}
}

type PluginAPIStatistics struct {
	Id                   string             `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Token                *SDKToken          `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
	PluginApi            *PluginAPIKey      `protobuf:"bytes,3,opt,name=plugin_api,json=pluginApi,proto3" json:"plugin_api,omitempty"`
	Results              []*PluginAPIResult `protobuf:"bytes,4,rep,name=results,proto3" json:"results,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *PluginAPIStatistics) Reset()         { *m = PluginAPIStatistics{} }
func (m *PluginAPIStatistics) String() string { return proto.CompactTextString(m) }
func (*PluginAPIStatistics) ProtoMessage()    {}
func (*PluginAPIStatistics) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a4baee40ed718cc, []int{0}
}

func (m *PluginAPIStatistics) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PluginAPIStatistics.Unmarshal(m, b)
}
func (m *PluginAPIStatistics) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PluginAPIStatistics.Marshal(b, m, deterministic)
}
func (m *PluginAPIStatistics) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PluginAPIStatistics.Merge(m, src)
}
func (m *PluginAPIStatistics) XXX_Size() int {
	return xxx_messageInfo_PluginAPIStatistics.Size(m)
}
func (m *PluginAPIStatistics) XXX_DiscardUnknown() {
	xxx_messageInfo_PluginAPIStatistics.DiscardUnknown(m)
}

var xxx_messageInfo_PluginAPIStatistics proto.InternalMessageInfo

func (m *PluginAPIStatistics) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *PluginAPIStatistics) GetToken() *SDKToken {
	if m != nil {
		return m.Token
	}
	return nil
}

func (m *PluginAPIStatistics) GetPluginApi() *PluginAPIKey {
	if m != nil {
		return m.PluginApi
	}
	return nil
}

func (m *PluginAPIStatistics) GetResults() []*PluginAPIResult {
	if m != nil {
		return m.Results
	}
	return nil
}

type PluginAPIKey struct {
	PluginType           string   `protobuf:"bytes,1,opt,name=plugin_type,json=pluginType,proto3" json:"plugin_type,omitempty"`
	PluginName           string   `protobuf:"bytes,2,opt,name=plugin_name,json=pluginName,proto3" json:"plugin_name,omitempty"`
	PluginMethod         string   `protobuf:"bytes,3,opt,name=plugin_method,json=pluginMethod,proto3" json:"plugin_method,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PluginAPIKey) Reset()         { *m = PluginAPIKey{} }
func (m *PluginAPIKey) String() string { return proto.CompactTextString(m) }
func (*PluginAPIKey) ProtoMessage()    {}
func (*PluginAPIKey) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a4baee40ed718cc, []int{1}
}

func (m *PluginAPIKey) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PluginAPIKey.Unmarshal(m, b)
}
func (m *PluginAPIKey) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PluginAPIKey.Marshal(b, m, deterministic)
}
func (m *PluginAPIKey) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PluginAPIKey.Merge(m, src)
}
func (m *PluginAPIKey) XXX_Size() int {
	return xxx_messageInfo_PluginAPIKey.Size(m)
}
func (m *PluginAPIKey) XXX_DiscardUnknown() {
	xxx_messageInfo_PluginAPIKey.DiscardUnknown(m)
}

var xxx_messageInfo_PluginAPIKey proto.InternalMessageInfo

func (m *PluginAPIKey) GetPluginType() string {
	if m != nil {
		return m.PluginType
	}
	return ""
}

func (m *PluginAPIKey) GetPluginName() string {
	if m != nil {
		return m.PluginName
	}
	return ""
}

func (m *PluginAPIKey) GetPluginMethod() string {
	if m != nil {
		return m.PluginMethod
	}
	return ""
}

type PluginAPIResult struct {
	RetCode                string              `protobuf:"bytes,1,opt,name=ret_code,json=retCode,proto3" json:"ret_code,omitempty"`
	TotalRequestsPerMinute uint32              `protobuf:"varint,2,opt,name=total_requests_per_minute,json=totalRequestsPerMinute,proto3" json:"total_requests_per_minute,omitempty"`
	Type                   PluginAPIResultType `protobuf:"varint,3,opt,name=type,proto3,enum=v1.PluginAPIResultType" json:"type,omitempty"`
	DelayRange             string              `protobuf:"bytes,4,opt,name=delay_range,json=delayRange,proto3" json:"delay_range,omitempty"`
	XXX_NoUnkeyedLiteral   struct{}            `json:"-"`
	XXX_unrecognized       []byte              `json:"-"`
	XXX_sizecache          int32               `json:"-"`
}

func (m *PluginAPIResult) Reset()         { *m = PluginAPIResult{} }
func (m *PluginAPIResult) String() string { return proto.CompactTextString(m) }
func (*PluginAPIResult) ProtoMessage()    {}
func (*PluginAPIResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a4baee40ed718cc, []int{2}
}

func (m *PluginAPIResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PluginAPIResult.Unmarshal(m, b)
}
func (m *PluginAPIResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PluginAPIResult.Marshal(b, m, deterministic)
}
func (m *PluginAPIResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PluginAPIResult.Merge(m, src)
}
func (m *PluginAPIResult) XXX_Size() int {
	return xxx_messageInfo_PluginAPIResult.Size(m)
}
func (m *PluginAPIResult) XXX_DiscardUnknown() {
	xxx_messageInfo_PluginAPIResult.DiscardUnknown(m)
}

var xxx_messageInfo_PluginAPIResult proto.InternalMessageInfo

func (m *PluginAPIResult) GetRetCode() string {
	if m != nil {
		return m.RetCode
	}
	return ""
}

func (m *PluginAPIResult) GetTotalRequestsPerMinute() uint32 {
	if m != nil {
		return m.TotalRequestsPerMinute
	}
	return 0
}

func (m *PluginAPIResult) GetType() PluginAPIResultType {
	if m != nil {
		return m.Type
	}
	return PluginAPIResultType_UnknownResult
}

func (m *PluginAPIResult) GetDelayRange() string {
	if m != nil {
		return m.DelayRange
	}
	return ""
}

func init() {
	proto.RegisterEnum("v1.PluginAPIResultType", PluginAPIResultType_name, PluginAPIResultType_value)
	proto.RegisterType((*PluginAPIStatistics)(nil), "v1.PluginAPIStatistics")
	proto.RegisterType((*PluginAPIKey)(nil), "v1.PluginAPIKey")
	proto.RegisterType((*PluginAPIResult)(nil), "v1.PluginAPIResult")
}

func init() { proto.RegisterFile("plugininfo.proto", fileDescriptor_2a4baee40ed718cc) }

var fileDescriptor_2a4baee40ed718cc = []byte{
	// 379 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0xcf, 0x8e, 0xd3, 0x30,
	0x18, 0xc4, 0x49, 0xb6, 0x50, 0xfa, 0xa5, 0x0d, 0xc1, 0x2b, 0x41, 0x96, 0x0b, 0x55, 0xb8, 0x54,
	0x20, 0x8a, 0xb6, 0x9c, 0x38, 0x56, 0xfc, 0x91, 0xaa, 0xaa, 0x28, 0x72, 0xcb, 0x39, 0x32, 0xc9,
	0x47, 0xb1, 0x9a, 0xd8, 0xc1, 0x76, 0x8a, 0xf2, 0x48, 0x3c, 0x00, 0xef, 0x87, 0x62, 0xa7, 0xa5,
	0xa0, 0xbd, 0xce, 0xfc, 0x9c, 0x99, 0x89, 0x0d, 0x51, 0x5d, 0x36, 0x7b, 0x2e, 0xb8, 0xf8, 0x26,
	0xe7, 0xb5, 0x92, 0x46, 0x12, 0xff, 0x78, 0xfb, 0x2c, 0xd4, 0xc5, 0xc1, 0xc8, 0x03, 0x0a, 0xa7,
	0x25, 0xbf, 0x3c, 0xb8, 0x4e, 0x2d, 0xb8, 0x4c, 0x57, 0x5b, 0xc3, 0x0c, 0xd7, 0x86, 0xe7, 0x9a,
	0x84, 0xe0, 0xf3, 0x22, 0xf6, 0xa6, 0xde, 0x6c, 0x44, 0x7d, 0x5e, 0x90, 0x04, 0xee, 0xdb, 0x63,
	0xb1, 0x3f, 0xf5, 0x66, 0xc1, 0x62, 0x3c, 0x3f, 0xde, 0xce, 0xb7, 0x1f, 0xd6, 0xbb, 0x4e, 0xa3,
	0xce, 0x22, 0x6f, 0x00, 0x5c, 0x66, 0xc6, 0x6a, 0x1e, 0x5f, 0x59, 0x30, 0xea, 0xc0, 0x73, 0xc0,
	0x1a, 0x5b, 0x3a, 0x72, 0xcc, 0xb2, 0xe6, 0xe4, 0x35, 0x0c, 0x15, 0xea, 0xa6, 0x34, 0x3a, 0x1e,
	0x4c, 0xaf, 0x66, 0xc1, 0xe2, 0xfa, 0x1f, 0x9a, 0x5a, 0x8f, 0x9e, 0x98, 0xa4, 0x81, 0xf1, 0xe5,
	0x97, 0xc8, 0x73, 0x08, 0xfa, 0x3c, 0xd3, 0xd6, 0xd8, 0x97, 0xed, 0x2b, 0xec, 0xda, 0x1a, 0x2f,
	0x00, 0xc1, 0x2a, 0xb4, 0xd5, 0xcf, 0xc0, 0x67, 0x56, 0x21, 0x79, 0x01, 0x93, 0x1e, 0xa8, 0xd0,
	0x7c, 0x97, 0x85, 0x2d, 0x3d, 0xa2, 0x63, 0x27, 0x6e, 0xac, 0x96, 0xfc, 0xf6, 0xe0, 0xd1, 0x7f,
	0x9d, 0xc8, 0x0d, 0x3c, 0x54, 0x68, 0xb2, 0x5c, 0x16, 0xa7, 0xdc, 0xa1, 0x42, 0xf3, 0x5e, 0x16,
	0x48, 0xde, 0xc1, 0x8d, 0x91, 0x86, 0x95, 0x99, 0xc2, 0x1f, 0x0d, 0x6a, 0xa3, 0xb3, 0x1a, 0x55,
	0x56, 0x71, 0xd1, 0x18, 0x57, 0x61, 0x42, 0x9f, 0x58, 0x80, 0xf6, 0x7e, 0x8a, 0x6a, 0x63, 0x5d,
	0xf2, 0x0a, 0x06, 0x76, 0x49, 0xd7, 0x22, 0x5c, 0x3c, 0xbd, 0xe3, 0x67, 0x74, 0xb3, 0xa8, 0x85,
	0xba, 0x71, 0x05, 0x96, 0xac, 0xcd, 0x14, 0x13, 0x7b, 0x8c, 0x07, 0x6e, 0x9c, 0x95, 0x68, 0xa7,
	0xbc, 0xfc, 0x78, 0x71, 0xb3, 0x7f, 0x4f, 0x93, 0xc7, 0x30, 0xf9, 0x22, 0x0e, 0x42, 0xfe, 0x14,
	0x4e, 0x8c, 0xee, 0x91, 0x10, 0xa0, 0xbb, 0xfd, 0x26, 0xcf, 0x51, 0xeb, 0xc8, 0x23, 0x01, 0x0c,
	0x97, 0xe9, 0xea, 0x13, 0xe3, 0x65, 0xe4, 0x7f, 0x7d, 0x60, 0x1f, 0xca, 0xdb, 0x3f, 0x01, 0x00,
	0x00, 0xff, 0xff, 0x13, 0x72, 0xab, 0x23, 0x50, 0x02, 0x00, 0x00,
}
