package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"qq/features/ai/api/types"
	"qq/features/util/proxy"
	"time"

	"github.com/sashabaranov/go-openai"
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
	req.MaxTokens = 800
	req.Stream = false
	req.Messages = messages
	cfg := openai.DefaultConfig(gpt.apikey)
	cfg.HTTPClient = proxy.NewHttpProxyClient()
	c := openai.NewClientWithConfig(cfg)
	log.Println("send request")
	timeout, cancelFunc := context.WithTimeout(context.TODO(), time.Second*15)
	defer cancelFunc()
	stream, err := c.CreateChatCompletion(timeout, req)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}
	if len(stream.Choices) < 1 {
		return "", errors.New("data.Choices < 1")
	}
	return stream.Choices[0].Message.Content, nil
}
