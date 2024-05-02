package plugin

import (
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"gorm.io/gorm"
)

type Factory func(instance string) gorm.Plugin

var plugins = make(map[string]Factory)

func RegisterPluginFactory(name string, f Factory) {
	plugins[name] = f
}

func GetPluginFactory(name string) Factory {
	f, _ := plugins[name]
	return f
}

func GetPlugin(name, instance string) gorm.Plugin {
	f, ok := plugins[name]
	if !ok {
		logger.ErrorField("unknown gorm plugin", logger.String("name", name))
		return nil
	}
	return f(instance)
}
