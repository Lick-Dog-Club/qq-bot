package zaoan

import (
	"encoding/json"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
)

type response struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Result struct {
		Content string `json:"content"`
	} `json:"result"`
}

func get() string {
	resp, _ := http.Get("https://apis.tianapi.com/zaoan/index?key=" + config.TianApiKey())
	defer resp.Body.Close()
	var data response
	json.NewDecoder(resp.Body).Decode(&data)
	return data.Result.Content
}

func init() {
	cronjob.NewCommand("zaoan", func(robot bot.CronBot) error {
		robot.SendToUser(config.UserID(), get())
		return nil
	}).DailyAt("9")
}
