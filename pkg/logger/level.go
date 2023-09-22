package logger

import (
	"bytes"
	"fmt"
)

type Level int8

// String returns a lower-case ASCII representation of the global level.
func (l Level) String() string {
	switch l {
	case LvDebug:
		return "debug"
	case LvInfo:
		return "info"
	case LvWarn:
		return "warn"
	case LvError:
		return "error"
	case LvFault:
		return "fatal"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

// MarshalText marshals the Level to text. Note that the text representation
// drops the -Level suffix (see example).
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText unmarshals text to a level. Like MarshalText, UnmarshalText
// expects the text representation of a Level to drop the -Level suffix (see
// example).
//
// In particular, this makes it easy to configure logging levels using YAML,
// TOML, or JSON files.
func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errUnmarshalNilLevel
	}
	if !l.unmarshalText(text) && !l.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized level: %q", text)
	}
	return nil
}

func (l *Level) unmarshalText(text []byte) bool {
	switch string(text) {
	case "debug", "DEBUG":
		*l = LvDebug
	case "info", "INFO", "": // make the zero value useful
		*l = LvInfo
	case "warn", "WARN":
		*l = LvWarn
	case "error", "ERROR":
		*l = LvError
	case "fatal", "FATAL":
		*l = LvFault
	default:
		return false
	}
	return true
}

func (l Level) Enable(level Level) bool {
	return l <= level
}

const (
	LvDebug Level = iota - 1
	LvInfo
	LvWarn
	LvError
	LvFault
)
