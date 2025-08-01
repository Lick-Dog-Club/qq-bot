package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"qq/config"
	"qq/features/ai/api/client"
	"qq/features/ai/api/types"
	"qq/features/stock/ai"
	"qq/util/retry"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

const pro = `今天是 {{.Today}}, 当前的 UID 是: "{{.UID}}", 是否来自群聊: "{{.FromGroup}}", 群组 ID: "{{.GroupID}}, From: "{{.From}}"
{{- if .OnlySearch}}
你是一个ai机器人, 你可以使用网络搜索答案，操作步骤为:

  1. 使用 google_search 搜索页面
  2. 选取合适的页面，使用 mclick 点击查看页面内容
  3. 如果页面内容没有相关信息，则继续获取下一页
  4. 如果需要登录，则跳过此页内容，点击下一页，直到获取到符合条件的内容

页面需要一个一个访问，必须获取到符合的内容，至少获取五个，跳过 "www.zhihu.com" 等需要登录的页面
{{ else}}
你是一个ai机器人，如果你不知道答案，那么你需要去网络上搜索之后再回答用户, 你的回答必须满足下面的政策, 不要使用 markdown 格式返回:

- 使用网络搜索结果的步骤为, google_search->mclick->回答用户问题, 必须获取网页内容之后再回答
  1. 你需要调用 "google_search" 方法, 并且传入 "query" 和 "recency_days" 参数, "query" 输入用户问题的详细内容，确保搜索更加精确
  2. "mclick" 获取网页内容
  3. 根据内容回答用户问题

{{ if eq .From "QQ" }}
- 如果返回的是图片地址，你必须使用 "[CQ:image,file={imageURL}]" 这个格式返回, query 的参数也要完整的返回
	例如:
	  imageURL=https://xxx/img.jpg
	  你需要返回: [CQ:image,file=https://xxx/img.jpg]
	例如:
	  imageURL=https://xxx/img.jpg?type=png&uuid=xxx
	  你需要返回: [CQ:image,file=https://xxx/img.jpg?type=png&uuid=xxx]

- 使用 "draw" 画图时，如果要求你画一张类似的图片，你需要把图片详细描述传入到参数中，并且需要有创造力，同时确保有效的互动，当用户请求不明确时，它将根据经验做出推测来解释用户的请求，专注于高效地生成与用户指令相符的图像。

- 如果你要返回的是图片的本地路径，请严格按照如下格式返回：
    [CQ:image,file=file://{imagePath}]
    其中，{imagePath} 必须替换为实际的图片本地绝对路径。例如：
    假设图片路径为 /tmp/example.png，你应该返回：
    [CQ:image,file=file:///tmp/example.png]
    不要直接返回模板或示例内容，必须用真实的图片路径进行替换。
    只返回格式化后的内容，不要附加其他说明或多余信息。
{{- end }}

- 如果用户希望你给他一张图片, 按照以下优先级给图片
  - 优先返回 pixiv 的图片
  - 其次返回动漫图片

- 如果用户参与了 bilibili 抽奖，你需要告诉用户具体抽的奖项内容

- 任务或者提醒相关的提问，请你挑选 “canceltask” “listtask” “task” 方法, 并且执行完后要告诉用户怎么取消这个任务或者提醒 

- 用户查询高铁火车票信息, 按照以下步骤处理
	- 没告诉你时间就那么默认是今天, 需要告诉用户今天是什么日期
	- 调用 "Search12306" 查询班次信息
	- 一等座、二等座和无座已售罄的车次无需告诉用户，重点关注二等座，二等座优先级最高, 如果二等座都卖完了，可以告诉用户其他可选的班次
	- 已经发车的班次不需要告诉用户, 只需要告诉用户可以买哪些班次

- 如果问你，股票相关的问题，你需要化身为短线炒股专家，拥有丰富的炒股经验，请你从多个方面分析股票适不适合短线投资, 时间范围是距今(包括今天)近一个月或三个月的数据，如果用户给的是股票名称，那么先转化成股票代码，如果有多个股票代码，需要先询问用户使用哪个
	## 你需要从以下角度逐个分析
	
	1. 技术分析，例如多个技术指标（如RSI、MACD）给出超卖信号且股价接近支撑位，可能是抄底的机会。
	2. 市场情绪分析，例如极度悲观的情绪往往预示着潜在的反弹机会，但需要结合其他因素综合判断。
	3. 成交量分析，例如在重要支撑位附近，成交量突然增加，表明可能有买盘进入。
	4. 最新市场动态，例如没有重大负面新闻或公告影响股票基本面，短期内的价格下跌可能仅仅是市场情绪的反应。
	5. 短期价格动态
	6. 给出止盈止损的点位，给出具体的数值，并且给这只股票打分(0-100分)，分数越高越适合投资
    
    所有你给出的结论都需要有真实的数据作为支撑！

- 茅台申购步骤，如果用户没给出手机号，那么先询问用户手机号
   1. 通过用户给的手机号自动预约茅台
   2. 如果“用户未登陆，短信已发送”，那么需要询问用户6位短信验证码，添加用户
   3. 添加用户成功之后再次询问用户是否需要进行申购
   4. 返回申购结果详情
{{- end}}
`

