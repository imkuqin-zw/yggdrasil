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

package flag

import (
	"encoding/json"
	flag2 "flag"
	"testing"
)

var (
	dbuser = flag2.String("database-user", "default", "db user")
	dbhost = flag2.String("database-host", "", "db host")
	dbpw   = flag2.String("database-password", "", "db pw")
)

func initTestFlags() {
	flag2.Set("database-host", "localhost")
	flag2.Set("database-password", "some-password")
	flag2.Parse()
}

func TestFlagsrc_ReadAll(t *testing.T) {
	initTestFlags()
	source := NewSource()
	c, err := source.Read()
	if err != nil {
		t.Error(err)
	}

	var actual map[string]interface{}
	if err := json.Unmarshal(c.Data(), &actual); err != nil {
		t.Error(err)
	}
	actualDB := actual["database"].(map[string]interface{})

	// unset flag defaults should be loaded
	if actualDB["user"] != *dbuser {
		t.Errorf("expected %v got %v", *dbuser, actualDB["user"])
	}
}
