package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"qq/config"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var bmanager = newGptManager[*browserChatGPTClient](func() userImp {
	return newBrowserChatGPTClient()
})

type browserChatGPTClient struct {
	cache  *keyValue
	status *status
}

func BrowserRequest(userID string, ask string) string {
	user := bmanager.getByUser(userID)
	if user.lastAskTime().Add(10 * time.Minute).Before(time.Now()) {
		bmanager.deleteUser(userID)
		user = bmanager.getByUser(userID)
	}
	return user.send(ask)
}

func newBrowserChatGPTClient() *browserChatGPTClient {
	return &browserChatGPTClient{
		cache:  newKV(map[string]any{"namespace": "chatgpt-browser"}),
		status: &status{},
	}
}

func (gpt *browserChatGPTClient) lastAskTime() time.Time {
	return gpt.status.LastAskTime()
}

type browserUserMessage struct {
	action          string
	id              string
	parentMessageId string
	role            string
	message         string
}

func (gpt *browserChatGPTClient) send(msg string) string {
	if gpt.status.IsAsking() {
		return "正在回答上一个问题: " + gpt.status.Msg()
	}
	gpt.status.Asking()
	gpt.status.SetMsg(msg)
	var opts *sendOpts = gpt.status.GetOpts(true)

	var conversation []browserUserMessage

	if opts.ConversationId != "" {
		get := gpt.cache.Get(opts.ConversationId)
		if get == nil {
			conversation = []browserUserMessage{}
		} else {
			conversation = get.([]browserUserMessage)
		}
	}

	um := browserUserMessage{
		id:              uuid.NewString(),
		parentMessageId: opts.ParentMessageId,
		role:            "User",
		message:         msg,
	}
	conversation = append(conversation, um)
	resp := gpt.postConversation(browserUserMessage{
		id:              opts.ConversationId,
		parentMessageId: opts.ParentMessageId,
		action:          "next",
		message:         msg,
	})
	if resp == nil {
		gpt.status.Asked()
		return ""
	}

	reply := browserUserMessage{
		id:              resp.Message.ID,
		parentMessageId: um.id,
		role:            "ChatGPT",
		message:         resp.Message.Content.Parts[0],
	}
	opts.ConversationId = resp.ConversationID
	conversation = append(conversation, reply)
	gpt.cache.Set(opts.ConversationId, conversation)
	gpt.status.SetOpts(&sendOpts{
		ConversationId:  opts.ConversationId,
		ParentMessageId: reply.id,
	})
	gpt.status.Asked()

	return reply.message
}

type bmessage struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content struct {
		ContentType string   `json:"content_type"`
		Parts       []string `json:"parts"`
	} `json:"content"`
}

type webInput struct {
	ConversationId  string     `json:"conversation_id,omitempty"`
	Action          string     `json:"action"`
	Messages        []bmessage `json:"messages"`
	ParentMessageID string     `json:"parent_message_id"`
	Model           string     `json:"model"`
}

func (gpt *browserChatGPTClient) postConversation(message browserUserMessage) *response {
	var msgs []bmessage
	if message.message != "" {
		msgs = []bmessage{
			{
				ID:   uuid.NewString(),
				Role: "user",
				Content: struct {
					ContentType string   `json:"content_type"`
					Parts       []string `json:"parts"`
				}{
					ContentType: "text",
					Parts:       []string{message.message},
				},
			},
		}
	}
	var input = webInput{
		ConversationId:  message.id,
		Action:          message.action,
		Messages:        msgs,
		ParentMessageID: message.parentMessageId,
		Model:           config.AiBrowserModel(),
	}
	marshal, _ := json.Marshal(&input)
	log.Println(string(marshal))
	request, _ := http.NewRequest("POST", config.AiProxyUrl(), bytes.NewReader(marshal))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+config.AiAccessToken())
	do, err := (&http.Client{Timeout: 5 * time.Minute}).Do(request)
	if err != nil {
		return nil
	}
	defer do.Body.Close()
	scanner := bufio.NewScanner(do.Body)
	for scanner.Scan() {
		var resp response
		s := strings.TrimPrefix(scanner.Text(), "data: ")
		json.NewDecoder(strings.NewReader(s)).Decode(&resp)
		log.Println(s)
		if resp.Message.EndTurn {
			return &resp
		}
	}
	return nil
}

type response struct {
	Message struct {
		ID         string      `json:"id"`
		Role       string      `json:"role"`
		User       interface{} `json:"user"`
		CreateTime interface{} `json:"create_time"`
		UpdateTime interface{} `json:"update_time"`
		Content    struct {
			ContentType string   `json:"content_type"`
			Parts       []string `json:"parts"`
		} `json:"content"`
		EndTurn  bool    `json:"end_turn"`
		Weight   float64 `json:"weight"`
		Metadata struct {
			FinishDetails struct {
				Type string `json:"type"`
				Stop string `json:"stop"`
			} `json:"finish_details"`
			MessageType string `json:"message_type"`
			ModelSlug   string `json:"model_slug"`
		} `json:"metadata"`
		Recipient string `json:"recipient"`
	} `json:"message"`
	ConversationID string      `json:"conversation_id"`
	Error          interface{} `json:"error"`
}
