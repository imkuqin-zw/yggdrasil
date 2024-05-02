package otel

import "github.com/imkuqin-zw/yggdrasil/pkg/stats"

func init() {
	stats.RegisterHandlerBuilder("otel", func(isServer bool) stats.Handler {
		if isServer {
			return newSvrHandler()
		}
		return newCliHandler()
	})
}
