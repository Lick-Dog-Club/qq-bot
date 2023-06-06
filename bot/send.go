package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/eatmoreapple/openwechat"

	log "github.com/sirupsen/logrus"
)

type botimp interface {
	DeleteMsg(msgID string)
	SendGroup(gid string, s string) string
	SendToUser(uid string, s string) string
}
type Bot interface {
	botimp
	UserID() string
	IsGroupMessage() bool
	Send(msg string) string
	Message() *Message
}

type CronBot interface {
	botimp
}

type dummyBot struct{}

func NewDummyBot() Bot {
	return &dummyBot{}
}

func (d *dummyBot) Message() *Message {
	return &Message{
		WeReply: func(content string) (*openwechat.SentMessage, error) {
			log.Println("weReply: ", content)
			return nil, nil
		},
		WeSendImg: func(file io.Reader) (*openwechat.SentMessage, error) {
			log.Println("WeSendImg")
			return nil, nil
		},
	}
}

func (d *dummyBot) UserID() string {
	return ""
}

func (d *dummyBot) DeleteMsg(msgID string) {
	fmt.Printf("delete %s", msgID)
}

func (d *dummyBot) Send(msg string) string {
	fmt.Printf("Send:\n%s", msg)
	return ""
}

func (d *dummyBot) SendGroup(gid string, s string) string {
	fmt.Printf("Send:\ngid:%v\ncontent: %s", gid, s)
	return ""
}

func (d *dummyBot) SendToUser(uid string, s string) string {
	fmt.Printf("Send:\nuid:%v\ncontent: %s", uid, s)
	return ""
}

func (d *dummyBot) IsGroupMessage() bool {
	return false
}

type qqBot struct {
	msg *Message
}

func NewQQBot(msg *Message) Bot {
	return &qqBot{msg: msg}
}

func (m *qqBot) Message() *Message {
	return m.msg
}

func (m *qqBot) UserID() string {
	return m.msg.SenderUserID
}

func (m *qqBot) IsGroupMessage() bool {
	return m.msg.IsSendByGroup
}

func (m *qqBot) DeleteMsg(msgID string) {
	deleteMsg(msgID)
}

func (m *qqBot) Send(msg string) string {
	fmt.Println("send: ", msg)
	return send(m.msg, msg)
}

func (m *qqBot) SendGroup(gid string, s string) string {
	return send(&Message{GroupID: gid}, s)
}

func (m *qqBot) SendToUser(uid string, s string) string {
	return send(&Message{SenderUserID: uid}, s)
}

const cqHost = "http://127.0.0.1:5700"

var c = http.Client{}

type QQMessage struct {
	PostType      string `json:"post_type"`
	MetaEventType string `json:"meta_event_type"` // heartbeat
	MessageType   string `json:"message_type"`
	Time          int    `json:"time"`
	SelfID        int64  `json:"self_id"`
	SubType       string `json:"sub_type"`
	Font          int    `json:"font"`
	GroupID       int    `json:"group_id"`
	MessageSeq    int    `json:"message_seq"`
	RawMessage    string `json:"raw_message"`
	Anonymous     *anonymous
	Message       string `json:"message"`
	sender
}

type WechatMessage = openwechat.Message

type Message struct {
	SenderUserID  string
	Message       string
	IsSendByGroup bool
	GroupID       string

	// For WeChat
	WeReply   func(content string) (*openwechat.SentMessage, error)
	WeSendImg func(file io.Reader) (*openwechat.SentMessage, error)
}

/*
	{
	    "post_type":"message",
		"meta_event_type": "",
	    "message_type":"group",
	    "time":1670989954,
	    "self_id":2977648921,
	    "sub_type":"normal",
	    "font":0,
	    "group_id":656599174,
	    "message_seq":107685,
	    "raw_message":"先测试返回的东西",
	    "anonymous":null,
	    "message":"先测试下拿下返回的东西",
	    "sender":{
	        "age":0,
	        "area":"",
	        "card":"韭菜",
	        "level":"",
	        "nickname":"杰森跟班",
	        "role":"admin",
	        "sex":"unknown",
	        "title":"",
	        "user_id":1025434218,
	        "message_id":575298704
	    }
	}
*/
type sender struct {
	Age       int    `json:"age"`
	Area      string `json:"area"`
	Card      string `json:"card"`
	Level     string `json:"level"`
	Nickname  string `json:"nickname"`
	Role      string `json:"role"`
	Sex       string `json:"sex"`
	Title     string `json:"title"`
	UserID    int    `json:"user_id"`
	MessageID int    `json:"message_id"`
}

