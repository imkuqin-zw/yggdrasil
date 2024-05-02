package prometheus

import (
	"log"
	"net/http"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/governor"
	"github.com/imkuqin-zw/yggdrasil/pkg/otel"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

func init() {
	otel.RegisterMeterProviderBuilder("prometheus", builder)
}

func builder(string) api.MeterProvider {
	governor.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})
	exporter, err := prometheus.New(prometheus.WithNamespace(pkg.Namespace()))
	if err != nil {
		log.Fatal(err)
	}
	return metric.NewMeterProvider(metric.WithReader(exporter))
}
