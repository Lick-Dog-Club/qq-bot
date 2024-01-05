package smzdm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"qq/util/random"
)

type checkInReq struct {
	TouchstoneEvent string `json:"touchstone_event"`
	SK              string `json:"sk"`
	Token           string `json:"token"`
	Captcha         string `json:"captcha"`
}

func CheckIn(token, sk string) {
	marshal, _ := json.Marshal(&checkInReq{
		TouchstoneEvent: "",
		SK:              sk,
		Token:           token,
		Captcha:         "",
	})
	request, _ := http.NewRequest("POST", "https://user-api.smzdm.com/checkin", bytes.NewReader(marshal))
	request.Header.Add("Content-Type", "application/json")
	setHeaders(request, "")
}

const (
	SignKey       = `apr1$AwP!wRRT$gJ/q.X24poeBInlUJC`
	AppVersion    = "10.4.26"
	AppVersionRev = "866"
	UserAgentWeb  = `Mozilla/5.0 (Linux; Android 10.0; Redmi Build/Redmi Note 3; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/95.0.4638.74 Mobile Safari/537.36 smzdm_android_V` + AppVersion + ` rv:` + AppVersionRev + ` (Redmi;Android10.0;zh) jsbv_1.0.0 webv_2.0 smzdmapp`
)

func setHeaders(req *http.Request, cookie string) {
	ua := UserAgentWeb
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "zh-Hans-CN;q=1")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("request_key", random.String(16))
	req.Header.Add("User-Agent", ua)
	req.Header.Add("Cookie", cookie)
}
