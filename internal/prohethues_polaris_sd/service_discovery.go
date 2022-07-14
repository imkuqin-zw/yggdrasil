package prohethues_polaris_sd

import (
	"fmt"
	"regexp"
	"time"

	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/api"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/model"
)

const (
	DefaultNamespace   = "default"
	DefaultServiceName = "yggdrasil.governor"
)

type ServiceConfig struct {
	Namespace   string
	ServiceName string
	Region      string
	Zone        string
	Campus      string
}

type ServiceDiscovery struct {
	RegionRegx  *regexp.Regexp
	ZoneRegx    *regexp.Regexp
	CampusRegx  *regexp.Regexp
	namespace   string
	serviceName string
	api         api.ConsumerAPI
}

func NewSD() *ServiceDiscovery {
	cfg := &ServiceConfig{}
	if err := config2.Scan("service", cfg); err != nil {
		log.Fatalf("fault to load service config, err: %s", err)
	}
	sd, err := newSD(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return sd
}

func newSD(cfg *ServiceConfig) (*ServiceDiscovery, error) {
	sdkCtx, err := Context()
	if nil != err {
		return nil, err
	}
	consumerAPI := api.NewConsumerAPIByContext(sdkCtx)
	namespace := DefaultNamespace
	if cfg.Namespace != "" {
		namespace = cfg.Namespace
	}
	serviceName := DefaultServiceName
	if cfg.ServiceName != "" {
		serviceName = cfg.ServiceName
	}
	sd := &ServiceDiscovery{api: consumerAPI, namespace: namespace, serviceName: serviceName}
	if cfg.Region != "" {
		if sd.RegionRegx, err = regexp.Compile(cfg.Region); err != nil {
			return nil, err
		}
	}
	if cfg.Zone != "" {
		if sd.ZoneRegx, err = regexp.Compile(cfg.Zone); err != nil {
			return nil, err
		}
	}
	if cfg.Campus != "" {
		if sd.CampusRegx, err = regexp.Compile(cfg.Campus); err != nil {
			return nil, err
		}
	}
	return sd, nil
}

func (sd *ServiceDiscovery) doWatch() (<-chan model.SubScribeEvent, error) {
	watchRequest := &api.WatchServiceRequest{}
	watchRequest.Key = model.ServiceKey{
		Namespace: sd.namespace,
		Service:   sd.serviceName,
	}
	resp, err := sd.api.WatchService(watchRequest)
	if nil != err {
		return nil, err
	}
	return resp.EventChannel, nil
}

func (sd *ServiceDiscovery) GetAllInstance() ([]Instance, error) {
	instancesRequest := api.GetAllInstancesRequest{}
	instancesRequest.Namespace = sd.namespace
	instancesRequest.Service = sd.serviceName
	resp, err := sd.api.GetAllInstances(&instancesRequest)
	if nil != err {
		return nil, err
	}
	governors := make([]Instance, 0)
	for _, instance := range resp.Instances {
		if instance.GetProtocol() == "http" || instance.GetProtocol() == "https" {
			if (sd.ZoneRegx != nil && !sd.ZoneRegx.MatchString(instance.GetZone())) ||
				(sd.RegionRegx != nil && !sd.RegionRegx.MatchString(instance.GetRegion())) ||
				(sd.CampusRegx != nil && !sd.CampusRegx.MatchString(instance.GetCampus())) {
				continue
			}

			ins := Instance{
				Namespace: sd.namespace,
				Region:    instance.GetRegion(),
				Zone:      instance.GetZone(),
				Campus:    instance.GetCampus(),
				Endpoint:  fmt.Sprintf("%s:%d", instance.GetHost(), instance.GetPort()),
			}
			if md := instance.GetMetadata(); md != nil {
				ins.ServiceName, _ = md["serviceName"]
			}
			governors = append(governors, ins)
		}
	}
	return governors, err
}

func (sd *ServiceDiscovery) Watch() (<-chan []Instance, error) {
	ch, err := sd.doWatch()
	if err != nil {
		return nil, err
	}

	instancesCh := make(chan []Instance, 1)
	go func() {
		defer close(instancesCh)
		for {
			select {
			case w, ok := <-ch:
				if !ok {
					return
				}
				e, ok := w.(*model.InstanceEvent)
				if ok {
					if e.AddEvent == nil && e.DeleteEvent == nil {
						continue
					}
				}
				backoff := time.Second * 20
				for {
					instances, err := sd.GetAllInstance()
					if err != nil {
						log.Warnf("fault to get all instance, %+v", err)
						time.Sleep(backoff)
						if backoff < time.Minute*10 {
							backoff *= 2
						}
						continue
					}
					instancesCh <- instances
					break
				}
			}
		}
	}()
	return instancesCh, nil
}
