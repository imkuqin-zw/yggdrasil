// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package application

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/governor"
	"github.com/imkuqin-zw/yggdrasil/pkg/registry"
	"github.com/imkuqin-zw/yggdrasil/pkg/server"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xmap"
	"golang.org/x/sync/errgroup"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/defers"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
)

type Stage uint32

const (
	StageMin Stage = iota
	// StageBeforeStop before app start
	StageBeforeStart
	// StageBeforeStop before app stop
	StageBeforeStop
	// StageAfterStop after app stop
	StageAfterStop

	StageMax
)

var defaultShutdownTimeout = time.Second * 30

const (
	registryStateInit = iota
	registryStateDone
	registryStateCancel
)

type Endpoint struct {
	address string
	scheme  string
	Attr    map[string]string
}

func (e Endpoint) Scheme() string {
	return e.scheme
}

func (e Endpoint) Address() string {
	return e.address
}

func (e Endpoint) Metadata() map[string]string {
	return e.Attr
}

type Application struct {
	runOnce  sync.Once
	stopOnce sync.Once

	mu sync.Mutex

	optsMu  sync.RWMutex
	running bool

	server server.Server

	governor *governor.Server

	registryState int
	registry      registry.Registry

	shutdownTimeout time.Duration

	hooks map[Stage]*defers.Defer
	eg    *errgroup.Group
}

func New(inits ...Option) *Application {
	app := &Application{
		hooks: map[Stage]*defers.Defer{
			StageBeforeStart: defers.NewDefer(),
			StageBeforeStop:  defers.NewDefer(),
			StageAfterStop:   defers.NewDefer(),
		},
	}
	for i := StageMin; i < StageMax; i++ {
		app.hooks[i] = defers.NewDefer()
	}
	for _, o := range inits {
		if err := o(app); err != nil {
			logger.Fatal("fault to init app option, err: %s", err.Error())
		}
	}

	return app
}

func (app *Application) Init(opts ...Option) {
	app.optsMu.RLock()
	defer app.optsMu.RUnlock()
	if app.running {
		logger.Warn("the application has been started, and the settings are no longer applied")
		return
	}
	for _, o := range opts {
		if err := o(app); err != nil {
			logger.Fatal("fault to init application, err: %s", err.Error())
		}
	}

}

func (app *Application) Stop() error {
	var err error
	app.stopOnce.Do(func() {
		app.runHooks(StageBeforeStop)
		defer func() {
			app.runHooks(StageAfterStop)
			defers.Done()
		}()
		app.deregister()
		err = app.stopServers()
	})
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) Run() error {
	var err error
	app.runOnce.Do(func() {
		app.optsMu.Lock()
		app.running = true
		app.optsMu.Unlock()
		app.waitSignals()
		if err = app.startServers(context.Background()); err != nil {
			return
		}
		logger.Info("app shutdown")
	})

	return err
}

func (app *Application) runHooks(k Stage) {
	hooks, ok := app.hooks[k]
	if ok {
		hooks.Done()
	}
}

func (app *Application) register() {
	if app.registry == nil {
		return
	}
	app.mu.Lock()
	if app.registryState != registryStateInit {
		app.mu.Unlock()
		return
	}
	app.registryState = registryStateDone
	app.mu.Unlock()
	if err := app.registry.Register(context.TODO(), app); err != nil {
		logger.ErrorField("fault to register application", logger.Err(err))
		_ = app.Stop()
		return
	}
	logger.Info("application has been registered")
}

func (app *Application) deregister() {
	if app.registry == nil {
		return
	}
	app.mu.Lock()
	if app.registryState != registryStateDone {
		app.mu.Unlock()
		return
	}
	app.registryState = registryStateCancel
	app.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.TODO(), defaultShutdownTimeout)
	defer cancel()
	if err := app.registry.Deregister(ctx, app); err != nil {
		logger.ErrorField("fault to deregister application", logger.Err(err))
	}
}

func (app *Application) startServers(ctx context.Context) error {
	app.runHooks(StageBeforeStart)
	eg, ctx := errgroup.WithContext(ctx)
	waitServer := make(chan struct{})
	if app.server != nil {
		eg.Go(func() error {
			cancelCh, initedCh, stoppedCh := app.server.Serve()
			select {
			case <-initedCh:
				close(waitServer)
				select {
				case <-cancelCh:
					_ = app.Stop()
					err, _ := <-stoppedCh
					return err
				case err, _ := <-stoppedCh:
					return err
				}
			case <-cancelCh:
				_ = app.Stop()
				err, _ := <-stoppedCh
				return err
			case err, _ := <-stoppedCh:
				return err
			}
		})
	} else {
		close(waitServer)
	}
	eg.Go(func() error {
		return app.governor.Serve()
	})
	go func() {
		<-waitServer
		app.register()
	}()
	return eg.Wait()
}

func (app *Application) stopServers() error {
	eg := errgroup.Group{}
	eg.Go(func() error {
		return app.governor.Stop()
	})
	if app.server != nil {
		eg.Go(func() error {
			return app.server.Stop()
		})
	}
	return eg.Wait()
}

func (app *Application) Region() string {
	return pkg.Region()
}

func (app *Application) Zone() string {
	return pkg.Zone()
}

func (app *Application) Campus() string {
	return pkg.Campus()
}

func (app *Application) Namespace() string {
	return pkg.Namespace()
}

func (app *Application) Name() string {
	return pkg.Name()
}

func (app *Application) Version() string {
	return pkg.Version()
}

func (app *Application) Metadata() map[string]string {
	return pkg.Metadata()
}

func (app *Application) Endpoints() []registry.Endpoint {
	endpoints := make([]registry.Endpoint, 0)
	if app.server != nil {
		for _, item := range app.server.Endpoints() {
			attr := xmap.CloneStringMap(item.Metadata())
			attr[registry.MDServerKind] = pkg.ServerKindRpc
			endpoints = append(endpoints, Endpoint{
				address: item.Address(),
				scheme:  item.Scheme(),
				Attr:    attr,
			})
		}
	}
	governorInfo := app.governor.Info()
	attr := xmap.CloneStringMap(governorInfo.Attr)
	attr[registry.MDServerKind] = pkg.ServerKindGovernor
	endpoints = append(endpoints, Endpoint{
		address: governorInfo.Address,
		scheme:  governorInfo.Scheme,
		Attr:    attr,
	})
	return endpoints
}

func (app *Application) getShutdownTimeout() time.Duration {
	if app.shutdownTimeout < defaultShutdownTimeout {
		return defaultShutdownTimeout
	}
	return app.shutdownTimeout
}

func (app *Application) waitSignals() {
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, shutdownSignals...)
	go func() {
		s := <-sig
		go func() {
			<-time.After(app.getShutdownTimeout())
			os.Exit(128 + int(s.(syscall.Signal)))
		}()
		go func() {
			if err := app.Stop(); err != nil {
				logger.Errorf("fault to stop, err: %s", err.Error())
				return
			}
		}()
	}()
}
