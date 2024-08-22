package middleware

import (
	"net/http"
	"strings"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/rest/marshaler"
)

func init() {
	RegisterBuilder("marshaler", newMarshalerMiddleware)
}

func newMarshalerMiddleware() func(http.Handler) http.Handler {
	marshalerSupport := config.GetString(config.KeyRestMarshalerSupport, "jsonpb")
	names := strings.Split(marshalerSupport, ",")
	marshalerRegistry := marshaler.BuildMarshalerRegistry(names...)
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inbound, outbound := marshalerRegistry.GetMarshaler(r)
			ctx := marshaler.WithInboundContext(r.Context(), inbound)
			ctx = marshaler.WithOutboundContext(ctx, outbound)
			r = r.WithContext(ctx)
			handler.ServeHTTP(w, r)
		})
	}
}
