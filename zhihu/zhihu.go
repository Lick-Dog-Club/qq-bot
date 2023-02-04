package zhihu

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Top50() string {
	request, _ := http.NewRequest("GET", "https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=50", nil)
	request.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	request.Header.Add("referer", "https://www.zhihu.com/hot")
	get, _ := http.DefaultClient.Do(request)
	defer get.Body.Close()
	var data Response
	json.NewDecoder(get.Body).Decode(&data)
	var res string
	for idx, datum := range data.Data {
		res += fmt.Sprintf("%d. %s\n", idx+1, datum.Target.Title)
	}
	return res
}

type Response struct {
	Data []struct {
		Type   string `json:"type"`
		Target struct {
			Title string `json:"title"`
			Url   string `json:"url"`
		} `json:"target"`
	} `json:"data"`
}
