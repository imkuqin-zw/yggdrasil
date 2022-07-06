package xnet

import "strings"

func GetQuery(url string) string {
	i := strings.Index(url, "?")
	if i < 0 {
		return ""
	}
	return url[i:]
}
