package zap

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func Test_Logger(t *testing.T) {
	lg := (&Config{Console: struct {
		Enable  bool
		Encoder *zapcore.EncoderConfig
	}{Enable: true}}).Build()
	var dd = struct {
		A string
		B int
	}{"a", 2}
	lg.Warn("fdaf", "k1", 1, "k2", dd)
	lg.Fatalf()
}
