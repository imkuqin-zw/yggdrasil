package types

import "context"

type RegistryConstructor func() Registry

type Registry interface {
	Register(context.Context, RegistryInstance) error
	Deregister(context.Context, RegistryInstance) error
	Name() string
}

type RegistryInstance interface {
	Region() string
	Zone() string
	Campus() string
	Namespace() string
	Name() string
	Version() string
	Metadata() map[string]string
	Endpoints() []ServerInfo
}
