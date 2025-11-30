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

package xds

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Address != "localhost:15010" {
		t.Errorf("expected server address localhost:15010, got %s", cfg.Server.Address)
	}

	if cfg.Server.TLSPort != 15011 {
		t.Errorf("expected TLS port 15011, got %d", cfg.Server.TLSPort)
	}

	if !cfg.Resources.CDS {
		t.Error("expected CDS to be enabled by default")
	}

	if !cfg.Resources.EDS {
		t.Error("expected EDS to be enabled by default")
	}

	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", cfg.Timeout)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "missing server address",
			config: Config{
				Node: NodeInfo{
					Cluster: "test-cluster",
				},
			},
			wantErr: true,
		},
		{
			name: "missing cluster",
			config: Config{
				Server: ServerConfig{
					Address: "localhost:15010",
				},
			},
			wantErr: true,
		},
		{
			name: "TLS enabled without CA cert",
			config: Config{
				Server: ServerConfig{
					Address: "localhost:15010",
				},
				Node: NodeInfo{
					Cluster: "test-cluster",
				},
				TLS: TLSConfig{
					Enabled: true,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetServerAddress(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{
			name: "non-TLS",
			config: Config{
				Server: ServerConfig{
					Address: "localhost:15010",
					UseTLS:  false,
				},
			},
			want: "localhost:15010",
		},
		{
			name: "TLS enabled",
			config: Config{
				Server: ServerConfig{
					Address: "localhost:15010",
					UseTLS:  true,
					TLSPort: 15011,
				},
				TLS: TLSConfig{
					Enabled: true,
				},
			},
			want: "localhost:15010",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetServerAddress()
			if got != tt.want {
				t.Errorf("GetServerAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
