package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"qq/features/stock/ai"

	"github.com/sashabaranov/go-openai"
)

type toolCallChatWrapper struct {
	openaiClient *openaiClient
	send         func(msg string) string
}

func (t *toolCallChatWrapper) StreamCompletion(ctx context.Context, tm *ai.History) (<-chan ai.CompletionResponse, error) {
	completion, err := t.openaiClient.streamCompletion(ctx, tm.Messages())
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
			tm.Add(openai.ChatCompletionMessage{
				Role:      openai.ChatMessageRoleAssistant,
				ToolCalls: toolCalls,
			})
			for _, call := range toolCalls {
				if t.send != nil {
					pargs := call.Function.Arguments
					var v map[string]any
					if pargs != "" {
						json.Unmarshal([]byte(pargs), &v)
						delete(v, "UID")
						delete(v, "FromGroup")
						delete(v, "GroupID")
						delete(v, "From")
						indent, _ := json.Marshal(v)
						pargs = string(indent)
					}
					t.send(fmt.Sprintf(`正在调用函数: %v
参数为: %v`, call.Function.Name, pargs))
				}
				callResult, err := t.openaiClient.toolCall(call.Function.Name, call.Function.Arguments)
				if err != nil {
					resCh <- &ai.CompletionResponseImpl{Error: err}
					return
				}

				tm.Add(openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    callResult,
					ToolCallID: call.ID,
				})
			}
			streamCompletion, err := t.StreamCompletion(ctx, tm)
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
