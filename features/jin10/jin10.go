package jin10

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"qq/bot"
	"qq/features"
	"qq/util"
	"strings"
	"text/template"
	"time"
)

func init() {
	features.AddKeyword("jin10", "金十数据今日大事件", func(bot bot.Bot, content string) error {
		bot.Send(Get(util.Today()))
		return nil
	})
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
	return time.Now().Sub(i.PubTime) > 0 && time.Now().Sub(i.PubTime) <= t
}

type Events []EventItem

var eventTemp, _ = template.New("").Funcs(map[string]any{
	"star": func(n int) string {
		return strings.Repeat("⭐️", n)
	},
	"affect": func(i *EventItem) string {
		return i.AffectStr()
	},
	"date": func(t time.Time) string {
		return t.Local().Format("2006-01-02 15:04")
	},
}).Parse(`
日期: {{.Today}}
{{ range .Events }}
{{.Country}}{{.Name}}
重要程度: {{ .Star | star }}
公布时间: {{.PubTime | date}}
前值: {{ if .Revised }}{{.Revised}}{{else}}{{.Previous}}{{end}}
预测值: {{.Consensus}}
公布值: {{ if .Actual}}{{.Actual}}{{else}}未公布{{end}}
影响: {{ . | affect }}
{{end}}
`)

func (i *EventItem) AffectStr() string {
	if i.Actual == nil {
		return "未公布"
	}

	if util.ToFloat64(*i.Actual) > util.ToFloat64(i.Consensus) {
		return "利空"
	}

	return "利多"
}

func Get(day time.Time) string {
	importantData := ImportantItems(day)
	bf := bytes.Buffer{}
	eventTemp.Execute(&bf, map[string]any{
		"Events": importantData,
		"Today":  day.Format("2006-01-02"),
	})
	return bf.String()
}

func ImportantItems(day time.Time) Events {
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
