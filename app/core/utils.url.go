package core

import (
	"fmt"
	"net/url"
	"strings"
)

func ParseUrlHostName(urlStr string) (string, error) {
	urlStr = strings.Trim(urlStr, "\n \t")
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "www.")
	urlStr = strings.Trim(urlStr, "\n \t/")
	u, err := url.Parse(fmt.Sprintf("http://%s", urlStr))
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}
