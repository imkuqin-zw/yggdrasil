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

package prohethues_polaris_sd

import (
	"sync"

	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/api"
	"github.com/imkuqin-zw/yggdrasil/third_party/polarismesh/polaris-go/pkg/config"
)

var (
	polarisContext      api.SDKContext
	polarisConfig       config.Configuration
	mutexPolarisContext sync.Mutex
	oncePolarisConfig   sync.Once
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
