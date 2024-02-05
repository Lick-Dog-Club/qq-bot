package stock

import (
	"context"
	"fmt"
	"qq/bot"
	config2 "qq/config"
	"qq/features"
	"qq/features/stock/ai"
	openai2 "qq/features/stock/openai"
	"qq/features/stock/tools"
	"qq/features/stock/types"
	"qq/util/proxy"
	"time"
)

func init() {
	features.AddKeyword("stock", "股票分析", func(bot bot.Bot, content string) error {
		client := openai2.NewOpenaiClient(openai2.NewClientOption{
			HttpClient:  proxy.NewHttpProxyClient(),
			Token:       config2.AiToken(),
			Model:       "gpt-4-0125-preview",
			MaxToken:    4096,
			Temperature: 0.2,
			Tools:       tools.GetByNames(true),
		})
		completion, _ := (&openai2.ToolCallChatWrapper{
			Client: client,
		}).StreamCompletion(context.TODO(), []ai.Message{
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
		bot.Send(str)
		return nil
	})
}
