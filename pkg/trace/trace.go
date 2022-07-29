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

package trace

import "github.com/imkuqin-zw/yggdrasil/pkg/types"

var constructors = make(map[string]types.TracerProviderConstructor)

func RegisterConstructor(name string, constructor types.TracerProviderConstructor) {
	constructors[name] = constructor
}

func GetConstructor(name string) types.TracerProviderConstructor {
	constructor, _ := constructors[name]
	return constructor
}
