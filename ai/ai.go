package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"qq/ai/encoder"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	token   = os.Getenv("AI_TOKEN")
	manager = NewGptManager(token)
)

func Request(userID int, ask string) string {
	user := manager.GetByUser(userID)
	if user.LastAskTime().Add(10 * time.Minute).Before(time.Now()) {
		manager.DeleteUser(userID)
		user = manager.GetByUser(userID)
	}
	return user.Send(ask)
}

type GptManager struct {
	sync.RWMutex
	users  map[int]*ChatGPTClient
	apiKey string
}

func NewGptManager(apiKey string) *GptManager {
	return &GptManager{apiKey: apiKey, users: map[int]*ChatGPTClient{}}
}

func (m *GptManager) DeleteUser(userID int) {
	m.Lock()
	defer m.Unlock()
	delete(m.users, userID)
}

func (m *GptManager) GetByUser(userID int) *ChatGPTClient {
	m.Lock()
	defer m.Unlock()
	client, ok := m.users[userID]
	if !ok {
		client = NewChatGPTClient(m.apiKey)
		m.users[userID] = client
	}
	return client
}

type Status struct {
	sync.RWMutex
	isAsking    bool
	opts        *SendOpts
	msg         string
	lastAskTime time.Time
}

func (s *Status) GetOpts() *SendOpts {
	s.RLock()
	defer s.RUnlock()
	if s.opts == nil {
		s.opts = &SendOpts{
			ConversationId:  uuid.NewString(),
			ParentMessageId: uuid.NewString(),
		}
	}
	return s.opts
}

func (s *Status) SetOpts(opts *SendOpts) {
	s.Lock()
	defer s.Unlock()
	s.opts = &SendOpts{
		ConversationId:  opts.ConversationId,
		ParentMessageId: opts.ParentMessageId,
	}
}

func (s *Status) IsAsking() bool {
	s.RLock()
	defer s.RUnlock()
	return s.isAsking
}

func (s *Status) LastAskTime() time.Time {
	s.RLock()
	defer s.RUnlock()
	return s.lastAskTime
}

func (s *Status) Msg() string {
	s.RLock()
	defer s.RUnlock()
	return s.msg
}

func (s *Status) SetMsg(msg string) {
	s.Lock()
	defer s.Unlock()
	s.msg = msg
}
func (s *Status) Asked() {
	s.Lock()
	defer s.Unlock()
	s.isAsking = false
}

func (s *Status) Asking() {
	s.Lock()
	defer s.Unlock()
	s.isAsking = true
	s.lastAskTime = time.Now()
}

type ChatGPTClient struct {
	lastAskTime time.Time
	apiKey      string
	opt         CompletionRequest
	cache       *KV
	status      *Status
}

func NewChatGPTClient(apiKey string) *ChatGPTClient {
	return &ChatGPTClient{
		apiKey: apiKey,
		opt: CompletionRequest{
			Model:           "text-chat-davinci-002-20230126",
			Temperature:     0.7,
			Stop:            []string{"<|im_end|>"},
			PresencePenalty: 0.6,
		},
		cache:  NewKV(map[string]any{"namespace": "chatgpt"}),
		status: &Status{},
	}
}

func (gpt *ChatGPTClient) LastAskTime() time.Time {
	return gpt.status.LastAskTime()
}

type SendOpts struct {
	ConversationId  string
	ParentMessageId string
}

type userMessage struct {
	id              string
	parentMessageId string
	role            string
	message         string
}

func (gpt *ChatGPTClient) Send(msg string) string {
	if gpt.status.IsAsking() {
		return "正在回答上一个问题~: " + gpt.status.Msg()
	}
	gpt.status.Asking()
	gpt.status.SetMsg(msg)
	var opts *SendOpts = gpt.status.GetOpts()
	var conversation []userMessage
	get := gpt.cache.Get(opts.ConversationId)
	if get == nil {
		conversation = []userMessage{}
	} else {
		conversation = get.([]userMessage)
	}
	um := userMessage{
		id:              uuid.NewString(),
		parentMessageId: opts.ParentMessageId,
		role:            "User",
		message:         msg,
	}
	conversation = append(conversation, um)
	prompt := gpt.buildPrompt(conversation, um.id)
	result, err := gpt.getCompletion(prompt)
	if err != nil {
		gpt.status.Asked()
		return err.Error()
	}
	reply := userMessage{
		id:              uuid.NewString(),
		parentMessageId: um.id,
		role:            "Assistant",
		message:         result,
	}
	conversation = append(conversation, reply)
	gpt.cache.Set(opts.ConversationId, conversation)
	gpt.status.SetOpts(&SendOpts{
		ConversationId:  opts.ConversationId,
		ParentMessageId: reply.id,
	})
	gpt.status.Asked()

	return reply.message
}

