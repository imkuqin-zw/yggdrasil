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

package env

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestEnv_Read(t *testing.T) {
	expected := map[string]map[string]string{
		"database": {
			"host":       "localhost",
			"password":   "password",
			"datasource": "user:password@tcp(localhost:port)/db?charset=utf8mb4&parseTime=True&loc=Local",
		},
	}

	os.Setenv("DATABASE_HOST", "localhost")
	os.Setenv("DATABASE_PASSWORD", "password")
	os.Setenv("DATABASE_DATASOURCE", "user:password@tcp(localhost:port)/db?charset=utf8mb4&parseTime=True&loc=Local")

	source := NewSource(nil, nil)
	c, err := source.Read()
	if err != nil {
		t.Error(err)
	}

	var actual map[string]interface{}
	if err := json.Unmarshal(c.Data(), &actual); err != nil {
		t.Error(err)
	}

	actualDB := actual["database"].(map[string]interface{})

	for k, v := range expected["database"] {
		a := actualDB[k]

		if a != v {
			t.Errorf("expected %v got %v", v, a)
		}
	}
}

func TestEnvvar_Prefixes(t *testing.T) {
	os.Setenv("APP_DATABASE_HOST", "localhost")
	os.Setenv("APP_DATABASE_PASSWORD", "password")
	os.Setenv("VAULT_ADDR", "vault:1337")
	os.Setenv("MICRO_REGISTRY", "mdns")

	var prefixtests = []struct {
		prefixOpts   [][]string
		expectedKeys []string
	}{
		{[][]string{{"APP", "MICRO"}, nil}, []string{"app", "micro"}},
		{[][]string{{"MICRO"}, {"APP"}}, []string{"database", "micro"}},
	}

	for _, pt := range prefixtests {
		source := NewSource(pt.prefixOpts[0], pt.prefixOpts[1])

		c, err := source.Read()
		if err != nil {
			t.Error(err)
		}

		var actual map[string]interface{}
		if err := json.Unmarshal(c.Data(), &actual); err != nil {
			t.Error(err)
		}
		// assert other prefixes ignored
		if l := len(actual); l != len(pt.expectedKeys) {
			t.Errorf("expected %v top keys, got %v", len(pt.expectedKeys), l)
		}

		for _, k := range pt.expectedKeys {
			if !containsKey(actual, k) {
				t.Errorf("expected key %v, not found", k)
			}
		}
	}
}

func TestEnvvar_WatchNextNoOpsUntilStop(t *testing.T) {
	src := NewSource(nil, []string{"GOMICRO_"})
	w, err := src.Watch()
	if err != nil {
		t.Error(err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = src.Close()
	}()

	if w != nil || err != nil {
		t.Errorf("expected watcher stopped error, got %v", err)
	}
}

func containsKey(m map[string]interface{}, s string) bool {
	for k := range m {
		if k == s {
			return true
		}
	}
	return false
}
