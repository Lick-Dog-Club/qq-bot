package proxy

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"qq/config"
	"time"
)

var proxyClient = &http.Client{
	Timeout: 5 * time.Minute,
	Transport: &http.Transport{
		Proxy: proxyFunc,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxConnsPerHost: 1000,
	},
}

func proxyFunc(r *http.Request) (*url.URL, error) {
	parse, _ := url.Parse(config.HttpProxy())
	if parse != nil && parse.Host != "" {
		return parse, nil
	}
	environment, _ := http.ProxyFromEnvironment(r)
	if environment != nil && environment.Host != "" {
		return environment, nil
	}
	return nil, nil
}

//func localProxyFunc(r *http.Request) (*url.URL, error) {
//	parse, _ := url.Parse("http://localhost:7890")
//	if parse != nil && parse.Host != "" {
//		return parse, nil
//	}
//	environment, _ := http.ProxyFromEnvironment(r)
//	if environment != nil && environment.Host != "" {
//		return environment, nil
//	}
//	return nil, nil
//}

func NewHttpProxyClient() *http.Client {
	return proxyClient
}
