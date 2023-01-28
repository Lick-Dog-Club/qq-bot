package ai

import (
	"context"
	"log"
	"os"
	"strings"

	gogpt "github.com/sashabaranov/go-gpt3"
)

var (
	token = os.Getenv("AI_TOKEN")
	chat  = gogpt.NewClient(token)
)

func Request(ask string) string {
	ctx := context.Background()
	ask = strings.TrimSpace(ask)

	req := gogpt.CompletionRequest{
		Model:            gogpt.GPT3TextDavinci003,
		Prompt:           ask,
		Temperature:      0,
		BestOf:           1,
		Echo:             true,
		FrequencyPenalty: 0,
		LogProbs:         0,
		MaxTokens:        1000,
		PresencePenalty:  0,
		TopP:             1,
	}
	resp, err := chat.CreateCompletion(ctx, req)
	if err != nil {
		return err.Error()
	}
	log.Printf("%T", resp.Choices[0])
	return strings.TrimPrefix(strings.TrimPrefix(resp.Choices[0].Text, ask+"\n\n"), ask+"\n")
}
