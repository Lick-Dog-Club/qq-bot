package ai

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type completionRequest struct {
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

type gptResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

type sendOpts struct {
	ConversationId  string
	ParentMessageId string
}

type userMessage struct {
	id              string
	parentMessageId string
	role            string
	message         string
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

type keyValue struct {
	kv map[string]any
	sync.RWMutex
}

func newKV(kv map[string]any) *keyValue {
	return &keyValue{kv: kv}
}

func (kv *keyValue) Get(k string) any {
	kv.RLock()
	defer kv.RUnlock()
	return kv.kv[k]
}

func (kv *keyValue) Set(k string, v any) {
	kv.Lock()
	defer kv.Unlock()
	kv.kv[k] = v
}

type status struct {
	sync.RWMutex
	isAsking    bool
	opts        *sendOpts
	msg         string
	lastAskTime time.Time
}

func (s *status) GetOpts(noConversationId bool) *sendOpts {
	s.RLock()
	defer s.RUnlock()
	if s.opts == nil {
		s.opts = &sendOpts{
			ParentMessageId: uuid.NewString(),
		}
		if !noConversationId {
			s.opts.ConversationId = uuid.NewString()
		}
	}
	return s.opts
}

func (s *status) SetOpts(opts *sendOpts) {
	s.Lock()
	defer s.Unlock()
	s.opts = &sendOpts{
		ConversationId:  opts.ConversationId,
		ParentMessageId: opts.ParentMessageId,
	}
}

func (s *status) IsAsking() bool {
	s.RLock()
	defer s.RUnlock()
	return s.isAsking
}

func (s *status) LastAskTime() time.Time {
	s.RLock()
	defer s.RUnlock()
	return s.lastAskTime
}

func (s *status) Msg() string {
	s.RLock()
	defer s.RUnlock()
	return s.msg
}

func (s *status) SetMsg(msg string) {
	s.Lock()
	defer s.Unlock()
	s.msg = msg
}
func (s *status) Asked() {
	s.Lock()
	defer s.Unlock()
	s.isAsking = false
}

func (s *status) Asking() {
	s.Lock()
	defer s.Unlock()
	s.isAsking = true
	s.lastAskTime = time.Now()
}
