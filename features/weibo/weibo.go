package weibo

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qq/bot"
	"qq/features"
)

func init() {
	features.AddKeyword("微博", "获取热搜 top50", func(bot bot.Bot, s string) error {
		bot.Send(top())
		return nil
	})
}

func top() string {
	get, _ := http.Get("https://api.vvhan.com/api/wbhot")
	defer get.Body.Close()
	var data response
	json.NewDecoder(get.Body).Decode(&data)
	var res string
	for idx, datum := range data.Data {
		res += fmt.Sprintf("%d. %s\n", idx+1, datum.Title)
	}
	log.Printf("微博: %d\n", get.StatusCode)
	return res
}

type response struct {
	Success bool   `json:"success"`
	Time    string `json:"time"`
	Data    []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Hot   string `json:"hot"`
	} `json:"data"`
}
