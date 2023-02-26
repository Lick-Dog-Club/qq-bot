package proxy

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"qq/config"
	"time"
)

func proxyFunc(r *http.Request) (*url.URL, error) {
	parse, _ := url.Parse(config.HttpProxy())
	if parse != nil && parse.Host != "" {
		return parse, nil
	}
	environment, _ := http.ProxyFromEnvironment(r)
	if environment != nil && environment.Host != "" {
		return environment, nil
	}
	return nil, errors.New("proxy not found")
}

func NewHttpProxyClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Minute,
		Transport: &http.Transport{
			Proxy: proxyFunc,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			MaxConnsPerHost: 1000,
		},
	}
}
