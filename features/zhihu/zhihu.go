package zhihu

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
	features.AddKeyword("zhihu", "获取知乎热搜榜单", func(bot bot.Bot, s string) error {
		p := filepath.Join(config.ImageDir, "zhihu50.png")
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
	request, _ := http.NewRequest("GET", "https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=50", nil)
	request.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	request.Header.Add("referer", "https://www.zhihu.com/hot")
	get, err := http.DefaultClient.Do(request)
	if err != nil {
		return err.Error()
	}
	defer get.Body.Close()
	var data response
	json.NewDecoder(get.Body).Decode(&data)
	var res string
	for idx, datum := range data.Data {
		res += fmt.Sprintf("%d. %s\n", idx+1, datum.Target.Title)
	}
	log.Printf("知乎 top50: %d\n", get.StatusCode)
	return res
}

type response struct {
	Data []struct {
		Type   string `json:"type"`
		Target struct {
			Title string `json:"title"`
			Url   string `json:"url"`
		} `json:"target"`
	} `json:"data"`
}
