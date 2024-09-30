package xiaofeiquan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	cronjob.NewCommand("hangzhou-xiaofeiquan", func(bot bot.CronBot) error {
		res := Fetch("杭州消费券")
		if len(res) > 0 {
			bot.SendGroup(config.GroupID(), res)
		}
		return nil
	}).DailyAt("9:40")
	cronjob.NewCommand("shaoxin-xiaofeiquan", func(bot bot.CronBot) error {
		res := Fetch("绍兴消费券")
		if len(res) > 0 {
			bot.SendToUser(config.UserID(), res)
		}
		return nil
	}).DailyAt("9:40")
}

func isToday(t time.Time) bool {
	format := "20060102"
	return t.Format(format) == time.Now().Format(format)
}

func Fetch(keyword string) string {
	get, err := http.Get(fmt.Sprintf("https://so.zjol.com.cn/s?wd=%s&chnl=0&app=site176_zjol&field=title", keyword))
	if err != nil {
		log.Println(err)
	}
	defer get.Body.Close()
	var data response
	json.NewDecoder(get.Body).Decode(&data)
	var res string
	for _, item := range data.Msg.Result.Items {
		if isToday(strToDate(item.Issuetime)) {
			res += fmt.Sprintf("%s\nhttps:%s\n", item.Title, item.Docpuburl)
		} else {
			log.Println("[SKIP]: ", item.Title)
		}
	}
	return res
}

func strToDate(s string) time.Time {
	atoi, _ := strconv.Atoi(s)
	unix := time.Unix(int64(atoi), 0)
	return unix
}

type response struct {
	Msg struct {
		Status    string `json:"status"`
		RequestID string `json:"request_id"`
		Result    struct {
			Searchtime  float64 `json:"searchtime"`
			Total       int     `json:"total"`
			Num         int     `json:"num"`
			Viewtotal   int     `json:"viewtotal"`
			ComputeCost []struct {
				IndexName string  `json:"index_name"`
				Value     float64 `json:"value"`
			} `json:"compute_cost"`
			Items []struct {
				Abstract     string `json:"abstract"`
				Content      string `json:"content"`
				Docpuburl    string `json:"docpuburl"`
				Metalogourl  string `json:"metalogourl"`
				Title        string `json:"title"`
				Issuetime    string `json:"issuetime"`
				Topchannelid string `json:"topchannelid"`
				IndexName    string `json:"index_name"`
			} `json:"items"`
			Facet []interface{} `json:"facet"`
		} `json:"result"`
		Errors         []interface{} `json:"errors"`
		Tracer         string        `json:"tracer"`
		OpsRequestMisc string        `json:"ops_request_misc"`
	} `json:"msg"`
}
