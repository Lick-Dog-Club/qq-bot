package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/features/ai/encoder"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

var (
	manager = newGptManager[*chatGPTClient](func() userImp {
		return newChatGPTClient()
	})
)

type userImp interface {
	lastAskTime() time.Time
	send(string) string
}

func init() {
	features.AddKeyword("as", "ai  转换模式 api/browser", func(bot bot.Bot, content string) error {
		var m string
		switch config.AiMode() {
		case "browser":
			m = "api"
		case "api":
			fallthrough
		default:
			m = "browser"
		}
		config.Set(map[string]string{"ai_mode": m})
		return nil
	})
	features.SetDefault("ai 自动回答", func(bot bot.Bot, content string) error {
		req := Request
		if config.AiMode() == "api" && config.AiToken() == "" {
			bot.Send("请先设置变量: ai_token")
			return nil
		}
		if config.AiMode() == "browser" {
			if config.AiAccessToken() == "" {
				bot.Send("请先设置变量: ai_access_token")
				return nil
			}
			req = BrowserRequest
		}
		log.Printf("%s: %s", bot.UserID(), content)
		bot.Send(req(bot.UserID(), content))
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

type gptManager[T userImp] struct {
	sync.RWMutex
	users map[string]userImp
	newFn func() userImp
}

func newGptManager[T userImp](newFn func() userImp) *gptManager[T] {
	return &gptManager[T]{users: map[string]userImp{}, newFn: newFn}
}

func (m *gptManager[T]) deleteUser(userID string) {
	m.Lock()
	defer m.Unlock()
	delete(m.users, userID)
}

func (m *gptManager[T]) getByUser(userID string) userImp {
	m.Lock()
	defer m.Unlock()
	client, ok := m.users[userID]
	if !ok {
		client = m.newFn()
		m.users[userID] = client
	}
	return client
}

type chatGPTClient struct {
	opt    completionRequest
	cache  *keyValue
	status *status
}

func newChatGPTClient() *chatGPTClient {
	return &chatGPTClient{
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
	var opts *sendOpts = gpt.status.GetOpts(false)
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
	log.Printf("###########\n%s", prompt)
	result, err := gpt.getCompletion(prompt)
	if err != nil {
		gpt.status.Asked()
		return err.Error()
	}
	reply := userMessage{
		id:              uuid.NewString(),
		parentMessageId: um.id,
		role:            "ChatGPT",
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
	request.Header.Add("Authorization", "Bearer "+config.AiToken())
	do, err := (&http.Client{Timeout: 3 * time.Minute}).Do(request)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	var data gptResponse
	if err := json.NewDecoder(do.Body).Decode(&data); err != nil {
		return "", err
	}
	var res string = "没有结果"
	if len(data.Choices) > 0 {
		res = data.Choices[0].Text
	}
	return res, nil
}
