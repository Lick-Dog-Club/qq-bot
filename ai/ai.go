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
	manager = newGptManager(token)
)

func Request(userID int, ask string) string {
	user := manager.GetByUser(userID)
	if user.LastAskTime().Add(10 * time.Minute).Before(time.Now()) {
		manager.DeleteUser(userID)
		user = manager.GetByUser(userID)
	}
	return user.Send(ask)
}

type gptManager struct {
	sync.RWMutex
	users  map[int]*chatGPTClient
	apiKey string
}

func newGptManager(apiKey string) *gptManager {
	return &gptManager{apiKey: apiKey, users: map[int]*chatGPTClient{}}
}

func (m *gptManager) DeleteUser(userID int) {
	m.Lock()
	defer m.Unlock()
	delete(m.users, userID)
}

func (m *gptManager) GetByUser(userID int) *chatGPTClient {
	m.Lock()
	defer m.Unlock()
	client, ok := m.users[userID]
	if !ok {
		client = newChatGPTClient(m.apiKey)
		m.users[userID] = client
	}
	return client
}

type chatGPTClient struct {
	lastAskTime time.Time
	apiKey      string
	opt         completionRequest
	cache       *keyValue
	status      *status
}

func newChatGPTClient(apiKey string) *chatGPTClient {
	return &chatGPTClient{
		apiKey: apiKey,
		opt: completionRequest{
			Model:           "text-chat-davinci-002-20230126",
			Temperature:     0.7,
			Stop:            []string{"<|im_end|>"},
			PresencePenalty: 0.6,
		},
		cache:  newKV(map[string]any{"namespace": "chatgpt"}),
		status: &status{},
	}
}

func (gpt *chatGPTClient) LastAskTime() time.Time {
	return gpt.status.LastAskTime()
}

func (gpt *chatGPTClient) Send(msg string) string {
	if gpt.status.IsAsking() {
		return "正在回答上一个问题~: " + gpt.status.Msg()
	}
	gpt.status.Asking()
	gpt.status.SetMsg(msg)
	var opts *sendOpts = gpt.status.GetOpts()
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
	gpt.status.SetOpts(&sendOpts{
		ConversationId:  opts.ConversationId,
		ParentMessageId: reply.id,
	})
	gpt.status.Asked()

	return reply.message
}

func (gpt *chatGPTClient) buildPrompt(messages userMessageList, parentMessageId string) string {
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

func getTokenCount(text string) int {
	encoder, _ := encoder.NewEncoder()
	encode, _ := encoder.Encode(strings.ReplaceAll(text, `<|im_end|>`, `<|endoftext|>`))
	return len(encode)
}

func (gpt *chatGPTClient) getCompletion(prompt string) (string, error) {
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
	var data gptResponse
	json.NewDecoder(do.Body).Decode(&data)
	return data.Choices[0].Text, nil
}
