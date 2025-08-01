package types

import (
	"qq/features/stock/ai"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

type GptClientImpl interface {
	GetCompletion(his *ai.History, current openai.ChatCompletionMessage, send func(msg string) string) (string, error)
	Platform() string
}

type SendOpts struct {
	ConversationId  string
	ParentMessageId string
}

type UserMessage struct {
	ID              string
	ParentMessageId string
	Role            string
	Message         string
}

type UserMessageList []UserMessage

func (l UserMessageList) Find(id string) *UserMessage {
	for _, message := range l {
		if message.ID == id {
			return &message
		}
	}
	return nil
}

type Status struct {
	sync.RWMutex
	isAsking    bool
	msg         string
	lastAskTime time.Time
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
