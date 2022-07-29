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

package xtls

import (
	"crypto/tls"
)

// SSLConfig struct stores the necessary info for SSL configuration
type SSLConfig struct {
	CipherPlugin string
	VerifyPeer   bool
	CipherSuites string
	MinVersion   string
	MaxVersion   string
	CA           []byte
	Cert         []byte
	Key          []byte
	CertPWD      string
	ServerName   string
}

// ClientTLSConfig function gets client side TLS config
func (c *SSLConfig) ClientTLSConfig() (*tls.Config, error) {
	return getTLSConfig(c, "client")
}

// ServerTLSConfig function gets server side TLD config
func (c *SSLConfig) ServerTLSConfig() (*tls.Config, error) {
	return getTLSConfig(c, "server")
}
