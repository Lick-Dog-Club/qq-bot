package types

import (
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/google/uuid"
)

type GptClientImpl interface {
	GetCompletion(messages []openai.ChatCompletionMessage) (string, error)
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

type KeyValue struct {
	kv map[string]any
	sync.RWMutex
}

func NewKV(kv map[string]any) *KeyValue {
	return &KeyValue{kv: kv}
}

func (kv *KeyValue) Get(k string) any {
	kv.RLock()
	defer kv.RUnlock()
	return kv.kv[k]
}

func (kv *KeyValue) Set(k string, v any) {
	kv.Lock()
	defer kv.Unlock()
	kv.kv[k] = v
}

type Status struct {
	sync.RWMutex
	isAsking    bool
	opts        *SendOpts
	msg         string
	lastAskTime time.Time
}

func (s *Status) GetOpts(noConversationId bool) *SendOpts {
	s.RLock()
	defer s.RUnlock()
	if s.opts == nil {
		s.opts = &SendOpts{
			ParentMessageId: uuid.NewString(),
		}
		if !noConversationId {
			s.opts.ConversationId = uuid.NewString()
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
