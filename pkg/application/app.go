package application

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/defers"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/errgroup"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
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

type Application struct {
	runOnce  sync.Once
	stopOnce sync.Once

	optsMu  sync.RWMutex
	running bool

	servers []types.Server

	registry types.Registry

	shutdownTimeout time.Duration

	hooks map[Stage]*defers.Defer
	eg    *errgroup.Group
}

func New(inits ...Option) *Application {
	app := &Application{
		eg: errgroup.WithCancel(context.Background()),
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
			log.Fatal("fault to init app option, err: %s", err.Error())
		}
	}

	return app
}

func (app *Application) Init(opts ...Option) {
	app.optsMu.RLock()
	defer app.optsMu.RUnlock()
	if app.running {
		log.Warn("the application has been started, and the settings are no longer applied")
		return
	}
	for _, o := range opts {
		if err := o(app); err != nil {
			log.Fatal("fault to init application, err: %s", err.Error())
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
		err = app.stopServers()
	})
	if err != nil {
		return err
	}
	log.Info("application stopped")
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
		log.Info("app shutdown")
	})

	return err
}

func (app *Application) runHooks(k Stage) {
	hooks, ok := app.hooks[k]
	if ok {
		hooks.Done()
	}
}

func (app *Application) startServers(ctx context.Context) error {
	app.runHooks(StageBeforeStart)
	eg := errgroup.WithContext(ctx)
	var governorMeta []string
	for _, s := range app.servers {
		if s.Info().Kind() != types.ServerKindGovernor {
			continue
		}
		governorMeta = append(governorMeta, s.Info().Endpoint())
	}
	if len(governorMeta) > 0 {
		data, _ := json.Marshal(governorMeta)
		pkg.AddMetadata("governor", xstrings.Bytes2str(data))
	}
	for _, s := range app.servers {
		s := s
		eg.Go(func(ctx context.Context) (err error) {
			if log.Enable(types.LvInfo) {
				info := s.Info()
				data, _ := json.Marshal(map[string]interface{}{
					"kind":     info.Kind(),
					"endpoint": info.Endpoint(),
				})
				log.Infof("server start  %s", string(data))
			}
			err = s.Serve()
			return
		})
	}
	if app.registry != nil {
		if err := app.registry.Register(ctx, app); err != nil {
			return err
		}
		defer func() {
			if err := app.registry.Deregister(ctx, app); err != nil {
				log.Errorf("fault to deregister, err: %+v", err)
			}
		}()
	}
	return eg.Wait()
}

func (app *Application) stopServers() error {
	eg := &errgroup.Group{}
	for _, s := range app.servers {
		s := s
		eg.Go(func(ctx context.Context) (err error) {
			err = s.Stop()
			return
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

func (app *Application) Endpoints() []types.ServerInfo {
	endpoints := make([]types.ServerInfo, 0)
	for _, svr := range app.servers {
		if svr.Info().Kind() == types.ServerKindRpc || svr.Info().Kind() == types.ServerKindGovernor {
			endpoints = append(endpoints, svr.Info())
		}
	}
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
				log.Errorf("fault to stop, err: %s", err.Error())
				return
			}
		}()
	}()
}
