package md

import "context"

type inKey struct{}
type outKey struct{}

// WithContext creates a new context with incoming md attached.
func WithInContext(ctx context.Context, md MD) context.Context {
	oldMd, ok := FromInContext(ctx)
	if ok {
		return context.WithValue(ctx, inKey{}, Join(oldMd, md))
	}
	return context.WithValue(ctx, inKey{}, md)
}

func FromInContext(ctx context.Context) (md MD, ok bool) {
	if ctx == nil {
		return MD{}, false
	}

	md, ok = ctx.Value(inKey{}).(MD)
	return
}

// WithContext creates a new context with outgoing md attached.
func WithOutContext(ctx context.Context, md MD) context.Context {
	oldMd, ok := FromOutContext(ctx)
	if ok {
		return context.WithValue(ctx, outKey{}, Join(oldMd, md))
	}
	return context.WithValue(ctx, outKey{}, md)
}

func FromOutContext(ctx context.Context) (md MD, ok bool) {
	if ctx == nil {
		return MD{}, false
	}

	md, ok = ctx.Value(outKey{}).(MD)
	return
}
