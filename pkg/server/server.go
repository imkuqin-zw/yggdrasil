package server

import "github.com/imkuqin-zw/yggdrasil/pkg/types"

var serverConstructors []types.ServerConstructor

func RegisterConstructor(constructor types.ServerConstructor) {
	serverConstructors = append(serverConstructors, constructor)
}

func GetConstructors() []types.ServerConstructor {
	return serverConstructors
}
