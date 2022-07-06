package log

import (
	"log"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
)

var loggerConstructors = make(map[string]types.LoggerConstructor)

func RegisterConstructor(name string, f types.LoggerConstructor) {
	loggerConstructors[name] = f
}

func GetConstructor(name string) types.LoggerConstructor {
	f, _ := loggerConstructors[name]
	return f
}

func GetLogger(name string) types.Logger {
	f := GetConstructor(name)
	if f == nil {
		log.Fatalf("unknown logger constructor, name: %s", name)
	}
	return f()
}
