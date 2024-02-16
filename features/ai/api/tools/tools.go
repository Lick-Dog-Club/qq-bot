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
	"qq/features/trainticket"
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
	case "GetStationCodeByName":
		var a = struct {
			Name string `json:"name"`
		}{}
		json.Unmarshal([]byte(params), &a)
		return trainticket.GetStationCode(a.Name), nil
	case "Search12306":
		var input trainticket.SearchInput
		json.Unmarshal([]byte(params), &input)
		return trainticket.Search(input).String(), nil
	default:
	}
	return "", errors.New("not support")
}

func List() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "CurrentDate",
				Description: "返回当前的时间信息",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "返回当前的时间信息",
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "GetStationCodeByName",
				Description: "返回高铁/火车车站名称和 code 的对应关系表",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "返回高铁/火车车站名称和 code 的对应关系表",
					Properties: map[string]jsonschema.Definition{
						"name": {
							Type:        jsonschema.String,
							Description: "地点，例如 '杭州东' '绍兴北' 等",
						},
					},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "Search12306",
				Description: "高铁/火车票查询，返回高铁班次信息，以及余票数量",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "高铁/火车票查询，返回高铁班次信息，以及余票数量",
					Properties: map[string]jsonschema.Definition{
						"from": {
							Type:        jsonschema.String,
							Description: "出发地, 需要通过 GetStationCodeByName 函数获取 code 值, 例如: 出发去杭州东, 需要根据 GetStationCodeByName 函数, 然后查到对应 from='HGH'",
						},
						"to": {
							Type:        jsonschema.String,
							Description: "目的地, 需要通过 GetStationCodeByName 函数获取 code 值, 例如: 出发去杭州东, 需要根据 GetStationCodeByName 函数, 然后查到对应 to='HGH'",
						},
						"date": {
							Type:        jsonschema.String,
							Description: "查询日期, 默认今天，日期格式: '2006-01-02', 例如: '2024-02-19'",
						},
						"only_show_ticket": {
							Type:        jsonschema.Boolean,
							Description: "是否只显示有票的班次",
						},
					},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "GetWeather",
				Description: "获取天气、气象数据",
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
				Name:        "Holidays",
				Description: "获取节假日数据, 获取法定节假日数据, 返回节日名称和具体的放假时间",
				Parameters: &jsonschema.Definition{
					Type: jsonschema.Object,
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
				Name:        "CreateImageByPrompt",
				Description: "通过提示词创建图片，并且返回图片的 url 地址",
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
				Name:        "SendPicture",
				Description: "获取一张随机的图片url地址",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "获取一张图片的url地址",
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "WeiBo",
				Description: "获取微博热搜数据",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "获取微博热搜数据",
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "Comic",
				Description: "获取动漫/漫画/蕃剧的资讯信息",
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
				Name:        "SystemVersion",
				Description: "获取系统版本",
				Parameters: &jsonschema.Definition{
					Type:        jsonschema.Object,
					Description: "获取系统版本",
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "KFC",
				Description: "获取每周四 kfc v50 骚话",
				Parameters: &jsonschema.Definition{
					Description: "获取每周四 kfc v50 骚话",
					Type:        jsonschema.Object,
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "GetZhiHuTop50",
				Description: "获取 知乎 热搜top50",
				Parameters: &jsonschema.Definition{
					Description: "获取 知乎 热搜top50",
					Type:        jsonschema.Object,
				},
			},
		},
	}
}
