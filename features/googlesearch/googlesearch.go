package googlesearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/features/stock/ai"
	"qq/features/stock/httpproxy"
	openai2 "qq/features/stock/openai"
	"qq/features/stock/types"
	"qq/util/proxy"

	"github.com/k3a/html2text"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func init() {
	features.AddKeyword("google_search", "使用 google 搜索", func(bot bot.Bot, content string) error {
		bot.Send("不支持此功能！")
		return nil
	}, features.WithGroup("google_search"), features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"query": {
				Type:        jsonschema.String,
				Description: "搜索关键词",
			},
			"recency_days": {
				Type:        jsonschema.Integer,
				Description: "搜索的时间范围, 单位是天",
				Enum:        []string{"7", "30", "60"},
			},
		},
		Call: func(args string) (string, error) {
			var s = struct {
				Query       string `json:"query"`
				RecencyDays int    `json:"recency_days"`
			}{}
			json.Unmarshal([]byte(args), &s)
			search, _ := Search(s.Query, s.RecencyDays)
			marshal, _ := json.Marshal(search)
			return string(marshal), nil
		},
	}), features.WithHidden())

	features.AddKeyword("mclick", "通过链接获取内容详情", func(bot bot.Bot, content string) error {
		bot.Send("不支持此功能！")
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"links": {
				Type:        jsonschema.Array,
				Description: "链接列表, 最多给 5 个链接地址",
				Items: &jsonschema.Definition{
					Type: jsonschema.String,
				},
			},
		},
		Call: func(args string) (string, error) {
			var s = struct {
				Links []string `json:"links"`
			}{}
			json.Unmarshal([]byte(args), &s)
			if len(s.Links) > 5 {
				s.Links = s.Links[:5]
			}
			var result = Mclick(s.Links...)
			marshal, _ := json.Marshal(result)
			return string(marshal), nil
		},
	}), features.WithHidden())
}

func Search(query string, recencyDays int) (*Response, error) {
	date := fmt.Sprintf("d[%d]", recencyDays)
	client := httpproxy.NewHttpProxyClient(config.HttpProxy())

	query = url.QueryEscape(query)
	resp, err := client.Get(fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&dateRestrict=%s", config.GoogleSearchKey(), config.GoogleSearchCX(), query, date))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	all, _ := io.ReadAll(resp.Body)
	var data Response
	json.Unmarshal(all, &data)
	return &data, nil
}

type ClickResult struct {
	Link    string `json:"link"`
	Summary string `json:"summary"`
}

func Mclick(links ...string) []*ClickResult {
	var result []*ClickResult
	for _, link := range links {
		page, err := viewPage(link)
		if err != nil {
			fmt.Println(err)
			continue
		}
		result = append(result, page)
	}
	return result
}

func viewPage(link string) (*ClickResult, error) {
	parse, err2 := url.Parse(link)
	if err2 != nil {
		return nil, err2
	}
	request, err2 := http.NewRequest("GET", parse.String(), nil)
	request.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	request.Header.Add("Referer", fmt.Sprintf("%s://%s", parse.Scheme, parse.Host))
	resp, err2 := http.DefaultClient.Do(request)
	if err2 != nil {
		return nil, err2
	}
	defer resp.Body.Close()
	all, _ := io.ReadAll(resp.Body)
	text := html2text.HTML2Text(string(all))

	client := openai2.NewOpenaiClient(openai2.NewClientOption{
		HttpClient:  proxy.NewHttpProxyClient(),
		Token:       config.AiToken(),
		Model:       openai.GPT3Dot5Turbo16K0613,
		MaxToken:    2000,
		Temperature: 0.2,
	})
	completion, err := client.Completion(context.TODO(), []ai.Message{
		{
			Role:    types.RoleUser,
			Content: fmt.Sprintf("总结以下内容, 500字以内：\n%s", text),
		},
	})
	if err != nil {
		return nil, err
	}
	var msg string
	if len(completion.GetChoices()) > 0 {
		msg = completion.GetChoices()[0].Message.Content
	}
	return &ClickResult{
		Link:    link,
		Summary: msg,
	}, nil
}

type Response struct {
	Items []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"items"`
}
