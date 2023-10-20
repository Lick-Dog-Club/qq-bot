package bitget

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type RestClient struct {
	ApiKey       string
	ApiSecretKey string
	Passphrase   string
	BaseUrl      string
	HttpClient   http.Client
	Signer       *Signer
}

func (p *RestClient) DoGet(uri string, params map[string]string) (string, error) {
	timesStamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	body := BuildGetParams(params)

	sign := p.Signer.Sign("GET", uri, body, timesStamp)

	requestUrl := p.BaseUrl + uri + body

	request, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return "", err
	}
	Headers(request, p.ApiKey, timesStamp, sign, p.Passphrase)

	response, err := p.HttpClient.Do(request)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	bodyStr, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	responseBodyString := string(bodyStr)
	return responseBodyString, err
}

func BuildGetParams(params map[string]string) string {
	urlParams := url.Values{}
	if params != nil && len(params) > 0 {
		for k := range params {
			urlParams.Add(k, params[k])
		}
	}
	return "?" + urlParams.Encode()
}

const (
	/*
	  http headers
	*/
	ContentType        = "Content-Type"
	BgAccessKey        = "ACCESS-KEY"
	BgAccessSign       = "ACCESS-SIGN"
	BgAccessTimestamp  = "ACCESS-TIMESTAMP"
	BgAccessPassphrase = "ACCESS-PASSPHRASE"
	ApplicationJson    = "application/json"
)

func Headers(request *http.Request, apikey string, timestamp string, sign string, passphrase string) {
	request.Header.Add(ContentType, ApplicationJson)
	request.Header.Add(BgAccessKey, apikey)
	request.Header.Add(BgAccessSign, sign)
	request.Header.Add(BgAccessTimestamp, timestamp)
	request.Header.Add(BgAccessPassphrase, passphrase)
}
