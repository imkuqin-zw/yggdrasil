package md

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

type inKey struct{}
type outKey struct{}
type streamKey struct{}

type stream struct {
	mu      sync.Mutex
	header  MD
	trailer MD
}

func WithInContext(ctx context.Context, md MD) context.Context {
	oldMd, ok := ctx.Value(inKey{}).(MD)
	if ok {
		return context.WithValue(ctx, inKey{}, Join(oldMd, md))
	}
	return context.WithValue(ctx, inKey{}, md)
}

func FromInContext(ctx context.Context) (md MD, ok bool) {
	md, ok = ctx.Value(inKey{}).(MD)
	if !ok {
		return MD{}, false
	}
	return md.Copy(), true
}

func WithOutContext(ctx context.Context, md MD) context.Context {
	oldMd, ok := ctx.Value(outKey{}).(MD)
	if ok {
		return context.WithValue(ctx, outKey{}, Join(oldMd, md))
	}
	return context.WithValue(ctx, outKey{}, md)
}

func FromOutContext(ctx context.Context) (md MD, ok bool) {
	md, ok = ctx.Value(outKey{}).(MD)
	if !ok {
		return MD{}, false
	}
	return md.Copy(), true
}

func WithStreamContext(ctx context.Context) context.Context {
	_, ok := ctx.Value(streamKey{}).(*stream)
	if !ok {
		return context.WithValue(ctx, streamKey{}, &stream{})
	}
	return ctx
}

func SetTrailer(ctx context.Context, md MD) error {
	h, ok := ctx.Value(streamKey{}).(*stream)
	if !ok {
		return errors.Errorf("failed to fetch the stream from the context %v", ctx)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.trailer = Join(h.trailer, md)
	return nil
}

func FromTrailerCtx(ctx context.Context) (md MD, ok bool) {
	h, ok := ctx.Value(streamKey{}).(*stream)
	if !ok {
		return MD{}, false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.trailer.Copy(), true
}

func SetHeader(ctx context.Context, md MD) error {
	h, ok := ctx.Value(streamKey{}).(*stream)
	if !ok {
		return errors.Errorf("failed to fetch the stream from the context %v", ctx)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.header = Join(h.header, md)
	return nil
}

func FromHeaderCtx(ctx context.Context) (md MD, ok bool) {
	h, ok := ctx.Value(streamKey{}).(*stream)
	if !ok {
		return MD{}, false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.header.Copy(), true
}

func WithHeaderOptCtx(ctx context.Context) context.Context {
	s, ok := ctx.Value(streamKey{}).(*stream)
	if !ok {
		return context.WithValue(ctx, streamKey{}, &stream{header: MD{}})
	}
	s.header = MD{}
	return ctx
}

func WithTrailerOptCtx(ctx context.Context) context.Context {
	s, ok := ctx.Value(streamKey{}).(*stream)
	if !ok {
		return context.WithValue(ctx, streamKey{}, &stream{trailer: MD{}})
	}
	s.trailer = MD{}
	return ctx
}
