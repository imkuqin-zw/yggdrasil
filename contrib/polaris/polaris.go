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
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg"
	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
	"github.com/polarismesh/polaris-go/pkg/model"
)

var (
	name = "polaris"

	configKeyBase       = name
	configKeyClient     = config2.Join(configKeyBase, "client")
	configKeySourceFile = config2.Join(configKeyBase, "source_file")
	configKeyRegistry   = config2.Join(configKeyBase, "registry")
	configKeyDst        = config2.Join(configKeyBase, "dst")
	configKeyClientDst  = config2.Join(config2.KeyClientInstance, name, "dst")
)

var (
	polarisCallerServiceKey   = "polaris.request.caller.service"
	polarisCallerNamespaceKey = "polaris.request.caller.namespace"
)

var (
	polarisContext      api.SDKContext
	polarisConfig       config.Configuration
	mutexPolarisContext sync.Mutex
	oncePolarisConfig   sync.Once

	DefaultNamespace = "default"
)

// Context get or init the global polaris context
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

// Configuration get or init the global polaris configuration
func Configuration() config.Configuration {
	oncePolarisConfig.Do(func() {
		cfg := &config.ConfigurationImpl{}
		cfg.Init()
		if err := config2.Scan(configKeyClient, cfg); err != nil {
			logger.FatalField("fault to load polaris config", logger.Err(err))
		}
		cfg.SetDefault()
		polarisConfig = cfg
	})
	return polarisConfig
}

func buildSourceInfo() *model.ServiceInfo {
	return &model.ServiceInfo{
		Namespace: namespace(pkg.Namespace()),
		Service:   pkg.Name(),
		Metadata:  pkg.Metadata(),
	}
}

func namespace(namespace string) string {
	if len(namespace) == 0 {
		return DefaultNamespace
	}
	return namespace
}
