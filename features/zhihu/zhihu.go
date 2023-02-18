package zhihu

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qq/bot"
	"qq/features"

	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("知乎", "获取热搜 top30", func(bot bot.Bot, s string) error {
		bot.Send(top())
		return nil
	})
}

func top() string {
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
	// 消息太长发不出去, 取前 30 个
	data.Data = data.Data[:30]
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
