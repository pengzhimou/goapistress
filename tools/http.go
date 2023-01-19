package tools

import (
	"fmt"
	"net/url"
	"strings"
)

func UrlFull(scheme, host, path string, query url.Values) *url.URL {
	urlF := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
		// RawQuery: string(query.Encode()),
		RawQuery: query.Encode(),
	}
	return urlF
}

// GetHeaderValue 获取 header
func GetHeaderValue(v string, headers map[string]string) {
	index := strings.Index(v, ":")
	if index < 0 {
		return
	}
	vIndex := index + 1
	if len(v) >= vIndex {
		value := strings.TrimPrefix(v[vIndex:], " ")
		if _, ok := headers[v[:index]]; ok {
			headers[v[:index]] = fmt.Sprintf("%s; %s", headers[v[:index]], value)
		} else {
			headers[v[:index]] = value
		}
	}
}
