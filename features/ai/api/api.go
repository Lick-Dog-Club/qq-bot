package api

import (
	"log"
	"qq/config"
	"qq/features/ai/api/client"
	"qq/features/ai/api/tools"
	"qq/features/ai/api/types"
	"qq/util/retry"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

var (
	manager = newGptManager[*chatGPTClient](func(uid string) userImp {
		switch config.AiMode() {
		case "azure":
			return newAzureClient(uid)
		default:
			return newChatGPTClient(uid)
		}
	})
)

type userImp interface {
	lastAskTime() time.Time
	send(string) string
}

func Request(userID string, ask string) string {
	user := manager.getByUser(userID)
	if user.lastAskTime().Add(10 * time.Minute).Before(time.Now()) {
		manager.deleteUser(userID)
		user = manager.getByUser(userID)
	}
	result := user.send(ask)
	log.Printf("%s: %s\ngpt: %s\n", userID, ask, result)
	return result
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
	cache  *types.KeyValue
	status *types.Status

	client types.GptClientImpl
}

func newChatGPTClient(uid string) *chatGPTClient {
	return &chatGPTClient{
		uid:    uid,
		cache:  types.NewKV(map[string]any{"namespace": "chatgpt"}),
		status: &types.Status{},
		client: client.NewOpenaiClient(config.AiToken(), config.ChatGPTApiModel(), openai.ChatCompletionRequest{
			Temperature:     0.8,
			PresencePenalty: 1,
			TopP:            1,
			Tools:           tools.List(),
		}),
	}
}

func newAzureClient(uid string) *chatGPTClient {
	return &chatGPTClient{
		uid:    uid,
		cache:  types.NewKV(map[string]any{"namespace": "chatgpt"}),
		status: &types.Status{},
		client: client.NewAzureClient(config.AzureToken(), config.AzureModel(), openai.ChatCompletionRequest{
			Temperature:     0.8,
			PresencePenalty: 1,
			TopP:            1,
			Tools:           tools.List(),
		}, config.AzureUrl()),
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
	var opts *types.SendOpts = gpt.status.GetOpts(false)
	var conversation []types.UserMessage
	get := gpt.cache.Get(opts.ConversationId)
	if get == nil {
		conversation = []types.UserMessage{}
	} else {
		conversation = get.([]types.UserMessage)
	}
	um := types.UserMessage{
		ID:              uuid.NewString(),
		ParentMessageId: opts.ParentMessageId,
		Role:            openai.ChatMessageRoleUser,
		Message:         msg,
	}
	conversation = append(conversation, um)
	prompt := gpt.BuildPrompt(conversation, um.ID)
	//log.Printf("###########\n%s%s", gpt.uid, prompt)
	prompt = append([]openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `
你是一个ai机器人，能回答用户的任何问题, 你的回答必须满足下面的格式:
1. 如果返回的是图片地址，你必须使用 "[CQ:image,file={imageURL}]" 这个格式返回
例如:
  imageURL=https://xxx/img.jpg
  你需要返回: [CQ:image,file=https://xxx/img.jpg]
2. 如果返回的是图片本地路径，你必须使用 "[CQ:image,file=file://{imagePath}]" 这个格式返回
例如:
  imagePath=/tmp/a.png
  你需要返回: [CQ:image,file=file:///tmp/a.png]
`,
		},
	}, prompt...)
	var result string
	err := retry.Times(10, func() error {
		var err error
		result, err = gpt.client.GetCompletion(prompt)
		return err
	})
	for strings.HasPrefix(result, "\n") {
		result = strings.TrimPrefix(result, "\n")
	}
	if err != nil {
		gpt.status.Asked()
		log.Println(err.Error())
		return "前方拥挤，请稍后再试~"
	}
	reply := types.UserMessage{
		ID:              uuid.NewString(),
		ParentMessageId: um.ID,
		Role:            openai.ChatMessageRoleAssistant,
		Message:         result,
	}
	conversation = append(conversation, reply)
	gpt.cache.Set(opts.ConversationId, conversation)
	gpt.status.SetOpts(&types.SendOpts{
		ConversationId:  opts.ConversationId,
		ParentMessageId: reply.ID,
	})
	gpt.status.Asked()

	return reply.Message
}

func (gpt *chatGPTClient) BuildPrompt(messages types.UserMessageList, parentMessageId string) (res []openai.ChatCompletionMessage) {
	var orderedMessages []types.UserMessage
	var currentMessageId = parentMessageId
	for currentMessageId != "" {
		m := messages.Find(currentMessageId)
		if m == nil {
			break
		}
		orderedMessages = append([]types.UserMessage{*m}, orderedMessages...)
		currentMessageId = m.ParentMessageId
	}
	for _, message := range orderedMessages {
		res = append(res, openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Message,
		})
	}

	return
}
