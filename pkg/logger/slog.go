package logger

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*SlogHandler)(nil)

type SlogHandler struct {
	lg *Logger
	// List of unapplied groups.
	//
	// These are applied only if we encounter a real field
	// to avoid creating empty namespaces -- which is disallowed by slog's
	// usage contract.
	groups []string
}

func NewSlogHandler(lg *Logger, groups []string) slog.Handler {
	return &SlogHandler{lg: lg, groups: groups}
}

// Enabled reports whether the handler handles records at the given level.
func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.lg.Enable(convertSlogLevel(level))
}

// Handle handles the Record.
func (h *SlogHandler) Handle(ctx context.Context, record slog.Record) error {
	fields := make([]Field, 0, record.NumAttrs()+len(h.groups)+1)
	fields = append(fields, Context(ctx))
	var addedNamespace bool
	record.Attrs(func(attr slog.Attr) bool {
		f := convertSlogAttrToField(attr)
		if !addedNamespace && len(h.groups) > 0 && f.Type == SkipType {
			// Namespaces are added only if at least one field is present
			// to avoid creating empty groups.
			fields = h.appendGroups(fields)
			addedNamespace = true
		}
		fields = append(fields, f)
		return true
	})
	h.lg.write(convertSlogLevel(record.Level), record.Time, record.Message, nil, fields...)
	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make([]Field, 0, len(attrs)+len(h.lg.fields))
	var addedNamespace bool
	for _, attr := range attrs {
		f := convertSlogAttrToField(attr)
		if !addedNamespace && len(h.groups) > 0 && f.Type != SkipType {
			// Namespaces are added only if at least one field is present
			// to avoid creating empty groups.
			fields = h.appendGroups(fields)
			addedNamespace = true
		}
		fields = append(fields, f)
	}

	for _, attr := range attrs {
		f := convertSlogAttrToField(attr)
		fields = append(fields, f)
	}
	cloned := *h
	cloned.lg = h.lg.WithFields(fields...)
	if addedNamespace {
		// These groups have been applied so we can clear them.
		cloned.groups = nil
	}
	return &cloned
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
func (h *SlogHandler) WithGroup(group string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = group

	cloned := *h
	cloned.groups = newGroups
	return &cloned
}

func (h *SlogHandler) appendGroups(fields []Field) []Field {
	for _, g := range h.groups {
		fields = append(fields, Namespace(g))
	}
	return fields
}

func convertSlogAttrToField(attr slog.Attr) Field {
	if attr.Equal(slog.Attr{}) {
		// Ignore empty attrs.
		return Skip()
	}

	switch attr.Value.Kind() {
	case slog.KindBool:
		return Bool(attr.Key, attr.Value.Bool())
	case slog.KindDuration:
		return Duration(attr.Key, attr.Value.Duration())
	case slog.KindFloat64:
		return Float64(attr.Key, attr.Value.Float64())
	case slog.KindInt64:
		return Int64(attr.Key, attr.Value.Int64())
	case slog.KindString:
		return String(attr.Key, attr.Value.String())
	case slog.KindTime:
		return Time(attr.Key, attr.Value.Time())
	case slog.KindUint64:
		return Uint64(attr.Key, attr.Value.Uint64())
	case slog.KindGroup:
		if attr.Key == "" {
			// Inlines recursively.
			return Inline(slogGroupObject(attr.Value.Group()))
		}
		return Object(attr.Key, slogGroupObject(attr.Value.Group()))
	case slog.KindLogValuer:
		return convertSlogAttrToField(slog.Attr{
			Key: attr.Key,
			// TODO: resolve the value in a lazy way.
			// This probably needs a new Zap field type
			// that can be resolved lazily.
			Value: attr.Value.Resolve(),
		})
	default:
		return Any(attr.Key, attr.Value.Any())
	}
}

// slogGroupObject holds all the Attrs saved in a slog.GroupValue.
type slogGroupObject []slog.Attr

func (gs slogGroupObject) MarshalLogObject(enc ObjectEncoder) error {
	for _, attr := range gs {
		convertSlogAttrToField(attr).AddTo(enc)
	}
	return nil
}

// convertSlogLevel maps slog Levels to zap Levels.
// Note that there is some room between slog levels while zap levels are continuous, so we can't 1:1 map them.
// See also https://go.googlesource.com/proposal/+/master/design/56345-structured-logging.md?pli=1#levels
func convertSlogLevel(l slog.Level) Level {
	switch {
	case l >= slog.LevelError:
		return LvError
	case l >= slog.LevelWarn:
		return LvWarn
	case l >= slog.LevelInfo:
		return LvInfo
	default:
		return LvDebug
	}
}
