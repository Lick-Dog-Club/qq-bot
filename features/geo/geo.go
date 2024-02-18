package geo

import (
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai/jsonschema"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"
	"strings"
)

func init() {
	features.AddKeyword("geo", "geo 查询", func(bot bot.Bot, content string) error {
		geo := Geo(config.GeoKey(), content)
		split := strings.Split(geo, ",")
		if len(split) == 2 {
			lat := split[1]
			lng := split[0]
			bot.Send(fmt.Sprintf(`%s:
lat: %v
lng: %v
`, content, lat, lng))
		}
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"addr": {
				Type:        jsonschema.String,
				Description: "地址，例如杭州云水地铁口",
			},
		},
		Call: func(args string) (string, error) {
			var input = struct {
				Addr string `json:"addr"`
			}{}
			json.Unmarshal([]byte(args), &input)
			geo := Geo(config.GeoKey(), input.Addr)
			split := strings.Split(geo, ",")
			if len(split) == 2 {
				lat := split[1]
				lng := split[0]
				return fmt.Sprintf(`%s:
				lat: %v
				lng: %v
				`, input.Addr, lat, lng), nil
			}
			return "", nil
		},
	}))
}

func Geo(key string, addr string) string {
	resp, _ := http.Get(fmt.Sprintf("https://restapi.amap.com/v3/geocode/geo?key=%s&address=%s", key, addr))
	defer resp.Body.Close()
	var data response
	json.NewDecoder(resp.Body).Decode(&data)
	if len(data.Geocodes) > 0 {
		return data.Geocodes[0].Location
	}
	return ""
}

type response struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Count    string `json:"count"`
	Geocodes []struct {
		FormattedAddress string        `json:"formatted_address"`
		Country          string        `json:"country"`
		Province         string        `json:"province"`
		Citycode         string        `json:"citycode"`
		City             string        `json:"city"`
		District         string        `json:"district"`
		Township         []interface{} `json:"township"`
		Neighborhood     struct {
			Name []interface{} `json:"name"`
			Type []interface{} `json:"type"`
		} `json:"neighborhood"`
		Building struct {
			Name []interface{} `json:"name"`
			Type []interface{} `json:"type"`
		} `json:"building"`
		Adcode   string        `json:"adcode"`
		Street   []interface{} `json:"street"`
		Number   []interface{} `json:"number"`
		Location string        `json:"location"`
		Level    string        `json:"level"`
	} `json:"geocodes"`
}
