package types

type ServerKind string

const (
	ServerKindRpc      ServerKind = "rpc"
	ServerKindJob      ServerKind = "job"
	ServerKindTask     ServerKind = "task"
	ServerKindGovernor ServerKind = "governor"
)

type ServerInfo interface {
	Scheme() string
	Host() string
	Kind() ServerKind
	Endpoint() string
	Metadata() map[string]string
}

type ServerConstructor func() Server

type Server interface {
	Serve() error
	Stop() error
	Info() ServerInfo
}
