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

package xarray

func ReverseStringArray(ss []string) {
	for i := len(ss)/2 - 1; i >= 0; i-- {
		opp := len(ss) - 1 - i
		ss[i], ss[opp] = ss[opp], ss[i]
	}
}

func ReverseArray[T any](ss []T) {
	for i := len(ss)/2 - 1; i >= 0; i-- {
		opp := len(ss) - 1 - i
		ss[i], ss[opp] = ss[opp], ss[i]
	}
}

func RemoveReplaceStrings(arr []string) []string {
	set := make(map[string]struct{})
	j := 0
	for _, item := range arr {
		if _, ok := set[item]; ok {
			continue
		}
		arr[j] = item
		set[item] = struct{}{}
		j++
	}
	return arr[:j]
}

func RemoveReplace[T any](arr []T) []T {
	set := make(map[any]struct{})
	j := 0
	for _, item := range arr {
		if _, ok := set[item]; ok {
			continue
		}
		arr[j] = item
		set[item] = struct{}{}
		j++
	}
	return arr[:j]
}

func RemoveEmptyStrings(arr []string) []string {
	j := 0
	for _, item := range arr {
		if item == "" {
			continue
		}
		arr[j] = item
		j++
	}
	return arr[:j]
}
