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

package polaris

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	registry2 "github.com/imkuqin-zw/yggdrasil/pkg/registry"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xmap"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/model"
	"go.uber.org/multierr"
)

func init() {
	registry2.RegisterBuilder(name, buildRegistry)
}

type RegistryConfig struct {
	ServiceToken     string
	Protocol         *string
	Weight           *int
	Priority         *int
	TTL              int `default:"5"`
	Isolate          *bool
	Healthy          *bool
	Timeout          *time.Duration
	RetryCount       *int
	RegisterGovernor bool
}

type registry struct {
	cfg      RegistryConfig
	provider api.ProviderAPI
	ids      []string
}

func (r *registry) Register(ctx context.Context, info registry2.Instance) error {
	for _, endpoint := range info.Endpoints() {
		meta := endpoint.Metadata()
		governor := false
		serverKind := meta[registry2.MDServerKind]
		if serverKind == pkg.ServerKindGovernor {
			if !r.cfg.RegisterGovernor {
				continue
			}
			governor = true
		}
		if err := r.registerService(ctx, info, endpoint, governor); err != nil {
			return err
		}
	}
	return nil
}

func (r *registry) Deregister(ctx context.Context, info registry2.Instance) error {
	endpoints := info.Endpoints()
	var errs []error
	for idx, id := range r.ids {
		if err := r.deregisterService(ctx, info, id, endpoints[idx]); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return multierr.Combine(errs...)
}

func (r *registry) registerService(_ context.Context, info registry2.Instance, endpoint registry2.Endpoint, governor bool) error {
	host, port, err := net.SplitHostPort(endpoint.Address())
	if err != nil {
		return err
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	ns := namespace(info.Namespace())
	version := info.Version()
	protocol := endpoint.Scheme()
	serviceName := info.Name()
	if governor {
		serviceName = fmt.Sprintf("%s.%s", serviceName, "governor")
	}
	metadata := make(map[string]string)
	xmap.MergeKVMap(metadata, info.Metadata(), endpoint.Metadata())

	service, err := r.provider.RegisterInstance(&api.InstanceRegisterRequest{
		InstanceRegisterRequest: model.InstanceRegisterRequest{
			Service:      serviceName,
			ServiceToken: r.cfg.ServiceToken,
			Namespace:    ns,
			Host:         host,
			Port:         portNum,
			Protocol:     &protocol,
			Weight:       r.cfg.Weight,
			Priority:     r.cfg.Priority,
			Version:      &version,
			Metadata:     metadata,
			Isolate:      r.cfg.Isolate,
			Healthy:      r.cfg.Healthy,
			TTL:          &r.cfg.TTL,
			Timeout:      r.cfg.Timeout,
			RetryCount:   r.cfg.RetryCount,
			Location: &model.Location{
				Region: info.Region(),
				Zone:   info.Zone(),
				Campus: info.Campus(),
			},
		},
	})
	if err != nil {
		return err
	}
	//instanceID := service.InstanceID
	//xgo.Go(func() { r.heartbeat(ctx, info, endpoint) }, nil)
	r.ids = append(r.ids, service.InstanceID)
	return nil
}

func (r *registry) deregisterService(_ context.Context, info registry2.Instance, ID string, endpoint registry2.Endpoint) error {
	host, port, err := net.SplitHostPort(endpoint.Address())
	if err != nil {
		return err
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	serviceName := info.Name()
	// Deregister
	err = r.provider.Deregister(
		&api.InstanceDeRegisterRequest{
			InstanceDeRegisterRequest: model.InstanceDeRegisterRequest{
				Service:      serviceName,
				ServiceToken: r.cfg.ServiceToken,
				Namespace:    namespace(info.Namespace()),
				InstanceID:   ID,
				Host:         host,
				Port:         portNum,
				Timeout:      r.cfg.Timeout,
				RetryCount:   r.cfg.RetryCount,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

//func (r *registry) heartbeat(ctx context.Context, info registry2.Instance, endpoint registry2.Endpoint) {
//	ticker := time.NewTicker(time.Second * time.Duration(r.cfg.TTL))
//	defer ticker.Stop()
//
//	for {
//		select {
//		case <-ticker.C:
//			err := r.registerService(ctx, info, endpoint)
//			if err != nil {
//				logger.ErrorFiled("fault to register", logger.Err(err))
//			}
//		case <-ctx.Done():
//			return
//		}
//	}
//}

func (r *registry) Name() string {
	return "polaris"
}

func buildRegistry() registry2.Registry {
	cfg := RegistryConfig{}
	if err := config.Scan(configKeyRegistry, &cfg); err != nil {
		logger.FatalFiled("fault to load config", logger.Err(err))
		return nil
	}
	ctx, err := Context()
	if err != nil {
		logger.FatalFiled("fault to build provider api", logger.Err(err))
		return nil
	}
	return &registry{cfg: cfg, provider: api.NewProviderAPIByContext(ctx)}
}
