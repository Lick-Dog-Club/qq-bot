package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"qq/config"
	"qq/features/comic"
	"qq/features/holiday"
	"qq/features/kfc"
	"qq/features/picture"
	"qq/features/pixiv"
	"qq/features/sysupdate"
	"qq/features/weather"
	"qq/features/weibo"
	"qq/features/zhihu"
	"qq/util/proxy"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func CreateImage(prompt string) string {
	cfg := openai.DefaultConfig(config.AiToken())
	cfg.HTTPClient = proxy.NewHttpProxyClient()
	c := openai.NewClientWithConfig(cfg)
	image, err := c.CreateImage(context.TODO(), openai.ImageRequest{
		Prompt:  prompt,
		Model:   openai.CreateImageModelDallE3,
		N:       1,
		Quality: openai.CreateImageQualityStandard,
		Size:    openai.CreateImageSize1024x1024,
	})
	if err != nil {
		return ""
	}
	fmt.Println(image.Data[0].URL)
	return image.Data[0].URL
}

func Call(funcName string, params string) (string, error) {
	fmt.Println("call: ", funcName)
	switch funcName {
	case "Holidays":
		var city = struct {
			Year int `json:"year"`
		}{}
		json.Unmarshal([]byte(params), &city)

		return holiday.Get(city.Year), nil
	case "GetWeather":
		var city = struct {
			City string `json:"city"`
		}{}
		json.Unmarshal([]byte(params), &city)

		return weather.Get(city.City), nil
	case "GetZhiHuTop50":
		return zhihu.Top(), nil
	case "CreateImageByPrompt":
		var prompt = struct {
			Prompt string `json:"prompt"`
		}{}
		json.Unmarshal([]byte(params), &prompt)

		return CreateImage(prompt.Prompt), nil
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
				Name: "Holidays",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "获取放假的日期，返回节日名称和具体的放假时间",
					Properties: map[string]jsonschema.Definition{
						"year": {
							Type:        jsonschema.Integer,
							Description: "4位数的年份, 例如 2024, 2023",
						},
					},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name: "CreateImageByPrompt",
				Parameters: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"prompt": {
							Type:        jsonschema.String,
							Description: "通过提示词创建图片，并且返回图片的 url 地址",
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
