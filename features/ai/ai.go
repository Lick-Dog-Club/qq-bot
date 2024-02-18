package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"qq/bot"
	config2 "qq/config"
	"qq/features"
	"qq/features/ai/api"
	openai2 "qq/features/stock/openai"
	"qq/util/proxy"

	"github.com/sashabaranov/go-openai/jsonschema"

	log "github.com/sirupsen/logrus"
)

func init() {
	features.SetDefault("ai 自动回答", func(bot bot.Bot, content string) error {
		req := api.Request
		log.Printf("%s: %s", bot.UserID(), content)
		bot.Send(req(bot.UserID(), content))
		return nil
	})

	features.AddKeyword("draw", "<+prompt>: 使用 ai 画图", func(bot bot.Bot, content string) error {
		draw, _ := Draw(content)
		bot.Send(fmt.Sprintf("[CQ:image,file=%s]", draw))
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"prompt": {
				Type:        jsonschema.String,
				Description: "画图的提示词",
			},
		},
		Call: func(args string) (string, error) {
			var input = struct {
				Prompt string `json:"prompt"`
			}{}
			json.Unmarshal([]byte(args), &input)
			draw, _ := Draw(input.Prompt)
			return draw, nil
		},
	}))
}

func Draw(prompt string) (string, error) {
	client := openai2.NewOpenaiClient(openai2.NewClientOption{
		HttpClient:  proxy.NewHttpProxyClient(),
		Token:       config2.AiToken(),
		Model:       "gpt-4-0125-preview",
		MaxToken:    4096,
		Temperature: 0.2,
	})
	res, err := client.CreateImage(context.TODO(), prompt, "", "")
	if err != nil {
		return "", err
	}
	return res.URL(), nil
}
