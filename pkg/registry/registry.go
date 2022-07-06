package registry

import "github.com/imkuqin-zw/yggdrasil/pkg/types"

var registryConstructors = make(map[string]types.RegistryConstructor)

func RegisterConstructor(name string, constructor types.RegistryConstructor) {
	registryConstructors[name] = constructor
}

func GetConstructor(name string) types.RegistryConstructor {
	constructor, _ := registryConstructors[name]
	return constructor
}
