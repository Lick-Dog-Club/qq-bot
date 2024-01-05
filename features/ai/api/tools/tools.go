package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"qq/features/comic"
	"qq/features/kfc"
	"qq/features/picture"
	"qq/features/pixiv"
	"qq/features/sysupdate"
	"qq/features/weather"
	"qq/features/weibo"
	"qq/features/zhihu"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func Call(funcName string, params string) (string, error) {
	fmt.Println("call: ", funcName)
	switch funcName {
	case "GetWeather":
		var city = struct {
			City string `json:"city"`
		}{}
		json.Unmarshal([]byte(params), &city)

		return weather.Get(city.City), nil
	case "GetZhiHuTop50":
		return zhihu.Top(), nil
	case "KFC":
		return kfc.Get(), nil
	case "SendPicture":
		var err error
		img := ""
		if img, err = pixiv.Image("n"); err != nil {
			img = picture.Url()
		}
		return img, nil
	case "Comic":
		var t = struct {
			Title string `json:"title"`
		}{}
		json.Unmarshal([]byte(params), &t)
		return comic.Get(t.Title, -1).Render(), nil
	case "SystemVersion":
		return sysupdate.Version(), nil
	case "CurrentDate":
		return time.Now().Local().Format(time.DateTime), nil
	case "WeiBo":
		return weibo.Top(), nil
	default:
	}
	return "", errors.New("not support")
}

func List() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name: "GetWeather",
				Parameters: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"city": {
							Type:        jsonschema.String,
							Description: "The city and state, e.g. 天津, 北京",
						},
					},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name: "CurrentDate",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "返回当前的时间信息",
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name: "SendPicture",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "获取一张图片的url地址",
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name: "WeiBo",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "获取微博热搜数据",
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name: "Comic",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "获取动漫/漫画/蕃剧的资讯信息",
					Properties: map[string]jsonschema.Definition{
						"title": {
							Type:        jsonschema.String,
							Description: "动漫的名字, 中文或者拼音",
						},
					},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name: "SystemVersion",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "获取系统版本",
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name: "KFC",
				Parameters: &jsonschema.Definition{
					Description: "获取每周四 kfc v50 骚话",
					Type:        jsonschema.Object,
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name: "GetZhiHuTop50",
				Parameters: &jsonschema.Definition{
					Description: "获取 知乎 热搜top50",
					Type:        jsonschema.Object,
				},
			},
		},
	}
}
