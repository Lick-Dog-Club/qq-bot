package ai

import (
	"context"
	"errors"
	"io"
	"qq/bot"
	"qq/features/stock/types"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/pkoukk/tiktoken-go"
	"github.com/samber/lo"

	"github.com/sashabaranov/go-openai"
)

var (
	_ CompletionResponse = (*CompletionResponseImpl)(nil)
	_ ImageResponse      = (*ImageResponseImpl)(nil)
	_ EmbeddingResponse  = (*EmbeddingResponseImpl)(nil)
)

type Chat interface {
	Completion(ctx context.Context, messages []Message) (CompletionResponse, error)
	StreamCompletion(ctx context.Context, messages *History) (<-chan CompletionResponse, error)
	CreateImage(ctx context.Context, prompt string, quality string, size string) (res ImageResponse, err error)
	CreateEmbeddings(context.Context, []string) (EmbeddingResponse, error)
}

type History struct {
	sync.RWMutex
	list      []openai.ChatCompletionMessage
	sysPrompt openai.ChatCompletionMessage
}

func (h *History) Messages() (resu []Message) {
	h.RLock()
	defer h.RUnlock()
	var n []openai.ChatCompletionMessage
	if h.sysPrompt.Role != "" {
		n = append(n, h.sysPrompt)
	}
	res := append(n, h.list...)
	for _, re := range res {
		resu = append(resu, Message{
			Role:       types.Role(re.Role),
			Content:    re.Content,
			ToolCall:   re.ToolCalls,
			ToolCallID: re.ToolCallID,
		})
	}
	return resu
}

func (h *History) SetSysPrompt(prompt string) {
	h.Lock()
	defer h.Unlock()
	h.sysPrompt = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: prompt,
	}
}

func (h *History) Add(message openai.ChatCompletionMessage) {
	h.Lock()
	defer h.Unlock()
	if ContentHasImage(message.Content) {
		message.Content = FormatImageContent(message.Content)
	}
	tokens := LastConversationsByLimitTokens(h.list, 4096)
	for len(tokens) > 0 {
		if tokens[0].Role != openai.ChatMessageRoleTool {
			break
		}
		tokens = tokens[1:]
	}
	h.list = tokens
	h.list = append(h.list, message)
}

var imageRegex = regexp.MustCompile(`\[\w+:image,file=(.*?),.*?]`)

func ContentHasImage(content string) bool {
	return imageRegex.MatchString(content)
}

func FormatImageContent(content string) string {
	submatch := imageRegex.FindAllStringSubmatch(content, -1)
	for _, i := range submatch {
		content = strings.ReplaceAll(content, i[0], bot.GetCQImage(i[1])+" ")
	}
	return content
}

func LastConversationsByLimitTokens(cs []openai.ChatCompletionMessage, limitTokenCount int64) []openai.ChatCompletionMessage {
	var (
		res        []openai.ChatCompletionMessage
		totalToken int
	)
	for _, item := range lo.Reverse(cs) {
		totalToken = totalToken + WordToToken(item.Content)
		if totalToken > int(limitTokenCount) {
			break
		}
		res = append(res, openai.ChatCompletionMessage{
			Role:         item.Role,
			Content:      item.Content,
			MultiContent: item.MultiContent,
			Name:         item.Name,
			FunctionCall: item.FunctionCall,
			ToolCalls:    item.ToolCalls,
			ToolCallID:   item.ToolCallID,
		})
	}
	return lo.Reverse(res)
}

// WordToToken 4,096 tokens
func WordToToken(s string) int {
	tkm, err := tiktoken.GetEncoding(tiktoken.MODEL_CL100K_BASE)
	if err != nil {
		return int(float64(utf8.RuneCountInString(s)) / 0.75)
	}
	return len(tkm.Encode(s, nil, nil))
}

type Message struct {
	Role         types.Role
	Content      string
	ImageUrls    []string
	MultiContent []openai.ChatMessagePart

	ToolCall []openai.ToolCall

	UUID       string
	ToolCallID string
}

type Usage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
	Message      Message `json:"message"`
}

type CompletionResponse interface {
	GetError() error
	Response() *CompletionResponseImpl
	GetModel() string
	GetUsage() *Usage
	GetChoices() []*Choice
	IsEnd() bool

	IsToolCall() bool
	GetToolCalls() []openai.ToolCall
}

type CompletionResponseImpl struct {
	Error   error
	Choices []*Choice `json:"choices"`
	Created time.Time `json:"created"`
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Usage   *Usage    `json:"usage"`

	ToolCalls []openai.ToolCall `json:"tool_calls"`
}

func (s *CompletionResponseImpl) IsEnd() bool {
	return errors.Is(io.EOF, s.Error) || s.Error != nil
}

func (s *CompletionResponseImpl) GetError() error {
	return s.Error
}

func (s *CompletionResponseImpl) Response() *CompletionResponseImpl {
	return s
}

func (s *CompletionResponseImpl) GetModel() string {
	return s.Model
}

func (s *CompletionResponseImpl) GetUsage() *Usage {
	return s.Usage
}

func (s *CompletionResponseImpl) GetChoices() []*Choice {
	return s.Choices
}

func (s *CompletionResponseImpl) IsToolCall() bool {
	return len(s.ToolCalls) > 0
}

func (s *CompletionResponseImpl) GetToolCalls() []openai.ToolCall {
	return s.ToolCalls
}

type EmbeddingResponse interface {
	GetModel() string
	GetUsage() *Usage
	GetEmbeddings() [][]float32
}

type EmbeddingResponseImpl struct {
	Model      string      `json:"model"`
	Usage      *Usage      `json:"usage"`
	Embeddings [][]float32 `json:"embeddings"`
}

func (e *EmbeddingResponseImpl) GetModel() string {
	return e.Model
}

func (e *EmbeddingResponseImpl) GetUsage() *Usage {
	return e.Usage
}

func (e *EmbeddingResponseImpl) GetEmbeddings() [][]float32 {
	return e.Embeddings
}

type ImageResponse interface {
	CreatedAt() time.Time
	URL() string
}

type ImageResponseImpl struct {
	Created int64
	Url     string
}

func (i *ImageResponseImpl) CreatedAt() time.Time {
	return time.Unix(i.Created, 0)
}

func (i *ImageResponseImpl) URL() string {
	return i.Url
}
