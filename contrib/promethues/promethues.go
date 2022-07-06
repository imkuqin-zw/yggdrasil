package promethues

import (
	"net/http"

	"github.com/imkuqin-zw/yggdrasil/pkg/governor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	governor.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})
}
