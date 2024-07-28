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

	config2 "github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xgo"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xstrings"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
	"github.com/polarismesh/polaris-go/pkg/model"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Namespace string
	Group     string
	Filename  string
	Polaris   config.ConfigFileConfigImpl
}

type polaris struct {
	namespace string
	group     string
	filename  string
	closeOnce sync.Once
	closeCh   chan struct{}
	closedCh  chan struct{}
	api       api.ConfigFileAPI
	cfgFile   model.ConfigFile
	firstKey  map[string]interface{}
}

func (c *polaris) Name() string {
	return name
}

func (c *polaris) Read() (source.SourceData, error) {
	var err error
	c.cfgFile, err = c.api.GetConfigFile(c.namespace, c.group, c.filename)
	if err != nil {
		return nil, err
	}
	content := xstrings.Str2bytes(c.cfgFile.GetContent())
	v := map[string]interface{}{}
	if err := yaml.Unmarshal(content, &v); err != nil {
		return nil, err
	}
	c.firstKey = make(map[string]interface{}, len(v))
	for key := range v {
		c.firstKey[key] = nil
	}
	return source.NewBytesSourceData(source.PriorityRemote, content, yaml.Unmarshal), nil
}

func (c *polaris) Changeable() bool {
	return true
}

func (c *polaris) Watch() (<-chan source.SourceData, error) {
	csdCh := make(chan source.SourceData, 1)
	changeEventChan := c.cfgFile.AddChangeListenerWithChannel()
	xgo.Go(func() {
		defer func() {
			close(csdCh)
			close(c.closedCh)
		}()
		for {
			select {
			case event := <-changeEventChan:
				kv := make(map[string]interface{})
				content := xstrings.Str2bytes(event.NewValue)
				if err := yaml.Unmarshal(content, &kv); err != nil {
					logger.ErrorField("fault to unmarshal config, err: %s", logger.Err(err))
					continue
				}
				for key := range c.firstKey {
					if _, ok := kv[key]; !ok {
						kv[key] = nil
					}
				}
				c.firstKey = make(map[string]interface{}, len(kv))
				for k := range kv {
					c.firstKey[k] = nil
				}
				csdCh <- source.NewBytesSourceData(source.PriorityRemote, content, yaml.Unmarshal)
			case _, _ = <-c.closeCh:
				return
			}
		}
	}, func(r interface{}) {
		close(csdCh)
		close(c.closedCh)
	})

	return csdCh, nil
}

func (c *polaris) Close() error {
	c.closeOnce.Do(func() {
		close(c.closeCh)
	})
	<-c.closedCh
	return nil
}

func NewConfig(namespace, group, filename string) (source.Source, error) {
	c, err := Context()
	if err != nil {
		return nil, err
	}
	sc := &polaris{namespace: namespace, group: group, filename: filename, api: api.NewConfigFileAPIBySDKContext(c)}
	return sc, nil
}

func LoadConfig(appName string) error {
	cfg := struct {
		Filenames []string
	}{}
	if err := config2.Scan(configKeySourceFile, &cfg); err != nil {
		return err
	}
	namespace := config2.Get(config2.KeyAppNamespace).String("default")
	var sources []source.Source
	for _, filename := range cfg.Filenames {
		sc, err := NewConfig(namespace, appName, filename)
		if err != nil {
			return err
		}
		sources = append(sources, sc)
	}
	if len(sources) > 0 {
		return config2.LoadSource(sources...)
	}
	return nil
}
