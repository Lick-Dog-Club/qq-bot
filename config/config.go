package config

import "os"

var (
	AIToken    = os.Getenv("AI_TOKEN")
	GroupId    = os.Getenv("GROUP_ID")
	Namespace  = os.Getenv("APP_NAMESPACE")
	Pod        = os.Getenv("POD_NAME")
	WeatherKey = os.Getenv("WEATHER_KEY")
)
