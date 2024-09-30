package translate

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/forPelevin/gomoji"
	log "github.com/sirupsen/logrus"
)

type TResponse struct {
	TSpeakUrl     string   `json:"tSpeakUrl"`
	RequestId     string   `json:"requestId"`
	Query         string   `json:"query"`
	Translation   []string `json:"translation"`
	MTerminalDict struct {
		Url string `json:"url"`
	} `json:"mTerminalDict"`
	ErrorCode string `json:"errorCode"`
	Dict      struct {
		Url string `json:"url"`
	} `json:"dict"`
	Webdict struct {
		Url string `json:"url"`
	} `json:"webdict"`
	L        string `json:"l"`
	IsWord   bool   `json:"isWord"`
	SpeakUrl string `json:"speakUrl"`
}

func EnToZh(appKey, appSecret, text string) string {
	// 添加请求参数
	paramsMap := map[string][]string{
		"q":    {gomoji.RemoveEmojis(text)},
		"from": {"auto"},
		"to":   {"zh-CHS"},
	}
	header := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}
	// 添加鉴权相关参数
	AddAuthParams(appKey, appSecret, paramsMap)
	// 请求api服务
	result := DoPost("https://openapi.youdao.com/api", header, paramsMap, "application/json")
	// 打印返回结果
	if result != nil {
		var data TResponse
		if err := json.Unmarshal(result, &data); err != nil {
			log.Println(err, string(result))
			return ""
		}
		if data.ErrorCode == "0" {
			return strings.Join(data.Translation, ",")
		}
	}
	return ""
}

func AddAuthParams(appKey string, appSecret string, params map[string][]string) {
	qs := params["q"]
	if qs == nil {
		qs = params["img"]
	}
	var q string
	for i := range qs {
		q += qs[i]
	}
	salt := getUuid()
	curtime := strconv.FormatInt(time.Now().Unix(), 10)
	sign := CalculateSign(appKey, appSecret, q, salt, curtime)
	params["appKey"] = []string{appKey}
	params["salt"] = []string{salt}
	params["curtime"] = []string{curtime}
	params["signType"] = []string{"v3"}
	params["sign"] = []string{sign}
}

func getUuid() string {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return ""
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func DoPost(url string, header map[string][]string, bodyMap map[string][]string, expectContentType string) []byte {
	client := &http.Client{
		Timeout: time.Second * 3,
	}
	params := neturl.Values{}
	for k, v := range bodyMap {
		for pv := range v {
			params.Add(k, v[pv])
		}
	}
	req, _ := http.NewRequest("POST", url, strings.NewReader(params.Encode()))
	for k, v := range header {
		for hv := range v {
			req.Header.Add(k, v[hv])
		}
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Print("request failed:", err)
		return nil
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, expectContentType) {
		print(string(body))
		return nil
	}
	return body
}

func CalculateSign(appKey string, appSecret string, q string, salt string, curtime string) string {
	strSrc := appKey + getInput(q) + salt + curtime + appSecret
	return encrypt(strSrc)
}

func encrypt(strSrc string) string {
	bt := []byte(strSrc)
	bts := sha256.Sum256(bt)
	return hex.EncodeToString(bts[:])
}

func getInput(q string) string {
	str := []rune(q)
	strLen := len(str)
	if strLen <= 20 {
		return q
	} else {
		return string(str[:10]) + strconv.Itoa(strLen) + string(str[strLen-10:])
	}
}
