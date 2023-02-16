package lifetip

import (
	"encoding/json"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"
)

func init() {
	features.AddKeyword("rkl", "绕口令", func(bot bot.Bot, content string) error {
		bot.Send(RaoKouLing())
		return nil
	})
}

func RaoKouLing() string {
	get, _ := http.Get("https://apis.tianapi.com/rkl/index?key=" + config.TianApiKey)
	defer get.Body.Close()
	var data response
	json.NewDecoder(get.Body).Decode(&data)
	if data.Code == 200 {
		return data.Result.Content
	}
	return ""
}

type response struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Result struct {
		Content string `json:"content"`
	} `json:"result"`
}
