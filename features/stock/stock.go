package stock

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"qq/bot"
	config2 "qq/config"
	"qq/features"
	"qq/features/stock/ai"
	"qq/features/stock/impl"
	openai2 "qq/features/stock/openai"
	"qq/features/stock/tools"
	"qq/features/stock/types"
	"qq/util/proxy"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func init() {
	features.AddKeyword("stock", "分析股票", func(bot bot.Bot, content string) error {
		bot.Send(Analyze(content))
		return nil
	})
	features.AddKeyword("now", "获取当前时间", func(bot bot.Bot, content string) error {
		bot.Send(time.Now().Format(time.DateTime))
		return nil
	}, features.WithHidden(), features.WithAIFunc(features.AIFuncDef{
		Properties: nil,
		Call: func(args string) (string, error) {
			return time.Now().Format(time.DateTime), nil
		},
	}))
	features.AddKeyword("stockcode", "根据名称获取股票代码", func(bot bot.Bot, content string) error {
		bot.Send(GetCodeByName(content))
		return nil
	}, features.WithGroup("stock"), features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"name": {
				Type:        jsonschema.String,
				Description: "股票名称, 例如 中国平安，浪潮信息",
			},
		},
		Call: func(args string) (string, error) {
			var s = struct {
				Name string `json:"name"`
			}{}
			json.Unmarshal([]byte(args), &s)
			return GetCodeByName(s.Name), nil
		},
	}))
	features.AddKeyword(impl.ToolGetStockPrice.Name, impl.ToolGetStockPrice.Define.Function.Description, func(bot bot.Bot, content string) error {
		return nil
	}, features.WithHidden(), features.WithAIFunc(features.AIFuncDef{
		Properties: impl.ToolGetStockPrice.Define.Function.Parameters.(*jsonschema.Definition).Properties,
		Call: func(args string) (string, error) {
			return impl.CallTool(impl.ToolGetStockPrice.Name, args)
		},
	}), features.WithGroup("stock"))
	features.AddKeyword(impl.ToolsGetCashFlow.Name, impl.ToolsGetCashFlow.Define.Function.Description, func(bot bot.Bot, content string) error {
		return nil
	}, features.WithHidden(), features.WithAIFunc(features.AIFuncDef{
		Properties: impl.ToolsGetCashFlow.Define.Function.Parameters.(*jsonschema.Definition).Properties,
		Call: func(args string) (string, error) {
			return impl.CallTool(impl.ToolsGetCashFlow.Name, args)
		},
	}), features.WithGroup("stock"))
	features.AddKeyword(impl.ToolsGetIndustryData.Name, impl.ToolsGetIndustryData.Define.Function.Description, func(bot bot.Bot, content string) error {
		return nil
	}, features.WithHidden(), features.WithAIFunc(features.AIFuncDef{
		Properties: impl.ToolsGetIndustryData.Define.Function.Parameters.(*jsonschema.Definition).Properties,
		Call: func(args string) (string, error) {
			return impl.CallTool(impl.ToolsGetIndustryData.Name, args)
		},
	}), features.WithGroup("stock"))
	features.AddKeyword(impl.ToolsGetMarketSentiment.Name, impl.ToolsGetMarketSentiment.Define.Function.Description, func(bot bot.Bot, content string) error {
		return nil
	}, features.WithHidden(), features.WithAIFunc(features.AIFuncDef{
		Properties: impl.ToolsGetMarketSentiment.Define.Function.Parameters.(*jsonschema.Definition).Properties,
		Call: func(args string) (string, error) {
			return impl.CallTool(impl.ToolsGetMarketSentiment.Name, args)
		},
	}), features.WithGroup("stock"))
	features.AddKeyword(impl.ToolsGetFinancialStatements.Name, impl.ToolsGetFinancialStatements.Define.Function.Description, func(bot bot.Bot, content string) error {
		return nil
	}, features.WithHidden(), features.WithAIFunc(features.AIFuncDef{
		Properties: impl.ToolsGetFinancialStatements.Define.Function.Parameters.(*jsonschema.Definition).Properties,
		Call: func(args string) (string, error) {
			return impl.CallTool(impl.ToolsGetFinancialStatements.Name, args)
		},
	}), features.WithGroup("stock"))
}

