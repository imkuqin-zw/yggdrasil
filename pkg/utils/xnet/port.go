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

package xnet

import (
	"net"
	"strings"
)

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func HasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

// removeEmptyPort strips the empty port in ":port" to ""
// as mandated by RFC 3986 Section 6.2.3.
func RemoveEmptyPort(host string) string {
	if HasPort(host) {
		return strings.TrimSuffix(host, ":")
	}
	return host
}

func GetPort() int {
	l, _ := net.Listen("tcp", ":0") // listen on localhost
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port

	return port
}

func GetAddrAndPort() (string, int) {
	l, _ := net.Listen("tcp", ":0") // listen on localhost
	defer l.Close()
	addr := l.Addr().(*net.TCPAddr).IP.String()
	port := l.Addr().(*net.TCPAddr).Port

	return addr, port
}
