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

package governor

import (
	"fmt"

	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"
)

type Config struct {
	Host string
	Port uint64
}

// Address
func (config *Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

// Build
func (config *Config) Build() *Server {
	var err error
	config.Host, err = xnet.Extract(config.Host)
	if err != nil {
		log.Fatalf("create governor server, error: %s", err.Error())
	}

	return newServer(config)
}
