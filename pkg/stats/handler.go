package stats

import (
	"context"
	"strings"
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

var (
	mu             sync.RWMutex
	handlerBuilder = map[string]HandlerBuilder{}
	svrOnce        sync.Once
	svrHandler     Handler
	cliOnce        sync.Once
	cliHandler     Handler
)

func RegisterHandlerBuilder(name string, builder HandlerBuilder) {
	mu.Lock()
	defer mu.Unlock()
	handlerBuilder[name] = builder
}

func GetHandlerBuilder(name string) HandlerBuilder {
	mu.Lock()
	defer mu.Unlock()
	builder, ok := handlerBuilder[name]
	if !ok {
		return nil
	}
	return builder
}

type Handler interface {
	// TagRPC can attach some information to the given context.
	// The context used for the rest lifetime of the RPC will be derived from
	// the returned context.
	TagRPC(context.Context, RPCTagInfo) context.Context
	// HandleRPC processes the RPC stats.
	HandleRPC(context.Context, RPCStats)

	// TagChannel can attach some information to the given context.
	// The returned context will be used for stats handling.
	// For channel stats handling, the context used in HandleChannel for this
	// channel will be derived from the context returned.
	// For RPC stats handling,
	//  - On server side, the context used in HandleRPC for all RPCs on this
	// channel will be derived from the context returned.
	//  - On client side, the context is not derived from the context returned.
	TagChannel(context.Context, ChanTagInfo) context.Context
	// HandleChannel processes the Channel stats.
	HandleChannel(context.Context, ChanStats)
}

type HandlerBuilder func(isServer bool) Handler

type handlerChain struct {
	handlers []Handler
}

func (h *handlerChain) TagRPC(ctx context.Context, info RPCTagInfo) context.Context {
	for _, handler := range h.handlers {
		ctx = handler.TagRPC(ctx, info)
	}
	return ctx
}

func (h *handlerChain) HandleRPC(ctx context.Context, rs RPCStats) {
	for _, handler := range h.handlers {
		handler.HandleRPC(ctx, rs)
	}
}

func (h *handlerChain) TagChannel(ctx context.Context, info ChanTagInfo) context.Context {
	for _, handler := range h.handlers {
		ctx = handler.TagChannel(ctx, info)
	}
	return ctx
}

func (h *handlerChain) HandleChannel(ctx context.Context, cs ChanStats) {
	for _, handler := range h.handlers {
		handler.HandleChannel(ctx, cs)
	}
}

func GetServerHandler() Handler {
	svrOnce.Do(func() {
		names := config.Get(config.Join(config.KeyStats, "server")).String("")
		h := &handlerChain{handlers: make([]Handler, 0)}
		for _, name := range strings.Split(names, ",") {
			if name == "" {
				continue
			}
			builder := GetHandlerBuilder(name)
			if builder == nil {
				logger.Warnf("fault to get stats handler builder: %s", name)
				continue
			}
			h.handlers = append(h.handlers, builder(true))
		}
		svrHandler = h
	})
	return svrHandler
}

func GetClientHandler() Handler {
	cliOnce.Do(func() {
		names := config.Get(config.Join(config.KeyStats, "client")).String("")
		h := &handlerChain{handlers: make([]Handler, 0)}
		for _, name := range strings.Split(names, ",") {
			builder := GetHandlerBuilder(name)
			if builder == nil {
				logger.Warnf("fault to get stats handler builder: %s", name)
				continue
			}
			h.handlers = append(h.handlers, builder(false))
		}
		cliHandler = h
	})
	return cliHandler
}
