package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"qq/bot"
	config2 "qq/config"
	"qq/features"
	"qq/features/ai/api"
	"qq/features/stock/ai"
	openai2 "qq/features/stock/openai"
	"qq/features/stock/types"
	"qq/util/proxy"
	"qq/util/retry"

	"github.com/sashabaranov/go-openai"

	"github.com/sashabaranov/go-openai/jsonschema"

	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("clear", "清除 ai 历史对话记录", func(bot bot.Bot, content string) error {
		api.Clear(userID(bot))
		bot.Send("done")
		return nil
	}, features.WithGroup("ai"))
	features.SetDefault("ai 自动回答", func(bot bot.Bot, content string) error {
		req := api.Request
		log.Printf("%s: %s", bot.UserID(), content)
		bot.Send(req(userID(bot), content, bot.From(), bot.UserID(), bot.GroupID()))
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
	}), features.WithGroup("ai"))
	features.AddKeyword("see", "<+图片url>: 根据 url 识别图片内容", func(bot bot.Bot, content string) error {
		bot.Send(See([]string{content}))
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"images": {
				Type:        jsonschema.Array,
				Description: "图片的url地址列表",
				Items: &jsonschema.Definition{
					Type:        jsonschema.String,
					Description: "图片url地址",
				},
			},
		},
		Call: func(args string) (string, error) {
			var input = struct {
				Images []string `json:"images"`
			}{}
			json.Unmarshal([]byte(args), &input)
			return See(input.Images), nil
		},
	}), features.WithGroup("ai"))
}

func userID(bot bot.Bot) string {
	return fmt.Sprintf("%s:%v", bot.UserID(), bot.IsGroupMessage())
}

func See(images []string) string {
	client := openai2.NewOpenaiClient(openai2.NewClientOption{
		HttpClient: proxy.NewHttpProxyClient(),
		MaxToken:   4096,
		Token:      config2.AiToken(),
		Model:      openai.GPT4VisionPreview,
	})
	var cnt []openai.ChatMessagePart
	for _, image := range images {
		cnt = append(cnt, openai.ChatMessagePart{
			Type: "image_url",
			ImageURL: &openai.ChatMessageImageURL{
				URL: image,
			},
		})
	}
	cnt = append(cnt, openai.ChatMessagePart{
		Type: "text",
		Text: "详细描述图片内容",
	})
	var res ai.CompletionResponse
	var err error
	e := retry.Times(5, func() error {
		res, err = client.Completion(context.TODO(), []ai.Message{
			{
				Role:         types.RoleUser,
				MultiContent: cnt,
			},
		})
		if err != nil {
			return err
		}
		return nil
	})
	if e != nil {
		return e.Error()
	}
	return res.GetChoices()[0].Message.Content
}

func Draw(prompt string) (string, error) {
	client := openai2.NewOpenaiClient(openai2.NewClientOption{
		HttpClient: proxy.NewHttpProxyClient(),
		Token:      config2.AiToken(),
	})
	res, err := client.CreateImage(context.TODO(), prompt, "", "")
	if err != nil {
		return "", err
	}
	return res.URL(), nil
}
