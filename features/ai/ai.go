package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/features/ai/encoder"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	token   = config.AIToken
	manager = newGptManager(token)
)

func init() {
	features.SetDefault("ai 自动回答", func(bot bot.Bot, content string) error {
		if token == "" {
			bot.Send("请先设置环境变量: AI_TOKEN")
			return nil
		}
		log.Printf("%s: %s", bot.UserID(), content)
		bot.Send(Request(bot.UserID(), content))
		return nil
	})
}

func Request(userID string, ask string) string {
	user := manager.getByUser(userID)
	if user.lastAskTime().Add(10 * time.Minute).Before(time.Now()) {
		manager.deleteUser(userID)
		user = manager.getByUser(userID)
	}
	return user.send(ask)
}

type gptManager struct {
	sync.RWMutex
	users  map[string]*chatGPTClient
	apiKey string
}

func newGptManager(apiKey string) *gptManager {
	return &gptManager{apiKey: apiKey, users: map[string]*chatGPTClient{}}
}

func (m *gptManager) deleteUser(userID string) {
	m.Lock()
	defer m.Unlock()
	delete(m.users, userID)
}

func (m *gptManager) getByUser(userID string) *chatGPTClient {
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
	apiKey string
	opt    completionRequest
	cache  *keyValue
	status *status
}

func newChatGPTClient(apiKey string) *chatGPTClient {
	return &chatGPTClient{
		apiKey: apiKey,
		opt: completionRequest{
			Model:           "text-davinci-003",
			Temperature:     0.8,
			Stop:            []string{endToken},
			PresencePenalty: 1,
		},
		cache:  newKV(map[string]any{"namespace": "chatgpt"}),
		status: &status{},
	}
}

func (gpt *chatGPTClient) lastAskTime() time.Time {
	return gpt.status.LastAskTime()
}

const (
	endToken       = "<|endoftext|>"
	separatorToken = "<|endoftext|>"
)

func (gpt *chatGPTClient) send(msg string) string {
	if gpt.status.IsAsking() {
		return "正在回答上一个问题: " + gpt.status.Msg()
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
	promptPrefix := fmt.Sprintf("\n%sInstructions: \nYou are ChatGPT, a large language model trained by OpenAI. \nCurrent date: %s%s\n\n", separatorToken, currentDateString, separatorToken)
	promptSuffix := "ChatGPT:\n"
	currentTokenCount := getTokenCount(promptPrefix + promptSuffix)
	promptBody := ""
	maxTokenCount := 3097

	for currentTokenCount < maxTokenCount && len(orderedMessages) > 0 {
		m := orderedMessages[len(orderedMessages)-1]
		roleLabel := "User"
		if m.role != "User" {
			roleLabel = "ChatGPT"
		}
		orderedMessages = append([]userMessage{}, orderedMessages[:len(orderedMessages)-1]...)
		messageString := fmt.Sprintf("%s:\n%s%s\n", roleLabel, m.message, endToken)
		newPromptBody := ""
		newTokenCount := getTokenCount(promptPrefix + newPromptBody + promptSuffix)

		if promptBody != "" {
			newPromptBody = fmt.Sprintf("%s%s", messageString, promptBody)
		} else {
			newPromptBody = fmt.Sprintf("%s%s%s", promptPrefix, messageString, promptBody)
		}

		newTokenCount = getTokenCount(fmt.Sprintf("%s%s%s", promptPrefix, promptBody, promptSuffix))
		if promptBody != "" && newTokenCount > maxTokenCount {
			break
		}
		promptBody = newPromptBody
		currentTokenCount = newTokenCount
	}

	var prompt = promptBody + promptSuffix

	var numTokens = getTokenCount(prompt)

	gpt.opt.MaxTokens = int(math.Min(4097-float64(numTokens), 1000))
	return prompt
}

const (
	imEnd = "<|im_end|>"
	imSep = "<|im_sep|>"
)

func getTokenCount(text string) int {
	encoder, _ := encoder.NewEncoder()
	encode, _ := encoder.Encode(strings.ReplaceAll(strings.ReplaceAll(text, imSep, endToken), imEnd, endToken))
	return len(encode)
}

func (gpt *chatGPTClient) getCompletion(prompt string) (string, error) {
	var input = gpt.opt
	input.Prompt = prompt
	marshal, _ := json.Marshal(&input)
	request, _ := http.NewRequest("POST", "https://api.openai.com/v1/completions", bytes.NewReader(marshal))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+gpt.apiKey)
	do, err := (&http.Client{Timeout: 3 * time.Minute}).Do(request)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	var data gptResponse
	if err := json.NewDecoder(do.Body).Decode(&data); err != nil {
		return "", err
	}
	log.Println(do.Status, data)
	var res string = "没有结果"
	if len(data.Choices) > 0 {
		res = data.Choices[0].Text
	}
	return res, nil
}
