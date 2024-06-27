package jin10

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"qq/bot"
	"qq/features"
	"qq/util"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func init() {
	features.AddKeyword("jin10", "金十数据今日大事件", func(bot bot.Bot, content string) error {
		bot.Send(Get(util.Today()))
		return nil
	}, features.WithHidden())
}

type EventItem struct {
	Actual      *string     `json:"actual"`
	Affect      int         `json:"affect"`
	ShowAffect  int         `json:"show_affect"`
	Consensus   *string     `json:"consensus"`
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
{{if .Today}}日期: {{.Today}}{{end}}
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

	var (
		r   string
		rOk bool
	)
	if i.Consensus == nil {
		rOk = false
	} else {
		rOk = true
		r = *i.Consensus
	}

	t := i.Affect
	n, _ := strconv.ParseFloat(*i.Actual, 64)
	o, _ := strconv.ParseFloat(i.Previous, 64)
	c := i.Star
	l := ""
	B := 5

	var f float64
	if rOk {
		f, _ = strconv.ParseFloat(r, 64)
	} else {
		f = o
	}

	if n != 0 {
		if (f == 0 && rOk) || n == f {
			l = "影响较小"
			if c >= 3 {
				B = 6
			} else {
				B = 5
			}
		} else if t == 0 {
			if n > f {
				l = "利多"
				if c >= 3 {
					B = 3
				} else {
					B = 1
				}
			} else {
				l = "利空"
				if c >= 3 {
					B = 4
				} else {
					B = 2
				}
			}
		} else {
			if n > f {
				l = "利空"
				if c >= 3 {
					B = 4
				} else {
					B = 2
				}
			} else {
				l = "利多"
				if c >= 3 {
					B = 3
				} else {
					B = 1
				}
			}
		}
	} else {
		l = "未公布"
		B = 0
	}

	//fmt.Printf("影响描述: %s\n", l)
	_ = B
	//fmt.Printf("影响等级: %d\n", B)
	return l
}

func (i *EventItem) Render() string {
	bf := bytes.Buffer{}
	eventTemp.Execute(&bf, map[string]any{
		"Events": Events{*i},
	})
	return bf.String()
}

func Get(day time.Time) string {
	bf := bytes.Buffer{}
	eventTemp.Execute(&bf, map[string]any{
		"Events": BigEvents(day),
		"Today":  day.Format("2006-01-02"),
	})
	return bf.String()
}

func BigEvents(day time.Time) Events {
	url := fmt.Sprintf("https://cdn-rili.jin10.com/web_data/%d/daily/%02d/%02d/economics.json?t=%d", day.Year(), day.Month(), day.Day(), time.Now().UnixMilli())
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
