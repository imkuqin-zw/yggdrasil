package xnet

import (
	netHttp "net/http"
	"strings"
)

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func DelHopHeaders(header netHttp.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func CopyHeader(dst, src netHttp.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func SetCookie(w netHttp.ResponseWriter, src []*netHttp.Cookie) {
	for _, v := range src {
		netHttp.SetCookie(w, v)
	}
}

func AppendHostToXForwardHeader(header netHttp.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}
