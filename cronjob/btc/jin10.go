package btc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/util"
	"strings"
	"time"
)

func init() {
	cronjob.Manager().NewCommand("jin10", func(bot bot.CronBot) error {
		bot.SendToUser(config.UserID(), "明日大事件播报")
		bot.SendToUser(config.UserID(), Get(util.Today().Add(24*time.Hour)))
		return nil
	}).DailyAt("09:30")
	cronjob.Manager().NewCommand("jin10-watch", func(bot bot.CronBot) error {
		for _, item := range importantItems(time.Now()) {
			if item.IsRecentlyPub(time.Second*10) && item.Actual != nil {
				util.Bark(
					affect(item),
					fmt.Sprintf("%s%s, 预测值: %s, 公布值: %s", item.Country, item.Name, item.Consensus, *item.Actual),
					config.BarkUrls()...,
				)
			}
		}
		return nil
	}).EveryTenSeconds()
}

type EventItem struct {
	Actual      *string     `json:"actual"`
	Affect      int         `json:"affect"`
	ShowAffect  int         `json:"show_affect"`
	Consensus   string      `json:"consensus"`
	Country     string      `json:"country"`
	ID          int         `json:"id"`
	IndicatorID int         `json:"indicator_id"`
	Name        string      `json:"name"`
	Previous    string      `json:"previous"`
	PubTime     time.Time   `json:"pub_time"`
	Revised     *string     `json:"revised"`
	Star        int         `json:"star"`
	TimePeriod  string      `json:"time_period"`
	Unit        string      `json:"unit"`
	VideoURL    interface{} `json:"video_url"`
	VipResource interface{} `json:"vip_resource"`
	PubTimeUnix int         `json:"pub_time_unix"`
	TimeStatus  interface{} `json:"time_status"`
}

func (i *EventItem) IsRecentlyPub(t time.Duration) bool {
	if time.Now().Sub(i.PubTime) > 0 && time.Now().Sub(i.PubTime) <= t {
		return true
	}

	return false
}

type Events []EventItem

var eventTemp, _ = template.New("").Funcs(map[string]any{
	"star": func(n int) string {
		return strings.Repeat("⭐️", n)
	},
	"affect": affect,
}).Parse(`
日期: {{.Today}}
{{ range .Events }}
{{.Country}}{{.Name}}
重要程度: {{ .Star | star }}
前值: {{ if .Revised }}{{.Revised}}{{else}}{{.Previous}}{{end}}
预测值: {{.Consensus}}
公布值: {{ if .Actual}}{{.Actual}}{{else}}未公布{{end}}
影响: {{ . | affect }}
{{end}}
`)

func affect(item EventItem) string {
	if item.Actual == nil {
		return "未公布"
	}

	if util.ToFloat64(*item.Actual) > util.ToFloat64(item.Consensus) {
		return "利空"
	}

	return "利多"
}

func Get(day time.Time) string {
	importantData := importantItems(day)
	bf := bytes.Buffer{}
	eventTemp.Execute(&bf, map[string]any{
		"Events": importantData,
		"Today":  day.Format("2006-01-02"),
	})
	return bf.String()
}

func importantItems(day time.Time) Events {
	url := fmt.Sprintf("https://cdn-rili.jin10.com/web_data/%d/daily/%02d/%d/economics.json", day.Year(), day.Month(), day.Day())
	fmt.Println(url)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	var data Events
	json.NewDecoder(resp.Body).Decode(&data)
	var importantData Events
	for _, datum := range data {
		if datum.Star >= 3 && datum.Country == "美国" {
			importantData = append(importantData, datum)
		}
	}
	return importantData
}