var systemPrompt, _ = template.New("").Parse(pro)

var (
	manager = newGptManager[*chatGPTClient](func(uid, from string) userImp {
		return newChatGPTClient(uid, from)
	})
)

type userImp interface {
	lastAskTime() time.Time
	send(s, uid, gid string, send func(msg string) string) string
}

func Request(uuid string, ask, from, uid, gid string, send func(msg string) string) string {
	user := manager.getByUser(uuid, from)
	if user.lastAskTime().Add(10 * time.Minute).Before(time.Now()) {
		manager.deleteUser(uuid)
		user = manager.getByUser(uuid, from)
	}
	result := user.send(ask, uid, gid, send)
	log.Printf("%s: %s\ngpt: %s\n", uuid, ask, result)
	return result
}

func Clear(userID string) {
	manager.deleteUser(userID)
}

type gptManager[T userImp] struct {
	sync.RWMutex
	users map[string]userImp
	newFn func(userID, from string) userImp
}

func newGptManager[T userImp](newFn func(uid, from string) userImp) *gptManager[T] {
	return &gptManager[T]{users: map[string]userImp{}, newFn: newFn}
}

func (m *gptManager[T]) deleteUser(userID string) {
	m.Lock()
	defer m.Unlock()
	delete(m.users, userID)
}

func (m *gptManager[T]) getByUser(userID, from string) userImp {
	m.Lock()
	defer m.Unlock()
	client, ok := m.users[userID]
	if !ok {
		client = m.newFn(userID, from)
		m.users[userID] = client
	}
	return client
}

type chatGPTClient struct {
	uid     string
	status  *types.Status
	history *ai.History

	client func() types.GptClientImpl
	from   string
}

func newChatGPTClient(uid, from string) *chatGPTClient {
	return &chatGPTClient{
		uid:     uid,
		status:  &types.Status{},
		history: &ai.History{},
		client: func() types.GptClientImpl {
			return client.NewOpenaiClientV2(config.AiToken(), config.ChatGPTApiModel(), openai.ChatCompletionRequest{
				Temperature:     0.8,
				PresencePenalty: 1,
				TopP:            1,
			})
		},
		from: from,
	}
}

func (gpt *chatGPTClient) lastAskTime() time.Time {
	return gpt.status.LastAskTime()
}

type SysPrompt struct {
	From       string
	Today      time.Time
	UserID     string
	GroupID    string
	OnlySearch bool
}

func buildSysPrompt(s SysPrompt) string {
	bf := bytes.Buffer{}
	systemPrompt.Execute(&bf, map[string]any{
		"From":       s.From,
		"Today":      s.Today.Format("2006-01-02"),
		"UID":        s.UserID,
		"FromGroup":  s.GroupID != "",
		"GroupID":    s.GroupID,
		"OnlySearch": s.OnlySearch,
	})
	return bf.String()
}

func (gpt *chatGPTClient) send(msg string, userid, gid string, send func(msg string) string) string {
	if gpt.status.IsAsking() {
		return "正在回答上一个问题: " + gpt.status.Msg()
	}
	gpt.status.Asking()
	gpt.status.SetMsg(msg)
	sysPrompt := fmt.Sprintf(`当前时间是：%s
%s
`, time.Now().Format(time.DateTime), buildSysPrompt(SysPrompt{
		From:       gpt.from,
		Today:      time.Now(),
		UserID:     userid,
		GroupID:    gid,
		OnlySearch: config.GPTOnlySearch(),
	}))
	gpt.history.SetSysPrompt(sysPrompt)
	var result string
	err := retry.Times(3, func() error {
		var err error
		result, err = gpt.client().GetCompletion(gpt.history, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: msg,
		}, send)
		if err != nil {
			fmt.Println(err)
		}
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
	gpt.history.Add(openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: result,
	})
	gpt.status.Asked()
	resu := gpt.history.Messages()
	indent, _ := json.MarshalIndent(resu, "", "  ")
	fmt.Println(string(indent))

	return result
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
