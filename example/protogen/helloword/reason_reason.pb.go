// Code generated by protoc-gen-yggdrasil-error. DO NOT EDIT.

package helloword

// This is a compile-time assertion to ensure that this generated file
// is compatible with the yggdrasil package it is being compiled against.

func (r Reason) Reason() string {
	return Reason_name[int32(r)]
}

func (r Reason) Domain() string {
	return "example.proto.helloword"
}