package openai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"qq/features/stock/ai"
	"qq/features/stock/tools"
	"qq/features/stock/types"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
)

var _ ai.Chat = (*openaiClient)(nil)

type openaiClient struct {
	token       string
	model       string
	temperature float64
	maxToken    int
	httpClient  *http.Client
	tools       []tools.Tool

	client *openai.Client
}

type NewClientOption struct {
	// optional
	HttpClient *http.Client
	// required
	Token string
	// required
	Model string
	// optional
	MaxToken int
	// required
	Temperature float64
	// optional
	Tools    []tools.Tool
	ToolCall func(name string, args string) (string, error)
}

// NewOpenaiClient
//
// stream token 需要自己算： https://github.com/pkoukk/tiktoken-go
func NewOpenaiClient(opt NewClientOption) ai.Chat {
	config := openai.DefaultConfig(opt.Token)
	if opt.HttpClient != nil {
		config.HTTPClient = opt.HttpClient
	}
	return &openaiClient{
		maxToken:    opt.MaxToken,
		token:       opt.Token,
		model:       opt.Model,
		temperature: opt.Temperature,
		httpClient:  opt.HttpClient,
		tools:       opt.Tools,
		client:      openai.NewClientWithConfig(config),
	}
}

func (o *openaiClient) Completion(ctx context.Context, messages []ai.Message) (ai.CompletionResponse, error) {
	response, err := o.client.CreateChatCompletion(
		ctx,
		o.toRequest(messages, false),
	)
	if err != nil {
		return nil, o.toError(err)
	}

	var toolCalls []*openai.ToolCall

	choices := make([]*ai.Choice, 0, len(response.Choices))
	for _, choice := range response.Choices {
		// 判断是不是 toolCall
		if len(choice.Message.ToolCalls) > 0 {
			for _, call := range choice.Message.ToolCalls {
				fillToolCalls(&toolCalls, call)
			}
			continue
		}
		choices = append(choices, formatChoice(choice))
	}

	return &ai.CompletionResponseImpl{
		Choices: choices,
		Created: time.Unix(response.Created, 0),
		ID:      response.ID,
		Model:   response.Model,
		Usage: &ai.Usage{
			CompletionTokens: response.Usage.CompletionTokens,
			PromptTokens:     response.Usage.PromptTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
		ToolCalls: lo.Map(toolCalls, func(item *openai.ToolCall, index int) openai.ToolCall {
			return *item
		}),
	}, nil
}

func (o *openaiClient) CreateEmbeddings(ctx context.Context, texts []string) (ai.EmbeddingResponse, error) {
	resp, err := o.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: texts,
		Model: openai.AdaEmbeddingV2,
	})
	if err != nil {
		return nil, err
	}

	var embeddings [][]float32
	for _, data := range resp.Data {
		embeddings = append(embeddings, data.Embedding)
	}
	return &ai.EmbeddingResponseImpl{
		Model: "text-embedding-ada-002",
		Usage: &ai.Usage{
			CompletionTokens: resp.Usage.CompletionTokens,
			PromptTokens:     resp.Usage.PromptTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		Embeddings: embeddings,
	}, nil
}

func (o *openaiClient) StreamCompletion(ctx context.Context, messages []ai.Message) (<-chan ai.CompletionResponse, error) {
	fmt.Println(messages)
	return (&toolCallChatWrapper{stream: o}).StreamCompletion(ctx, messages)
}

func (o *openaiClient) streamCompletion(ctx context.Context, messages []ai.Message) (<-chan ai.CompletionResponse, error) {
	ch := make(chan ai.CompletionResponse, 100)
	stream, err := o.client.CreateChatCompletionStream(
		ctx,
		o.toRequest(messages, true),
	)
	if err != nil {
		close(ch)
		return nil, o.toError(err)
	}
	go func() {
		defer func() {
			stream.Close()
			close(ch)
		}()

		var (
			streamErr  error
			response   openai.ChatCompletionStreamResponse
			toolCalls  []*openai.ToolCall
			isToolCall bool
		)

		for {
			response, streamErr = stream.Recv()
			if streamErr != nil {
				if !isToolCall {
					ch <- &ai.CompletionResponseImpl{Error: streamErr}
				}
				break
			}

			var hasContent bool
			choices := make([]*ai.Choice, 0, len(response.Choices))
			for _, choice := range response.Choices {
				// 判断是不是 toolCall
				if len(choice.Delta.ToolCalls) > 0 {
					isToolCall = true
					for _, call := range choice.Delta.ToolCalls {
						fillToolCalls(&toolCalls, call)
					}
					continue
				}
				cho := formatStreamChoice(choice)
				if cho.Message.Content != "" {
					hasContent = true
				}
				choices = append(choices, cho)
			}

			if !hasContent || isToolCall {
				continue
			}

			if !isToolCall {
				ch <- &ai.CompletionResponseImpl{
					Choices: choices,
					Created: time.Unix(response.Created, 0),
					ID:      response.ID,
					Model:   response.Model,
					Usage:   nil,
				}
			}
		}
		if isToolCall {
			ch <- &ai.CompletionResponseImpl{
				Created: time.Unix(response.Created, 0),
				ID:      response.ID,
				Model:   response.Model,
				ToolCalls: lo.Map(toolCalls, func(item *openai.ToolCall, index int) openai.ToolCall {
					return *item
				}),
			}
		}
	}()
	return ch, nil
}