type userMessageList []userMessage

func (l userMessageList) Find(id string) *userMessage {
	for _, message := range l {
		if message.id == id {
			return &message
		}
	}
	return nil
}

type GptResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

type KV struct {
	kv map[string]any
	sync.RWMutex
}

func NewKV(kv map[string]any) *KV {
	return &KV{kv: kv}
}

func (kv *KV) Get(k string) any {
	kv.RLock()
	defer kv.RUnlock()
	return kv.kv[k]
}

func (kv *KV) Set(k string, v any) {
	kv.Lock()
	defer kv.Unlock()
	kv.kv[k] = v
}

func getTokenCount(text string) int {
	encoder, _ := encoder.NewEncoder()
	encode, _ := encoder.Encode(strings.ReplaceAll(text, `<|im_end|>`, `<|endoftext|>`))
	return len(encode)
}

func (gpt *ChatGPTClient) buildPrompt(messages userMessageList, parentMessageId string) string {
	var orderedMessages []userMessage
	var currentMessageId = parentMessageId
	for currentMessageId != "" {
		m := messages.Find(currentMessageId)
		if m == nil {
			break
		}
		orderedMessages = append([]userMessage{*m}, orderedMessages...)
		currentMessageId = m.parentMessageId
	}

	currentDateString := time.Now().Format("2006-01-02")
	promptPrefix := fmt.Sprintf(`你是 ChatGPT，OpenAI 训练的大型语言模型。 您对每个回复都尽可能简洁地回答（例如，不要冗长）。 尽可能简洁地回答是非常重要的，所以请记住这一点。 如果要生成列表，则不要有太多项目。 保持项目数量简短。
Current date: %s\n\n`, currentDateString)
	promptSuffix := "\n"
	currentTokenCount := getTokenCount(promptPrefix + promptSuffix)
	promptBody := ""
	maxTokenCount := 3097

	for currentTokenCount < maxTokenCount && len(orderedMessages) > 0 {
		m := orderedMessages[len(orderedMessages)-1]
		orderedMessages = append([]userMessage{}, orderedMessages[:len(orderedMessages)-1]...)
		messageString := fmt.Sprintf(`%s<|im_end|>\n`, m.message)
		newPromptBody := messageString + promptBody
		newTokenCount := getTokenCount(promptPrefix + newPromptBody + promptSuffix)
		if promptBody != "" && newTokenCount > maxTokenCount {
			break
		}
		promptBody = newPromptBody
		currentTokenCount = newTokenCount
	}

	var prompt = promptPrefix + promptBody + promptSuffix

	var numTokens = getTokenCount(prompt)

	gpt.opt.MaxTokens = int(math.Min(4097-float64(numTokens), 1000))
	return prompt
}

type CompletionRequest struct {
	Model            string         `json:"model"`
	Prompt           string         `json:"prompt,omitempty"`
	Suffix           string         `json:"suffix,omitempty"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	Temperature      float32        `json:"temperature,omitempty"`
	TopP             float32        `json:"top_p,omitempty"`
	N                int            `json:"n,omitempty"`
	Stream           bool           `json:"stream,omitempty"`
	LogProbs         int            `json:"logprobs,omitempty"`
	Echo             bool           `json:"echo,omitempty"`
	Stop             []string       `json:"stop,omitempty"`
	PresencePenalty  float32        `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32        `json:"frequency_penalty,omitempty"`
	BestOf           int            `json:"best_of,omitempty"`
	LogitBias        map[string]int `json:"logit_bias,omitempty"`
	User             string         `json:"user,omitempty"`
}

func (gpt *ChatGPTClient) getCompletion(prompt string) (string, error) {
	var input = gpt.opt
	input.Prompt = prompt
	marshal, _ := json.Marshal(&input)
	request, _ := http.NewRequest("POST", "https://api.openai.com/v1/completions", bytes.NewReader(marshal))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+gpt.apiKey)
	do, err := (&http.Client{Timeout: 60 * time.Second}).Do(request)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	var data GptResponse
	json.NewDecoder(do.Body).Decode(&data)
	return data.Choices[0].Text, nil
}
