package weather

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"

	"github.com/sashabaranov/go-openai/jsonschema"
)

const weatherURL = "https://restapi.amap.com/v3/weather/weatherInfo?key=%s&city=%s&extensions=all&output=json"

var temp, _ = template.New("").Parse(`城市: {{.City}}

白天: {{.Dayweather}} {{.Daytemp}} ℃
夜晚: {{.Nightweather}} {{.Nighttemp}} ℃`)

type input struct {
	City         string
	Date         string
	Dayweather   string
	Daytemp      string
	Nightweather string
	Nighttemp    string
}

func init() {
	features.AddKeyword("weather", "<+城市> 获取天气信息, 默认杭州", func(bot bot.Bot, city string) error {
		if city == "" {
			city = "杭州"
		}
		bot.Send(Get(city))
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"city": {
				Type:        jsonschema.String,
				Description: "The city and state, e.g. 天津, 北京",
			},
		},
		Call: func(args string) (string, error) {
			var city = struct {
				City string `json:"city"`
			}{}
			json.Unmarshal([]byte(args), &city)
			return Get(city.City), nil
		},
	}))
}

func Get(city string) string {
	if config.WeatherKey() == "" {
		return "请先设置环境变量: WEATHER_KEY "
	}
	resp, _ := http.Get(fmt.Sprintf(weatherURL, config.WeatherKey(), city))
	defer closeBody(resp.Body)
	var res response
	json.NewDecoder(resp.Body).Decode(&res)
	if len(res.Forecasts) < 1 {
		return "未找到天气信息"
	}
	bf := &bytes.Buffer{}
	temp.Execute(bf, input{
		City:         res.Forecasts[0].City,
		Date:         res.Forecasts[0].Casts[0].Date,
		Dayweather:   res.Forecasts[0].Casts[0].Dayweather,
		Daytemp:      res.Forecasts[0].Casts[0].Daytemp,
		Nightweather: res.Forecasts[0].Casts[0].Nightweather,
		Nighttemp:    res.Forecasts[0].Casts[0].Nighttemp,
	})
	return bf.String()
}

func closeBody(rc io.ReadCloser) {
	io.Copy(io.Discard, rc)
	rc.Close()
}

type response struct {
	Status    string `json:"status"`
	Count     string `json:"count"`
	Info      string `json:"info"`
	Infocode  string `json:"infocode"`
	Forecasts []struct {
		City       string `json:"city"`
		Adcode     string `json:"adcode"`
		Province   string `json:"province"`
		Reporttime string `json:"reporttime"`
		Casts      []struct {
			Date         string `json:"date"`
			Week         string `json:"week"`
			Dayweather   string `json:"dayweather"`
			Nightweather string `json:"nightweather"`
			Daytemp      string `json:"daytemp"`
			Nighttemp    string `json:"nighttemp"`
			Daywind      string `json:"daywind"`
			Nightwind    string `json:"nightwind"`
			Daypower     string `json:"daypower"`
			Nightpower   string `json:"nightpower"`
		} `json:"casts"`
	} `json:"forecasts"`
}
