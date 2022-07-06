package trace

import "github.com/imkuqin-zw/yggdrasil/pkg/types"

var constructors = make(map[string]types.TracerProviderConstructor)

func RegisterConstructor(name string, constructor types.TracerProviderConstructor) {
	constructors[name] = constructor
}

func GetConstructor(name string) types.TracerProviderConstructor {
	constructor, _ := constructors[name]
	return constructor
}
