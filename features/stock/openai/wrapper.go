package openai

import (
	"context"
	"qq/features/stock/ai"

	"github.com/sashabaranov/go-openai"
)

type streamChat interface {
	streamCompletion(ctx context.Context, messages []ai.Message) (<-chan ai.CompletionResponse, error)
}

type toolCallChatWrapper struct {
	openaiClient *openaiClient
}

func (t *toolCallChatWrapper) StreamCompletion(ctx context.Context, messages []ai.Message) (<-chan ai.CompletionResponse, error) {
	completion, err := t.openaiClient.streamCompletion(ctx, messages)
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
				callResult, err := t.openaiClient.toolCall(call.Function.Name, call.Function.Arguments)
				if err != nil {
					resCh <- &ai.CompletionResponseImpl{Error: err}
					return
				}

				messages = append(messages, ai.Message{
					Role:       openai.ChatMessageRoleTool,
					Content:    callResult,
					ToolCallID: call.ID,
				})
			}
			streamCompletion, err := t.openaiClient.streamCompletion(ctx, messages)
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