type anonymous struct {
	ID   int64  `json:"id"`   //匿名用户 ID
	Name string `json:"name"` //匿名用户名称
	Flag string `json:"flag"` //匿名用户 flag, 在调用禁言 API 时需要传入
}

type sendResponse struct {
	Data struct {
		MessageID int `json:"message_id"`
	} `json:"data"`
}

func deleteMsg(msgID string) {
	req, _ := http.NewRequest("POST", cqHost+"/delete_msg", strings.NewReader(fmt.Sprintf(`{"message_id": %s}`, msgID)))
	req.Header.Add("content-type", "application/json")
	do, _ := c.Do(req)
	defer do.Body.Close()
}

func send(message *Message, msg string) string {
	log.Println(message.GroupID, message.SenderUserID, "message.GroupID, message.SenderUserID")
	if message.GroupID == "" && message.SenderUserID == "" {
		log.Println("GroupID == 0, UserID == 0")
		return ""
	}
	var req *http.Request
	if message.GroupID != "" {
		req, _ = http.NewRequest("POST", cqHost+"/send_group_msg", strings.NewReader(fmt.Sprintf(`{"group_id": %s, "message": %q}`, message.GroupID, strings.Trim(msg, "\n"))))
	} else {
		req, _ = http.NewRequest("POST", cqHost+"/send_msg", strings.NewReader(fmt.Sprintf(`{"user_id": %s, "message": %q}`, message.SenderUserID, strings.Trim(msg, "\n"))))
	}
	req.Header.Add("content-type", "application/json")
	do, _ := c.Do(req)
	defer do.Body.Close()
	var res sendResponse
	json.NewDecoder(do.Body).Decode(&res)
	return fmt.Sprintf("%d", res.Data.MessageID)
}

type wechatBot struct {
	message Message
}

type WeMsgMap struct {
	sync.RWMutex
	m map[string]*openwechat.SentMessage
}

func (w *WeMsgMap) Get(id string) *openwechat.SentMessage {
	w.RLock()
	defer w.RUnlock()
	return w.m[id]
}

func (w *WeMsgMap) Add(id string, text *openwechat.SentMessage) {
	w.Lock()
	defer w.Unlock()
	w.m[id] = text
}

func (w *WeMsgMap) Delete(id string) {
	w.Lock()
	defer w.Unlock()
	log.Println("WeMsgMap delete ", id)
	delete(w.m, id)
}

var WeMessageMap = &WeMsgMap{m: map[string]*openwechat.SentMessage{}}

func NewWechatBot(msg Message) Bot {
	return &wechatBot{message: msg}
}

func (w *wechatBot) DeleteMsg(msgID string) {
	log.Println("DeleteMsgID: ", msgID)
	if message := WeMessageMap.Get(msgID); message != nil {
		if message.CanRevoke() {
			message.Revoke()
		}
		WeMessageMap.Delete(msgID)
	}
}

func (w *wechatBot) SendGroup(gid string, s string) string {
	return ""
}

func (w *wechatBot) SendToUser(uid string, s string) string {
	return ""
}

func (w *wechatBot) UserID() string {
	return fmt.Sprintf("%v", w.message.SenderUserID)
}

func (w *wechatBot) IsGroupMessage() bool {
	return w.message.IsSendByGroup
}

func (w *wechatBot) Send(msg string) string {
	text, err := w.message.WeReply(msg)
	if err != nil {
		log.Println(err)
	}
	//WeMessageMap.Add(text.MsgId, text)
	log.Println("text.MsgId: ", text)
	return text.MsgId
}

func (w *wechatBot) Message() *Message {
	return &w.message
}
