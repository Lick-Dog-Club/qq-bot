package weibo

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func Top() string {
	get, _ := http.Get("https://api.vvhan.com/api/wbhot")
	defer get.Body.Close()
	var data Response
	json.NewDecoder(get.Body).Decode(&data)
	var res string
	for idx, datum := range data.Data {
		res += fmt.Sprintf("%d. %s\n", idx+1, datum.Title)
	}
	log.Printf("微博: %d\n", get.StatusCode)
	return res
}

type Response struct {
	Success bool   `json:"success"`
	Time    string `json:"time"`
	Data    []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Hot   string `json:"hot"`
	} `json:"data"`
}