func Analyze(content string) string {
	client := openai2.NewOpenaiClient(openai2.NewClientOption{
		HttpClient:  proxy.NewHttpProxyClient(),
		Token:       config2.AiToken(),
		Model:       "gpt-4-0125-preview",
		MaxToken:    4096,
		Temperature: 0.2,
		Tools: []tools.Tool{
			impl.ToolGetStockPrice,
			impl.ToolsGetCashFlow,
			impl.ToolsGetIndustryData,
			impl.ToolsGetMarketSentiment,
			impl.ToolsGetFinancialStatements,
		},
		ToolCall: impl.CallTool,
	})
	completion, _ := client.StreamCompletion(context.TODO(), []ai.Message{
		{
			Role: types.RoleSystem,
			Content: fmt.Sprintf(`当前时间: %s.
你是短线炒股专家，拥有丰富的炒股经验，请你从多个方面分析股票适不适合短线投资, 时间范围是距今(包括今天)近一个月或三个月的数据

## 你需要从以下角度逐个分析

1. 技术分析
2. 市场情绪分析
3. 成交量分析
4. 最新市场动态
5. 短期价格动态

## 抄底或投资的判断依据：

技术指标的信号：多个技术指标（如RSI、MACD）给出超卖信号且股价接近支撑位，可能是抄底的机会。
成交量的变化：在重要支撑位附近，成交量突然增加，表明可能有买盘进入。
市场情绪：极度悲观的情绪往往预示着潜在的反弹机会，但需要结合其他因素综合判断。
监管公告或新闻：没有重大负面新闻或公告影响股票基本面，短期内的价格下跌可能仅仅是市场情绪的反应。
给出止盈止损的点位, 并且说出你分析的思路，并且做一个总结。
`, time.Now().Format(time.DateTime)),
		},
		{
			Role:    types.RoleUser,
			Content: content,
		},
	})
	str := ""
	for resp := range completion {
		if resp.IsEnd() {
			break
		}
		str += resp.GetChoices()[0].Message.Content
	}
	return str
}

func GetCodeByName(name string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://searchadapter.eastmoney.com/api/suggest/get?cb=jQuery1124020941285714467317_1708265862684&input=%s&type=8&token=D43BF722C8E33BDC906FB84D85E326E8&markettype=&mktnum=&jys=&classify=&securitytype=&status=&count=4&_=%v", url.QueryEscape(name), time.Now().UnixMilli()), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://www.eastmoney.com/")
	req.Header.Set("Sec-Fetch-Dest", "script")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not A(Brand";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var data response
	index := strings.Index(string(bodyText), "{")
	if err := json.NewDecoder(strings.NewReader(string(bodyText)[index:])).Decode(&data); err != nil {
		log.Fatal(err)
	}
	var result []struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}
	for _, datum := range data.GubaCodeTable.Data {
		result = append(result, struct {
			Name string `json:"name"`
			Code string `json:"code"`
		}{Name: datum.ShortName, Code: datum.OuterCode})
	}
	marshal, _ := json.Marshal(&result)
	return string(marshal)
}

type response struct {
	GubaCodeTable struct {
		Data []struct {
			ShortName         string `json:"ShortName"`
			Url               string `json:"Url"`
			ProtocolFollowUrl string `json:"ProtocolFollowUrl"`
			OuterCode         string `json:"OuterCode"`
			HeadCharacter     string `json:"HeadCharacter"`
			RelatedCode       string `json:"RelatedCode"`
		} `json:"Data"`
		Status     int    `json:"Status"`
		Message    string `json:"Message"`
		TotalCount int    `json:"TotalCount"`
		BizCode    string `json:"BizCode"`
		BizMsg     string `json:"BizMsg"`
	} `json:"GubaCodeTable"`
}
