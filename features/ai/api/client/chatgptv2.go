package client

import (
	"context"
	"qq/features"
	"qq/features/ai/api/types"
	"qq/features/stock/ai"
	openai2 "qq/features/stock/openai"
	"qq/util/proxy"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/sashabaranov/go-openai"
)

type openaiClientV2 struct {
	apikey string
	model  string
	opt    openai.ChatCompletionRequest

	cli ai.Chat
}

func NewOpenaiClientV2(apikey string, model string, opt openai.ChatCompletionRequest) types.GptClientImpl {
	client := openai2.NewOpenaiClient(openai2.NewClientOption{
		HttpClient:  proxy.NewHttpProxyClient(),
		Token:       apikey,
		Model:       model,
		MaxToken:    4096,
		Temperature: 0.2,
		Tools:       features.AllFuncCalls(),
		ToolCall:    features.CallFunc,
	})
	return &openaiClientV2{
		apikey: apikey,
		model:  model,
		opt:    opt,
		cli:    client,
	}
}

func (gpt *openaiClientV2) Platform() string {
	return "chatgpt-v2"
}

func (gpt *openaiClientV2) GetCompletion(his *ai.History, current openai.ChatCompletionMessage) (string, error) {
	timeout, cancelFunc := context.WithTimeout(context.TODO(), 120*time.Second)
	defer cancelFunc()
	his.Add(current)
	completion, err := gpt.cli.StreamCompletion(timeout, his)
	if err != nil {
		return err.Error(), nil
	}
	str := ""
	for resp := range completion {
		if resp.IsEnd() {
			if resp.GetError() != nil {
				log.Println(resp.GetError())
			}
			break
		}
		if len(resp.GetChoices()) > 0 {
			str += resp.GetChoices()[0].Message.Content
		}
	}
	if str == "" {
		str = "······"
	}
	return str, nil
}
