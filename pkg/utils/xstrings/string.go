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

package xstrings

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
	"unsafe"
)

var (
	rxCamel = regexp.MustCompile(`[\p{L}\p{N}]+`)
)

// StrInSlice convert string to bool
func StrInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Str2bytes convert string to array of byte
func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// Bytes2str convert array of byte to string
func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// SplitToTwo split the string
func SplitToTwo(s, sep string) (string, string) {
	index := strings.Index(s, sep)
	if index < 0 {
		return "", s
	}
	return s[:index], s[index+len(sep):]
}

// SplitFirstSep split the string
func SplitFirstSep(s, sep string) string {
	index := strings.Index(s, sep)
	if index < 0 {
		return ""
	}
	return s[:index]
}

// MinInt check the minimum value of two integers
func MinInt(x, y int) int {
	if x <= y {
		return x
	}

	return y
}

// ClearStringMemory clear string memory, for very sensitive security related data
// //you should clear it in memory after use
func ClearStringMemory(src *string) {
	p := (*struct {
		ptr uintptr
		len int
	})(unsafe.Pointer(src))

	len := MinInt(p.len, 32)
	ptr := p.ptr
	for idx := 0; idx < len; idx = idx + 1 {
		b := (*byte)(unsafe.Pointer(&ptr))
		*b = 0
		ptr++
	}
}

// ClearByteMemory clear byte memory, for very sensitive security related data
// you should clear it in memory after use
func ClearByteMemory(src []byte) {
	len := MinInt(len(src), 32)
	for idx := 0; idx < len; idx = idx + 1 {
		src[idx] = 0
	}
}

// StrToCamelCase converts from underscore separated form to camel case form.
func StrToCamelCase(s string) string {
	byteSrc := []byte(s)

	chunks := rxCamel.FindAll(byteSrc, -1)
	for idx, val := range chunks {
		//chunks[idx] = []byte(cases.Title(language.Make(string(val))).String(string(val)))
		chunks[idx] = bytes.Title(val)
	}
	return Bytes2str(bytes.Join(chunks, nil))
}

// StrToSnakeCase converts from camel case form to underscore separated form.
func StrToSnakeCase(s string) string {
	s = StrToCamelCase(s)
	runes := []rune(s)
	length := len(runes)
	var out []rune
	for i := 0; i < length; i++ {
		out = append(out, unicode.ToLower(runes[i]))
		if i+1 < length && (unicode.IsUpper(runes[i+1]) && unicode.IsLower(runes[i])) {
			out = append(out, '_')
		}
	}

	return string(out)
}

// ToLowerFirstCamelCase returns the given string in camelcase formatted string
// but with the first letter being lowercase.
func StrToLowerFirstCamelCase(s string) string {
	if s == "" {
		return s
	}
	if len(s) == 1 {
		return strings.ToLower(string(s[0]))
	}
	s = StrToCamelCase(s)
	return strings.ToLower(string(s[0])) + s[1:]
}

// StrToUpperFirst returns the given string with the first letter being uppercase.
func StrToUpperFirst(s string) string {
	if s == "" {
		return s
	}
	if len(s) == 1 {
		return strings.ToLower(string(s[0]))
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

// StrToLowerSnakeCase the given string in snake-case format.
func StrToLowerSnakeCase(s string) string {
	return strings.ToLower(StrToSnakeCase(s))
}

func HasPrefix(key, pre, delimiter string) bool {
	if !strings.HasPrefix(key, pre) {
		return false
	}
	if len(key) != len(pre) && !strings.HasPrefix(key[len(pre):], delimiter) {
		return false
	}
	return true
}

// FullMatchWithRegex returns whether the full string matches the regex provided.
func FullMatchWithRegex(re *regexp.Regexp, string string) bool {
	re.Longest()
	rem := re.FindString(string)
	return len(rem) == len(string)
}
