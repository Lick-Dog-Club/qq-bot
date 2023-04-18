package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"qq/features/ai/api/types"
	"qq/features/util/proxy"

	"github.com/sashabaranov/go-openai"
)

type azureClient struct {
	apikey  string
	model   string
	opt     openai.ChatCompletionRequest
	baseUrl string
}

func NewAzureClient(apikey string, model string, opt openai.ChatCompletionRequest, baseurl string) types.GptClientImpl {
	return &azureClient{apikey: apikey, model: model, opt: opt, baseUrl: baseurl}
}

func (gpt *azureClient) Platform() string {
	return "azure"
}

func (gpt *azureClient) GetCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	req := gpt.opt
	req.Model = gpt.model
	req.MaxTokens = 800
	req.Stream = false
	req.Messages = messages
	marshal, _ := json.Marshal(req)
	request, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2023-03-15-preview", gpt.baseUrl, gpt.model),
		bytes.NewReader(marshal),
	)
	log.Println(fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2023-03-15-preview", gpt.baseUrl, gpt.model))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("api-key", gpt.apikey)

	do, err := proxy.NewHttpProxyClient().Do(request)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer do.Body.Close()
	var data response
	json.NewDecoder(do.Body).Decode(&data)
	if len(data.Choices) < 1 {
		log.Println(do.StatusCode)
		return "", errors.New("data.Choices < 1")
	}
	return data.Choices[0].Message.Content, nil
}

type response struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}
