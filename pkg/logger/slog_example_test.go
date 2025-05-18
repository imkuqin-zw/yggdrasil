package logger

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"
)

type Password string

func (p Password) LogValue() slog.Value {
	return slog.StringValue("REDACTED")
}

func Example_slog() {
	var buff = bytes.NewBuffer(nil)
	lg := NewLogger(LvDebug, NewMemoryWriter(buff, false, nil))
	sl := slog.New(NewSlogHandler(lg, nil))
	ctx := context.Background()

	sl.Info("user", "name", "Al", "secret", Password("secret"))
	sl.Error("oops", "err", net.ErrClosed, "status", 500)
	sl.LogAttrs(
		ctx,
		slog.LevelError,
		"oops",
		slog.Any("err", net.ErrClosed),
		slog.Int("status", 500),
	)
	sl.Info("message",
		slog.Group("group",
			slog.Float64("pi", 3.14),
			slog.Duration("1min", time.Minute),
		),
	)
	sl.WithGroup("s").LogAttrs(
		ctx,
		slog.LevelWarn,
		"warn msg", // message
		slog.Uint64("u", 1),
		slog.Any("m", map[string]any{
			"foo": "bar",
		}))
	sl.LogAttrs(ctx, slog.LevelDebug, "not show up")
	reader := bufio.NewReader(buff)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		fmt.Println(string(line))
	}
	// Output:
	// {"level":"info","msg":"user","name":"Al","secret":"REDACTED"}
	// {"level":"error","msg":"oops","err":"use of closed network connection","status":500}
	// {"level":"error","msg":"oops","err":"use of closed network connection","status":500}
	// {"level":"info","msg":"message","group":{"pi":3.14,"1min":60000}}
	// {"level":"warn","msg":"warn msg","u":1,"m":{"foo":"bar"}}
	// {"level":"debug","msg":"not show up"}
}
