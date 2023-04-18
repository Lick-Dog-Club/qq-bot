package ai

import (
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/features/ai/api"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type userImp interface {
	lastAskTime() time.Time
	send(string) string
}

var modes = map[string]string{
	"browser": "azure",
	"azure":   "chatgpt",
	"chatgpt": "browser",
}

func init() {
	features.AddKeyword("as", "ai  转换模式 chatgpt/chatgpt-browser/azure", func(bot bot.Bot, content string) error {
		var m string = modes[config.AiMode()]
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
		req := api.Request
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
