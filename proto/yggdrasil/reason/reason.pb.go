// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.22.2
// source: reason.proto

package reason

import (
	code "google.golang.org/genproto/googleapis/rpc/code"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var file_reason_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.EnumOptions)(nil),
		ExtensionType: (*int32)(nil),
		Field:         1108,
		Name:          "yggdrasil.reason.default_reason",
		Tag:           "varint,1108,opt,name=default_reason",
		Filename:      "reason.proto",
	},
	{
		ExtendedType:  (*descriptorpb.EnumValueOptions)(nil),
		ExtensionType: (*code.Code)(nil),
		Field:         1109,
		Name:          "yggdrasil.reason.code",
		Tag:           "varint,1109,opt,name=code,enum=google.rpc.Code",
		Filename:      "reason.proto",
	},
}

// Extension fields to descriptorpb.EnumOptions.
var (
	// optional int32 default_reason = 1108;
	E_DefaultReason = &file_reason_proto_extTypes[0]
)

// Extension fields to descriptorpb.EnumValueOptions.
var (
	// optional google.rpc.Code code = 1109;
	E_Code = &file_reason_proto_extTypes[1]
)

var File_reason_proto protoreflect.FileDescriptor

var file_reason_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10,
	0x79, 0x67, 0x67, 0x64, 0x72, 0x61, 0x73, 0x69, 0x6c, 0x2e, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e,
	0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72,
	0x70, 0x63, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3a, 0x44, 0x0a,
	0x0e, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x5f, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12,
	0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x45, 0x6e, 0x75, 0x6d, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd4, 0x08,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x3a, 0x48, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x21, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6e,
	0x75, 0x6d, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd5,
	0x08, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x10, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x72,
	0x70, 0x63, 0x2e, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x42, 0x7c, 0x0a,
	0x26, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x69, 0x6d, 0x6b, 0x75,
	0x71, 0x69, 0x6e, 0x5f, 0x7a, 0x77, 0x2e, 0x79, 0x67, 0x67, 0x64, 0x72, 0x61, 0x73, 0x69, 0x6c,
	0x2e, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x42, 0x0b, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x69, 0x6d, 0x6b, 0x75, 0x71, 0x69, 0x6e, 0x2d, 0x7a, 0x77, 0x2f, 0x79, 0x67,
	0x67, 0x64, 0x72, 0x61, 0x73, 0x69, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x79, 0x67,
	0x67, 0x64, 0x72, 0x61, 0x73, 0x69, 0x6c, 0x2f, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x3b, 0x72,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0xa2, 0x02, 0x03, 0x41, 0x50, 0x49, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var file_reason_proto_goTypes = []interface{}{
	(*descriptorpb.EnumOptions)(nil),      // 0: google.protobuf.EnumOptions
	(*descriptorpb.EnumValueOptions)(nil), // 1: google.protobuf.EnumValueOptions
	(code.Code)(0),                        // 2: google.rpc.Code
}
var file_reason_proto_depIdxs = []int32{
	0, // 0: yggdrasil.reason.default_reason:extendee -> google.protobuf.EnumOptions
	1, // 1: yggdrasil.reason.code:extendee -> google.protobuf.EnumValueOptions
	2, // 2: yggdrasil.reason.code:type_name -> google.rpc.Code
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	2, // [2:3] is the sub-list for extension type_name
	0, // [0:2] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_reason_proto_init() }
func file_reason_proto_init() {
	if File_reason_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_reason_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 2,
			NumServices:   0,
		},
		GoTypes:           file_reason_proto_goTypes,
		DependencyIndexes: file_reason_proto_depIdxs,
		ExtensionInfos:    file_reason_proto_extTypes,
	}.Build()
	File_reason_proto = out.File
	file_reason_proto_rawDesc = nil
	file_reason_proto_goTypes = nil
	file_reason_proto_depIdxs = nil
}
