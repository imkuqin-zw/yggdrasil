// Code generated by protoc-gen-yggdrasil-reason. DO NOT EDIT.

package helloword

import (
	code "google.golang.org/genproto/googleapis/rpc/code"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the yggdrasil package it is being compiled against.

var Reason_code = map[int32]code.Code{
	0: code.Code_OK,
	1: code.Code_NOT_FOUND,
}

func (r Reason) Reason() string {
	return Reason_name[int32(r)]
}

func (r Reason) Domain() string {
	return "yggdrasil.example.proto.helloword"
}

func (r Reason) Code() code.Code {
	return Reason_code[int32(r)]
}
