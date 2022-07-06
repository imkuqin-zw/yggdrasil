package types

import (
	"context"

	"github.com/imkuqin-zw/yggdrasil/pkg/md"
)

type ServerStream interface {
	SetHeader(md.MD) error
	SendHeader(md.MD) error
	SetTrailer(md.MD)
	Context() context.Context
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
}

type ClientStream interface {
	Header() (md.MD, error)
	Trailer() md.MD
	CloseSend() error
	Context() context.Context
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
}
