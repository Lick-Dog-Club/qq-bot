package ai

import (
	"context"
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/features/util/proxy"
	"qq/features/util/retry"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

var (
	manager = newGptManager[*chatGPTClient](func(uid string) userImp {
		return newChatGPTClient(uid)
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
		bot.Send("已设置 ai_mode: " + m)
		return nil
	}, features.WithHidden())
	features.AddKeyword("ap", "ai 切换 browser 代理", func(bot bot.Bot, content string) error {
		var p = config.AIProxyOne
		if config.AiProxyUrl() == p {
			p = config.AIProxyTwo
		}
		config.Set(map[string]string{"ai_browser_proxy_url": p})
		bot.Send(fmt.Sprintf("已设置: %s", p))
		return nil
	}, features.WithHidden())
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
	newFn func(userID string) userImp
}

func newGptManager[T userImp](newFn func(uid string) userImp) *gptManager[T] {
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
		client = m.newFn(userID)
		m.users[userID] = client
	}
	return client
}

type chatGPTClient struct {
	uid    string
	cache  *keyValue
	status *status
}

func newChatGPTClient(uid string) *chatGPTClient {
	return &chatGPTClient{
		uid:    uid,
		cache:  newKV(map[string]any{"namespace": "chatgpt"}),
		status: &status{},
	}
}

func (gpt *chatGPTClient) lastAskTime() time.Time {
	return gpt.status.LastAskTime()
}

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
		role:            openai.ChatMessageRoleUser,
		message:         msg,
	}
	conversation = append(conversation, um)
	prompt := gpt.buildPrompt(conversation, um.id)
	log.Printf("###########\n%s\n%s", gpt.uid, prompt)
	var result string
	err := retry.Times(5, func() error {
		var err error
		result, err = gpt.getCompletion(prompt)
		return err
	})
	for strings.HasPrefix(result, "\n") {
		result = strings.TrimPrefix(result, "\n")
	}
	if err != nil {
		gpt.status.Asked()
		return err.Error()
	}
	reply := userMessage{
		id:              uuid.NewString(),
		parentMessageId: um.id,
		role:            openai.ChatMessageRoleAssistant,
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

func (gpt *chatGPTClient) buildPrompt(messages userMessageList, parentMessageId string) (res []openai.ChatCompletionMessage) {
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
	for _, message := range orderedMessages {
		res = append(res, openai.ChatCompletionMessage{
			Role:    message.role,
			Content: message.message,
		})
	}

	return
}

func (gpt *chatGPTClient) getCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 800,
		Messages:  messages,
		Stream:    false,
	}
	cfg := openai.DefaultConfig(config.AiToken())
	cfg.HTTPClient = proxy.NewHttpProxyClient()
	c := openai.NewClientWithConfig(cfg)
	timeout, cancelFunc := context.WithTimeout(context.TODO(), 150*time.Second)
	defer cancelFunc()
	stream, err := c.CreateChatCompletion(timeout, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return "", err
	}
	return stream.Choices[0].Message.Content, nil
}
