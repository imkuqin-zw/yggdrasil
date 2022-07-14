package polaris

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	registry2 "github.com/imkuqin-zw/yggdrasil/pkg/registry"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xgo"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xmap"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/api"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/model"
	"go.uber.org/multierr"
)

func init() {
	registry2.RegisterConstructor("polaris", buildRegistry)
}

type RegistryConfig struct {
	ServiceToken string
	Protocol     *string
	Weight       *int
	Priority     *int
	TTL          *int
	Isolate      *bool
	Healthy      *bool
	Timeout      *time.Duration
	RetryCount   *int
}

type registry struct {
	cfg      RegistryConfig
	provider api.ProviderAPI
	ids      []string
}

func (r *registry) Register(ctx context.Context, info types.RegistryInstance) error {
	for _, endpoint := range info.Endpoints() {
		if err := r.registerService(ctx, info, endpoint); err != nil {
			return err
		}
	}
	return nil
}

func (r *registry) Deregister(ctx context.Context, info types.RegistryInstance) error {
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

func (r *registry) registerService(ctx context.Context, info types.RegistryInstance, endpoint types.ServerInfo) error {
	uri, err := url.Parse(endpoint.Endpoint())
	if err != nil {
		return err
	}
	host, port, err := net.SplitHostPort(uri.Host)
	if err != nil {
		return err
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	ns := namespace(info.Namespace())
	version := info.Version()
	scheme := endpoint.Scheme()
	serviceName := info.Name()
	metadata := make(map[string]string)
	if endpoint.Kind() == types.ServerKindGovernor {
		metadata["serviceName"] = serviceName
		serviceName = "yggdrasil.governor"
	} else {
		xmap.MergeKVMap(metadata, info.Metadata(), endpoint.Metadata())
	}
	service, err := r.provider.Register(&api.InstanceRegisterRequest{
		InstanceRegisterRequest: model.InstanceRegisterRequest{
			Service:      serviceName,
			ServiceToken: r.cfg.ServiceToken,
			Namespace:    ns,
			Host:         host,
			Port:         portNum,
			Protocol:     &scheme,
			Weight:       r.cfg.Weight,
			Priority:     r.cfg.Priority,
			Version:      &version,
			Metadata:     metadata,
			Isolate:      r.cfg.Isolate,
			Healthy:      r.cfg.Healthy,
			TTL:          r.cfg.TTL,
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
	instanceID := service.InstanceID
	xgo.Go(func() {
		r.heartbeat(ctx, instanceID, ns, serviceName, host, portNum)
	}, nil)
	r.ids = append(r.ids, instanceID)
	return nil
}

func (r *registry) deregisterService(_ context.Context, info types.RegistryInstance, ID string, endpoint types.ServerInfo) error {
	uri, err := url.Parse(endpoint.Endpoint())
	if err != nil {
		return err
	}
	host, port, err := net.SplitHostPort(uri.Host)
	if err != nil {
		return err
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	serviceName := fmt.Sprintf("%s_%s", info.Name(), uri.Scheme)
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

func (r *registry) heartbeat(ctx context.Context, ID, namespace, serviceName, host string, port int) {
	ticker := time.NewTicker(time.Second * time.Duration(*r.cfg.TTL))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := r.provider.Heartbeat(&api.InstanceHeartbeatRequest{
				InstanceHeartbeatRequest: model.InstanceHeartbeatRequest{
					Service:      serviceName,
					ServiceToken: r.cfg.ServiceToken,
					Namespace:    namespace,
					Host:         host,
					Port:         port,
					InstanceID:   ID,
					Timeout:      r.cfg.Timeout,
					RetryCount:   r.cfg.RetryCount,
				},
			})
			if err != nil {
				log.Error(err.Error())
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *registry) Name() string {
	return "polaris"
}

func buildRegistry() types.Registry {
	cfg := RegistryConfig{}
	if err := config.Scan("yggdrasil.registry.polaris", &cfg); err != nil {
		log.Fatalf("fault to load config, err: %+v", err)
		return nil
	}
	if cfg.TTL == nil || *cfg.TTL == 0 {
		cfg.TTL = &DefaultTTL
	}
	ctx, err := Context()
	if err != nil {
		log.Fatalf("fault to build provider api, err: %+v", err)
		return nil
	}
	return &registry{cfg: cfg, provider: api.NewProviderAPIByContext(ctx)}
}

func namespace(namespace string) string {
	if len(namespace) == 0 {
		return DefaultNamespace
	}
	return namespace
}
