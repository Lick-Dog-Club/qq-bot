package weibo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/util/text2png"

	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("weibo", "获取今日实时的微博热搜榜单", func(bot bot.Bot, s string) error {
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
	}, features.WithAIFunc(features.AIFuncDef{
		Call: func(args string) (string, error) {
			return Top(), nil
		},
	}))
}

func Top() string {
	get, _ := http.Get("https://apis.tianapi.com/weibohot/index?key=" + config.TianApiKey())
	defer get.Body.Close()
	var data response
	json.NewDecoder(get.Body).Decode(&data)
	var res string
	for idx, datum := range data.Result.List {
		res += fmt.Sprintf("%d. %s\n", idx+1, datum.Hotword)
	}
	log.Printf("微博: %d\n", get.StatusCode)
	return res
}

//type response struct {
//	Success bool   `json:"success"`
//	Time    string `json:"time"`
//	Data    []struct {
//		Title string `json:"title"`
//		URL   string `json:"url"`
//		Hot   string `json:"hot"`
//	} `json:"data"`
//}

type response struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Result struct {
		List []struct {
			Hotword    string `json:"hotword"`
			Hotwordnum string `json:"hotwordnum"`
			Hottag     string `json:"hottag"`
		} `json:"list"`
	} `json:"result"`
}
