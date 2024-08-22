package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

func init() {
	RegisterBuilder("logger", requestLogger)
}

type logEntry struct {
	r *http.Request
}

func (l *logEntry) Write(status, _ int, _ http.Header, elapsed time.Duration, _ interface{}) {
	var fields = []logger.Field{
		logger.String("method", l.r.Method),
		logger.String("path", l.r.URL.Path),
		logger.Int("status", status),
		logger.Float64("cost", float64(elapsed)/float64(time.Millisecond)),
	}
	if status < 400 {
		logger.InfoField("http access", fields...)
	} else if status < 500 {
		logger.WarnField("http access", fields...)
	} else {
		logger.ErrorField("http access", fields...)
	}
}

func (l *logEntry) Panic(v interface{}, stack []byte) {
	logger.ErrorField("http access",
		logger.String("method", l.r.Method),
		logger.String("path", l.r.RequestURI),
		logger.String("panic", fmt.Sprintf("%v", v)),
		logger.String("stack", string(stack)),
	)
}

type logFormatter struct{}

func (l *logFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &logEntry{r: r}
}

func requestLogger() func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&logFormatter{})
}
