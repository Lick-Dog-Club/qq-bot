package config

import (
	"os"
	"strconv"
)

var (
	AIToken    = os.Getenv("AI_TOKEN")
	GroupId    = toInt(os.Getenv("GROUP_ID"))
	Namespace  = os.Getenv("APP_NAMESPACE")
	Pod        = os.Getenv("POD_NAME")
	WeatherKey = os.Getenv("WEATHER_KEY")
	TianApiKey = os.Getenv("TIAN_API_KEY")
)

func toInt(s string) int {
	atoi, _ := strconv.Atoi(s)
	return atoi
}
