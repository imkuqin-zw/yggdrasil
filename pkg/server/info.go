package server

import (
	"fmt"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
)

type info struct {
	scheme   string
	kind     types.ServerKind
	host     string
	metadata map[string]string
}

func (si *info) Endpoint() string {
	return fmt.Sprintf("%s://%s", si.scheme, si.host)
}

func (si *info) Host() string {
	return si.host
}

func (si *info) Metadata() map[string]string {
	return si.metadata
}

func (si *info) Kind() types.ServerKind {
	return si.kind
}

func (si *info) Scheme() string {
	return si.scheme
}

func NewInfo(scheme string, kind types.ServerKind, host string, metadata map[string]string) types.ServerInfo {
	return &info{scheme: scheme, kind: kind, host: host, metadata: metadata}
}
