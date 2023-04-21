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

package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/env"
	"github.com/imkuqin-zw/yggdrasil/pkg/config/source/file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_pointerConfig(t *testing.T) {
	type Config struct {
		Val *int
	}
	cfg := &Config{}
	err := Scan("conf", cfg)
	assert.Nil(t, err)
	assert.Nil(t, cfg.Val)
	err = Set("conf.val", 1)
	assert.Nil(t, err)
	err = Scan("conf", cfg)
	assert.Nil(t, err)
	require.NotNil(t, cfg.Val)
	assert.Equal(t, 1, *cfg.Val)
}

func TestConfig_Scan(t *testing.T) {
	type Config struct {
		Target string
	}
	err := Set("yggdrasil.client.sample.grpc.target", "192.168.3.52:49613")
	require.Nil(t, err)
	c := &Config{}
	key := "yggdrasil.client.sample.grpc"
	err = Scan(key, c)
	fmt.Println(Get(""))
	fmt.Println(Get("yggdrasil"))
	fmt.Println(Get("yggdrasil.client"))
	fmt.Println(Get(key))
	require.Nil(t, err)
	fmt.Println(c)
}

func TestConfig_GetContainerDelimiterKey(t *testing.T) {
	if err := LoadSource(file.NewSource("./testdata/config.yaml", false)); err != nil {
		t.Fatal(err)
	}
	data := Get("yggdrasil.client.{example.polaris.server}.gd").Bytes([]byte("not found"))
	assert.Equal(t, []byte("not found"), data)
	data = Get("yggdrasil.client.{example.polaris.server}").Bytes()
	assert.Equal(t, []byte(`{"grpc":{"target":"127.0.0.1:30001"}}`), data)
}

func TestConfig_matchKey(t *testing.T) {
	key := "yggdrasil.client.{example.polaris.server}"
	regx, _ := regexp.Compile("{([\\w.]+)}")
	matches := make([]string, 0)
	key = regx.ReplaceAllStringFunc(key, func(s string) string {
		matches = append(matches, s[1:len(s)-1])
		return "{}"
	})
	paths := strings.Split(key, ".")
	j := 0
	for i, item := range paths {
		if item == "{}" {
			paths[i] = matches[j]
			j++
		}
	}
	assert.Equal(t, []string{"yggdrasil", "client", "example.polaris.server"}, paths)
}

func TestConfig_ScanTags(t *testing.T) {
	type TestConfig struct {
		Tag struct {
			Data string
		} `yaml:"testTag"`
	}
	if err := LoadSource(file.NewSource("./testdata/config.yaml", false)); err != nil {
		t.Fatal(err)
	}
	cfg := TestConfig{}
	assert.NoError(t, Scan("", &cfg))
	assert.Equal(t, "test", cfg.Tag.Data)
}

func TestConfig_EnvFieldCate(t *testing.T) {
	_ = os.Setenv("DATABASE_PASSWORD", "password")
	_ = os.Setenv("DATABASE_DATASOURCE", "user:password@tcp(localhost:port)/db?charset=utf8mb4&parseTime=True&loc=Local")
	_ = os.Setenv("DATABASE_PORT", "3306")
	type Database struct {
		Password string
		Port     int
		database string
	}
	if err := LoadSource(env.NewSource(nil, nil)); err != nil {
		t.Fatal(err)
	}
	cfg := Database{}
	assert.NoError(t, Scan("database", &cfg))
	assert.Equal(t, 3306, cfg.Port)
}
