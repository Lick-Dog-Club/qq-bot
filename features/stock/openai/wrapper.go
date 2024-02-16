package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"qq/features/stock/ai"
	"qq/features/stock/impl"
	"qq/features/stock/tools"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

type ToolCallChatWrapper struct {
	Client ai.Chat
}

func (t *ToolCallChatWrapper) StreamCompletion(ctx context.Context, messages []ai.Message) (<-chan ai.CompletionResponse, error) {
	completion, err := t.Client.StreamCompletion(ctx, messages)
	if err != nil {
		return nil, err
	}
	resCh := make(chan ai.CompletionResponse, 100)
	go func() {
		var (
			isToolCall bool
			toolCalls  []openai.ToolCall
		)
		defer close(resCh)
		for resp := range completion {
			if resp.IsToolCall() {
				isToolCall = true
				toolCalls = resp.GetToolCalls()
				continue
			}
			resCh <- resp
		}

		if isToolCall {
			messages = append(messages,
				ai.Message{
					Role:     openai.ChatMessageRoleAssistant,
					ToolCall: toolCalls,
				},
			)
			for _, call := range toolCalls {
				tool, err := CallTool(call)
				if err != nil {
					resCh <- &ai.CompletionResponseImpl{Error: err}
					return
				}

				messages = append(messages, ai.Message{
					Role:       openai.ChatMessageRoleTool,
					Content:    tool.Content,
					ToolCallID: call.ID,
				})
			}
			streamCompletion, err := t.StreamCompletion(ctx, messages)
			if err != nil {
				resCh <- &ai.CompletionResponseImpl{Error: err}
				return
			}
			for s := range streamCompletion {
				resCh <- s
			}
		}
	}()
	return resCh, nil
}

type CallResult struct {
	Content string
}

// CallTool
// TODO 重构下
func CallTool(
	tool openai.ToolCall,
) (*CallResult, error) {
	plugin, err := tools.GetPluginNameByFunctionName(tool.Function.Name)
	if err != nil {
		return nil, err
	}
	fmt.Println(plugin.Name, tool.Function.Arguments)
	switch plugin.Name {
	case "GetMarketSentiment":
		var input impl.GetMarketSentimentRequest
		json.NewDecoder(strings.NewReader(tool.Function.Arguments)).Decode(&input)
		return &CallResult{Content: impl.GetMarketSentiment(input)}, nil
	case "GetFinancialStatements":
		var input impl.GetFinancialStatementsRequest
		json.NewDecoder(strings.NewReader(tool.Function.Arguments)).Decode(&input)
		return &CallResult{Content: impl.GetFinancialStatements(input)}, nil
	case "GetCashFlow":
		var input impl.GetCashFlowRequest
		json.NewDecoder(strings.NewReader(tool.Function.Arguments)).Decode(&input)
		return &CallResult{Content: impl.GetCashFlow(input)}, nil
	case "GetIndustryData":
		return &CallResult{Content: impl.GetIndustryData()}, nil
	case "GetStockPrice":
		var input impl.GetStockPriceRequest
		json.NewDecoder(strings.NewReader(tool.Function.Arguments)).Decode(&input)
		price := impl.GetStockPrice(input)
		marshal, _ := json.Marshal(price)
		return &CallResult{
			Content: string(marshal),
		}, nil
	case tools.BuildInPluginCurrentDatetime.Name:
		return &CallResult{
			Content: time.Now().Format(time.DateTime),
		}, nil
	}
	return nil, fmt.Errorf("plugin call '%s' not impl", plugin.Name)
}
