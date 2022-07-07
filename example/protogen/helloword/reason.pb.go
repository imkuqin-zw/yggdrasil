// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.12.3
// source: helloword/reason.proto

package helloword

import (
	_ "github.com/imkuqin-zw/yggdrasil/proto/yggdrasil/error"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Reason int32

const (
	// Do not use this default value.
	Reason_ERROR_REASON_UNSPECIFIED Reason = 0
)

// Enum value maps for Reason.
var (
	Reason_name = map[int32]string{
		0: "ERROR_REASON_UNSPECIFIED",
	}
	Reason_value = map[string]int32{
		"ERROR_REASON_UNSPECIFIED": 0,
	}
)

func (x Reason) Enum() *Reason {
	p := new(Reason)
	*p = x
	return p
}

func (x Reason) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Reason) Descriptor() protoreflect.EnumDescriptor {
	return file_helloword_reason_proto_enumTypes[0].Descriptor()
}

func (Reason) Type() protoreflect.EnumType {
	return &file_helloword_reason_proto_enumTypes[0]
}

func (x Reason) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Reason.Descriptor instead.
func (Reason) EnumDescriptor() ([]byte, []int) {
	return file_helloword_reason_proto_rawDescGZIP(), []int{0}
}

var File_helloword_reason_proto protoreflect.FileDescriptor

var file_helloword_reason_proto_rawDesc = []byte{
	0x0a, 0x16, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f, 0x72, 0x64, 0x2f, 0x72, 0x65, 0x61, 0x73,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f, 0x72,
	0x64, 0x1a, 0x1c, 0x79, 0x67, 0x67, 0x64, 0x72, 0x61, 0x73, 0x69, 0x6c, 0x2f, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x2f, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2a,
	0x2c, 0x0a, 0x06, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x18, 0x45, 0x52, 0x52,
	0x4f, 0x52, 0x5f, 0x52, 0x45, 0x41, 0x53, 0x4f, 0x4e, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43,
	0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x1a, 0x04, 0xa0, 0x45, 0xf4, 0x03, 0x42, 0x46, 0x5a,
	0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x69, 0x6d, 0x6b, 0x75,
	0x71, 0x69, 0x6e, 0x2d, 0x7a, 0x77, 0x2f, 0x79, 0x67, 0x67, 0x64, 0x72, 0x61, 0x73, 0x69, 0x6c,
	0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x67, 0x65,
	0x6e, 0x2f, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f, 0x72, 0x64, 0x3b, 0x68, 0x65, 0x6c, 0x6c,
	0x6f, 0x77, 0x6f, 0x72, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_helloword_reason_proto_rawDescOnce sync.Once
	file_helloword_reason_proto_rawDescData = file_helloword_reason_proto_rawDesc
)

func file_helloword_reason_proto_rawDescGZIP() []byte {
	file_helloword_reason_proto_rawDescOnce.Do(func() {
		file_helloword_reason_proto_rawDescData = protoimpl.X.CompressGZIP(file_helloword_reason_proto_rawDescData)
	})
	return file_helloword_reason_proto_rawDescData
}

var file_helloword_reason_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_helloword_reason_proto_goTypes = []interface{}{
	(Reason)(0), // 0: example.proto.helloword.Reason
}
var file_helloword_reason_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_helloword_reason_proto_init() }
func file_helloword_reason_proto_init() {
	if File_helloword_reason_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_helloword_reason_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_helloword_reason_proto_goTypes,
		DependencyIndexes: file_helloword_reason_proto_depIdxs,
		EnumInfos:         file_helloword_reason_proto_enumTypes,
	}.Build()
	File_helloword_reason_proto = out.File
	file_helloword_reason_proto_rawDesc = nil
	file_helloword_reason_proto_goTypes = nil
	file_helloword_reason_proto_depIdxs = nil
}