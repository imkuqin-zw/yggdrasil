package polaris

import (
	"sync"

	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
)

var (
	polarisContext      api.SDKContext
	polarisConfig       config.Configuration
	mutexPolarisContext sync.Mutex
	oncePolarisConfig   sync.Once

	DefaultNamespace = "default"
	// DefaultTTL default ttl value when ttl is not set
	DefaultTTL = 20
)

// PolarisContext get or init the global polaris context
func Context() (api.SDKContext, error) {
	mutexPolarisContext.Lock()
	defer mutexPolarisContext.Unlock()
	if nil != polarisContext {
		return polarisContext, nil
	}
	var err error
	polarisContext, err = api.InitContextByConfig(Configuration())
	return polarisContext, err
}

// PolarisConfig get or init the global polaris configuration
func Configuration() config.Configuration {
	oncePolarisConfig.Do(func() {
		cfg := &config.ConfigurationImpl{}
		cfg.Init()
		if err := config2.Scan("polaris", cfg); err != nil {
			log.Fatalf("fault to load polaris config, err: %s", err)
		}
		cfg.SetDefault()
		polarisConfig = cfg
	})
	return polarisConfig
}