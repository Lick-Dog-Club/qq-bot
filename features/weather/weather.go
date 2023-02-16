package weather

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"qq/bot"
	"qq/features"
)

const weatherURL = "https://restapi.amap.com/v3/weather/weatherInfo?key=%s&city=%s&extensions=all&output=json"

var temp, _ = template.New("").Parse(`
城市: {{.City}}

白天: {{.Dayweather}} {{.Daytemp}} ℃
夜晚: {{.Nightweather}} {{.Nighttemp}} ℃
`)

type input struct {
	City         string
	Date         string
	Dayweather   string
	Daytemp      string
	Nightweather string
	Nighttemp    string
}

func init() {
	features.AddKeyword("天气", "+<城市> 获取天气信息, 默认杭州", func(bot bot.Bot, city string) error {
		if city == "" {
			city = "杭州"
		}
		bot.Send(get(city))
		return nil
	})
}

func get(city string) string {
	get, _ := http.Get(fmt.Sprintf(weatherURL, os.Getenv("WEATHER_KEY"), city))
	defer closeBody(get.Body)
	var res response
	json.NewDecoder(get.Body).Decode(&res)
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
