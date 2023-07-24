package weibo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"qq/bot"
	"qq/features"
	"qq/util/text2png"

	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("微博", "获取热搜 top50", func(bot bot.Bot, s string) error {
		p := filepath.Join("/data", "images", "weibo50.png")
		text2png.Draw([]string{Top()}, p)
		if bot.Message().WeSendImg != nil {
			open, _ := os.Open(p)
			defer open.Close()
			bot.Message().WeSendImg(open)
		} else {
			bot.Send(fmt.Sprintf("[CQ:image,file=file://%s]", p))
		}
		return nil
	})
}

func Top() string {
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
