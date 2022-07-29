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

package log

import (
	"log"

	"github.com/imkuqin-zw/yggdrasil/pkg/types"
)

var loggerConstructors = make(map[string]types.LoggerConstructor)

func RegisterConstructor(name string, f types.LoggerConstructor) {
	loggerConstructors[name] = f
}

func GetConstructor(name string) types.LoggerConstructor {
	f, _ := loggerConstructors[name]
	return f
}

func GetLogger(name string) types.Logger {
	f := GetConstructor(name)
	if f == nil {
		log.Fatalf("unknown logger constructor, name: %s", name)
	}
	return f()
}
