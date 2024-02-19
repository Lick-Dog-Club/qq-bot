package lifetip

import (
	"encoding/json"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"
	"strings"
)

func init() {
	features.AddKeyword("rkl", "获取一个绕口令", func(bot bot.Bot, content string) error {
		bot.Send(RaoKouLing())
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: nil,
		Call: func(args string) (string, error) {
			return RaoKouLing(), nil
		},
	}))
}

func RaoKouLing() string {
	get, _ := http.Get("https://apis.tianapi.com/rkl/index?key=" + config.TianApiKey())
	defer get.Body.Close()
	var data response
	json.NewDecoder(get.Body).Decode(&data)
	if data.Code == 200 {
		return strings.ReplaceAll(data.Result.List[0].Content, `<br/>`, "")
	}
	return ""
}

type response struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Result struct {
		List []struct {
			Content string `json:"content"`
		} `json:"list"`
	} `json:"result"`
}
