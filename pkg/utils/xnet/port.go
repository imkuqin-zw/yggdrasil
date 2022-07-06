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
