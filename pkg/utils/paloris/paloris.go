package paloris

import (
	"sync"

	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
)

var (
	polarisContext      api.SDKContext
	polarisConfig       config.Configuration
	mutexPolarisContext sync.Mutex
	oncePolarisConfig   sync.Once
)

// PolarisContext get or init the global polaris context
func PolarisContext() (api.SDKContext, error) {
	mutexPolarisContext.Lock()
	defer mutexPolarisContext.Unlock()
	if nil != polarisContext {
		return polarisContext, nil
	}
	var err error
	polarisContext, err = api.InitContextByConfig(PolarisConfig())
	return polarisContext, err
}

// PolarisConfig get or init the global polaris configuration
func PolarisConfig() config.Configuration {
	oncePolarisConfig.Do(func() {

	})
	return polarisConfig
}