func fillToolCalls(calls *[]*openai.ToolCall, call openai.ToolCall) {
	last, _ := lo.Last(*calls)
	if call.ID == "" && last != nil {
		last.Function.Arguments += call.Function.Arguments
	} else {
		*calls = append(*calls, &openai.ToolCall{
			Index: call.Index,
			ID:    call.ID,
			Type:  call.Type,
			Function: openai.FunctionCall{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		})
	}
}

type CreateImageInput struct {
	Prompts []string `json:"prompts"`
	Quality string   `json:"quality"`
	Size    string   `json:"size"`
}

func (o *openaiClient) CreateImage(ctx context.Context, prompt string, quality string, size string) (response ai.ImageResponse, err error) {
	if quality == "" {
		quality = openai.CreateImageQualityStandard
	}
	if size == "" {
		size = openai.CreateImageSize1024x1024
	}
	var resp openai.ImageResponse
	if err = backoff.Retry(func() error {
		resp, err = o.client.CreateImage(ctx, openai.ImageRequest{
			Prompt:  prompt,
			Model:   openai.CreateImageModelDallE3,
			N:       1,
			Quality: quality,
			Size:    size,
		})

		return err
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(500*time.Millisecond), 3)); err != nil {
		return nil, o.toError(err)
	}

	return &ai.ImageResponseImpl{
		Created: resp.Created,
		Url:     resp.Data[0].URL,
	}, nil
}

func (o *openaiClient) toError(err error) error {
	var e = &openai.APIError{}
	if errors.As(err, &e) {
		if e.HTTPStatusCode == 429 {
			err = ai.ErrorToManyRequests
		}
		if e.Code == "context_length_exceeded" {
			err = ai.ContextLengthExceeded
		}
	}

	return err
}

func (o *openaiClient) toRequest(messages []ai.Message, stream bool) openai.ChatCompletionRequest {
	msgs := make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, msg := range messages {
		msgs = append(msgs, formatMessage(msg))
	}
	req := openai.ChatCompletionRequest{
		Model:       o.model,
		Messages:    msgs,
		Temperature: float32(o.temperature),
		Stream:      stream,
	}
	if len(o.tools) > 0 {
		req.ToolChoice = "auto"
		req.Tools = lo.Map(o.tools, func(item tools.Tool, index int) openai.Tool {
			return item.Define
		})
	}
	if o.maxToken > 0 {
		req.MaxTokens = o.maxToken
	}
	return req
}

func formatStreamChoice(choice openai.ChatCompletionStreamChoice) *ai.Choice {
	return &ai.Choice{
		FinishReason: string(choice.FinishReason),
		Index:        choice.Index,
		Message: ai.Message{
			Role:    types.Role(choice.Delta.Role),
			Content: choice.Delta.Content,
		},
	}
}

func formatChoice(choice openai.ChatCompletionChoice) *ai.Choice {
	return &ai.Choice{
		FinishReason: string(choice.FinishReason),
		Index:        choice.Index,
		Message: ai.Message{
			Role:    types.Role(choice.Message.Role),
			Content: choice.Message.Content,
		},
	}
}

func formatMessage(message ai.Message) openai.ChatCompletionMessage {
	m := openai.ChatCompletionMessage{
		Role:       string(message.Role),
		Content:    message.Content,
		ToolCallID: message.ToolCallID,
	}

	if message.ToolCall != nil {
		m.ToolCalls = message.ToolCall
	}
	return m
}
