package trace

import (
	"strings"

	"go.opentelemetry.io/otel/propagation"
)

type MetadataReaderWriter struct {
	MD map[string][]string
}

// assert that MetadataReaderWriter implements the TextMapCarrier interface
var _ propagation.TextMapCarrier = (*MetadataReaderWriter)(nil)

func (w MetadataReaderWriter) Get(key string) string {
	values, ok := w.MD[key]
	if !ok {
		return ""
	}
	return strings.Join(values, ";")
}

func (w MetadataReaderWriter) Set(key, val string) {
	// The GRPC HPACK implementation rejects any uppercase keys here.
	//
	// As such, since the HTTP_HEADERS format is case-insensitive anyway, we
	// blindly lowercase the key (which is guaranteed to work in the
	// Inject/Extract sense per the OpenTracing spec).
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

func (w MetadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w MetadataReaderWriter) Keys() []string {
	keys := make([]string, 0, len(w.MD))
	for k := range w.MD {
		keys = append(keys, k)
	}
	return keys
}
