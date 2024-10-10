package star

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/util"
	"strings"
	"time"

	"github.com/golang-module/carbon/v2"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func init() {
	features.AddKeyword("star", "<+date: 2000-01-01> 根据日期获取对应的星座", func(bot bot.Bot, content string) error {
		bot.Send(GetStar(content))
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"date": {
				Type:        jsonschema.String,
				Description: "日期，格式为 2006-01-02",
			},
		},
		Call: func(args string) (string, error) {
			var s = struct {
				Date string `json:"date"`
			}{}
			json.Unmarshal([]byte(args), &s)
			return GetStar(s.Date), nil
		},
	}), features.WithGroup("star"))
	features.AddKeyword("starx", "<+date: 2000-01-01> 根据日期获取对应的星座运势", func(bot bot.Bot, content string) error {
		bot.Send(Get(content))
		return nil
	}, features.WithGroup("star"))
}

func Get(day string) string {
	get, _ := http.Get(fmt.Sprintf("https://apis.tianapi.com/star/index?key=%s&astro=%s", config.TianApiKey(), GetStar(day)))
	defer get.Body.Close()
	var data response
	json.NewDecoder(get.Body).Decode(&data)
	m := make(map[string]string)
	for _, s := range data.Result.List {
		m[s.Type] = s.Content
	}
	return m["今日概述"]
}

func GetStar(day string) string {
	for star, date := range startMap {
		if date.Between(carbon.Parse(day).StdTime().Local()) {
			return star
		}
	}
	return "未知星座"
}

type response struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Result struct {
		List []struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		} `json:"list"`
	} `json:"result"`
}

type DayBetween struct {
	Start string
	End   string
}

func (b DayBetween) StartMonth() int64 {
	return util.ToInt64(strings.Split(b.Start, "-")[0])
}
func (b DayBetween) StartDay() int64 {
	return util.ToInt64(strings.Split(b.Start, "-")[1])
}
func (b DayBetween) EndMonth() int64 {
	return util.ToInt64(strings.Split(b.End, "-")[0])
}
func (b DayBetween) EndDay() int64 {
	return util.ToInt64(strings.Split(b.End, "-")[1])
}

func (b DayBetween) Between(parse time.Time) bool {
	if b.StartMonth() > b.EndMonth() {
		if (b.StartMonth() <= int64(parse.Month()) && int64(parse.Month()) <= 12) || (1 <= int64(parse.Month()) && int64(parse.Month()) <= b.EndMonth()) {
			if b.StartMonth() == int64(parse.Month()) {
				return b.StartDay() <= int64(parse.Day())
			}
			if b.EndMonth() == int64(parse.Month()) {
				return int64(parse.Day()) <= b.EndDay()
			}
			return true
		}
		return false
	}
	if b.StartMonth() <= int64(parse.Month()) && int64(parse.Month()) <= b.EndMonth() {
		if b.StartMonth() == int64(parse.Month()) {
			return b.StartDay() <= int64(parse.Day())
		}
		if b.EndMonth() == int64(parse.Month()) {
			return int64(parse.Day()) <= b.EndDay()
		}
		return true
	}
	return false
}

var startMap = map[string]DayBetween{
	"白羊座": {"3-21", "4-19"},
	"金牛座": {"4-20", "5-20"},
	"双子座": {"5-21", "6-21"},
	"巨蟹座": {"6-22", "7-22"},
	"狮子座": {"7-23", "8-22"},
	"处女座": {"8-23", "9-22"},
	"天秤座": {"9-23", "10-23"},
	"天蝎座": {"10-24", "11-22"},
	"射手座": {"11-23", "12-21"},
	"摩羯座": {"12-22", "1-19"},
	"水瓶座": {"1-20", "2-18"},
	"双鱼座": {"2-19", "3-20"},
}
