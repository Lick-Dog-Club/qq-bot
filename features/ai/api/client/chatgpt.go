package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/sashabaranov/go-openai"
	"log"
	"qq/features/ai/api/tools"
	"qq/features/ai/api/types"
	"qq/util/proxy"
	"time"
)

type openaiClient struct {
	apikey string
	model  string
	opt    openai.ChatCompletionRequest
}

func NewOpenaiClient(apikey string, model string, opt openai.ChatCompletionRequest) types.GptClientImpl {
	return &openaiClient{apikey: apikey, model: model, opt: opt}
}

func (gpt *openaiClient) Platform() string {
	return "chatgpt"
}

func (gpt *openaiClient) GetCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	req := gpt.opt
	req.Model = gpt.model
	req.MaxTokens = 2048
	req.Stream = false
	req.Messages = messages
	cfg := openai.DefaultConfig(gpt.apikey)
	cfg.HTTPClient = proxy.NewHttpProxyClient()
	c := openai.NewClientWithConfig(cfg)
	log.Println("send request")
	timeout, cancelFunc := context.WithTimeout(context.TODO(), time.Second*150)
	defer cancelFunc()
	stream, err := c.CreateChatCompletion(timeout, req)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}
	if len(stream.Choices) < 1 {
		return "", errors.New("data.Choices < 1")
	}
	for len(stream.Choices) > 0 && len(stream.Choices[0].Message.ToolCalls) > 0 {
		messages = append(messages, stream.Choices[0].Message)
		for idx := range stream.Choices[0].Message.ToolCalls {
			cc := stream.Choices[0].Message.ToolCalls[idx]
			var content string
			content, err = tools.Call(cc.Function.Name, cc.Function.Arguments)
			fmt.Println(content, err)
			if err != nil {
				break
			}
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    content,
				Name:       cc.Function.Name,
				ToolCallID: cc.ID,
			})
		}
		req.Messages = messages
		backoff.Retry(func() error {
			stream, err = c.CreateChatCompletion(context.TODO(), req)
			fmt.Println(err)
			return err
		}, backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Second), 5))
	}

	if len(stream.Choices) > 0 {
		return stream.Choices[0].Message.Content, nil
	}
	return "", errors.New("no choice")
}
