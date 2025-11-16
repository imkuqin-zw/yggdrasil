package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlogHandlerBasic(t *testing.T) {
	// Test basic slog handler functionality
	var buf bytes.Buffer
	logger := NewLogger(LvDebug, &slogTestWriter{buf: &buf})
	handler := NewSlogHandler(logger, nil)

	// Test basic logging methods
	assert.NotNil(t, handler)

	// Test WithGroup method
	handlerWithGroup := handler.WithGroup("test-group")
	assert.NotNil(t, handlerWithGroup)

	// Test WithAttrs method
	handlerWithAttrs := handler.WithAttrs([]slog.Attr{slog.String("test", "value")})
	assert.NotNil(t, handlerWithAttrs)
}

// slogTestWriter is a custom writer that outputs slog-compatible JSON
type slogTestWriter struct {
	buf *bytes.Buffer
}

func (w *slogTestWriter) Write(lv Level, t time.Time, msg string, ext ...[]byte) {
	// Create slog-compatible JSON structure
	entry := map[string]any{
		"level":   lv.String(),
		"time":    t.Format(time.RFC3339Nano),
		"message": msg,
	}

	// Parse additional fields if any
	if len(ext) > 0 && len(ext[0]) > 0 {
		var additional map[string]any
		if err := json.Unmarshal(ext[0], &additional); err == nil {
			for k, v := range additional {
				entry[k] = v
			}
		}
	}

	data, _ := json.Marshal(entry)
	w.buf.Write(data)
	w.buf.WriteByte('\n')
}
