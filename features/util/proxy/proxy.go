package proxy

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"qq/config"
	"time"
)

func NewHttpProxyClient() *http.Client {
	parse, _ := url.Parse(config.HttpProxy())
	return &http.Client{
		Timeout: 5 * time.Minute,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parse),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			MaxConnsPerHost: 1000,
		},
	}
}
