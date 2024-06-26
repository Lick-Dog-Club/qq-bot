package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"qq/config"
	"qq/util/random"
	"qq/util/retry"
	"qq/util/text2png"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eatmoreapple/openwechat"

	log "github.com/sirupsen/logrus"
)

type botimp interface {
	DeleteMsg(msgID string)
	SendGroup(gid string, s string) string
	SendToUser(uid string, s string) string
	SendTextImageToUser(uid string, text string) (string, error)
	SendTextImageToGroup(gid string, text string) (string, error)
}
type Bot interface {
	botimp
	UserID() string
	From() string
	IsFromAdmin() bool
	IsGroupMessage() bool
	Send(msg string) string
	SendTextImage(text string) (string, error)
	Message() *Message
	GroupID() string
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
func (d *dummyBot) GroupID() string {
	return ""
}

func (d *dummyBot) From() string {
	return "dummy"
}

func (d *dummyBot) IsFromAdmin() bool {
	return true
}

func (d *dummyBot) DeleteMsg(msgID string) {
	fmt.Printf("delete %s", msgID)
}

func (d *dummyBot) Send(msg string) string {
	fmt.Printf("Send:\n%s", msg)
	return ""
}

func (d *dummyBot) SendTextImage(text string) (string, error) {
	fmt.Printf("Send:\n%s", text)
	return "", nil
}

func (d *dummyBot) SendTextImageToUser(uid string, text string) (string, error) {
	fmt.Printf("uid: %v, text: %v", uid, text)
	return "", nil
}

func (d *dummyBot) SendTextImageToGroup(gid string, text string) (string, error) {
	fmt.Printf("gid: %v, text: %v", gid, text)
	return "", nil
}

func (d *dummyBot) SendGroup(gid string, s string) string {
	fmt.Printf("Send:\ngid:%v\ncontent: %s", gid, s)
	return ""
}

func (d *dummyBot) SendToUser(uid string, s string) string {
	fmt.Printf("Send:\nuid:%v\ncontent: %s\n", uid, s)
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

func (m *qqBot) GroupID() string {
	return m.msg.GroupID
}

func (m *qqBot) From() string {
	return "QQ"
}

func (m *qqBot) IsGroupMessage() bool {
	return m.msg.IsSendByGroup
}

func (m *qqBot) IsFromAdmin() bool {
	return config.AdminIDs().Contains(m.UserID())
}

func (m *qqBot) DeleteMsg(msgID string) {
	deleteMsg(msgID)
}

func (m *qqBot) Send(msg string) string {
	return send(m.msg, msg)
}

func (m *qqBot) SendTextImageToUser(uid string, text string) (string, error) {
	return m.sendImage(&Message{SenderUserID: uid}, text)
}

func (m *qqBot) SendTextImageToGroup(gid string, text string) (string, error) {
	return m.sendImage(&Message{GroupID: gid}, text)
}

func (m *qqBot) SendTextImage(text string) (string, error) {
	return m.sendImage(m.msg, text)
}

func (m *qqBot) sendImage(msg *Message, text string) (string, error) {
	if text == "" {
		m.Send("无数据: " + text)
		fmt.Println(text, errors.New("empty text"))
		return "", errors.New("empty text")
	}
	path := tmpPath()
	if err := text2png.Draw([]string{text}, path); err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println(path)
	defer func() {
		os.Remove(path)
	}()
	send(msg, fmt.Sprintf("[CQ:image,file=file://%s]", path))
	return path, nil
}

func tmpPath() string {
	return filepath.Join(config.ImageDir, fmt.Sprintf("tmp-%s-%s.png", time.Now().Format("2006-01-02"), random.String(10)))
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

func toInt(s string) int {
	atoi, _ := strconv.Atoi(strings.TrimSpace(s))
	return atoi
}

func send(message *Message, msg string) string {
	var gid int = toInt(message.GroupID)
	var sid int = toInt(message.SenderUserID)
	if gid == 0 && sid == 0 {
		log.Println("GroupID == 0, UserID == 0")
		return ""
	}
	var req *http.Request
	if gid > 0 {
		log.Println("send to group: ", gid)
		req, _ = http.NewRequest("POST", cqHost+"/send_group_msg", strings.NewReader(fmt.Sprintf(`{"group_id": %d, "message": %q}`, gid, strings.Trim(msg, "\n"))))
	} else {
		log.Println("send to user: ", message.SenderUserID)
		req, _ = http.NewRequest("POST", cqHost+"/send_msg", strings.NewReader(fmt.Sprintf(`{"user_id": %d, "message": %q}`, sid, strings.Trim(msg, "\n"))))
	}
	req.Header.Add("content-type", "application/json")
	do, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer do.Body.Close()
	all, _ := io.ReadAll(do.Body)
	fmt.Println(string(all))
	var res sendResponse
	json.NewDecoder(bytes.NewReader(all)).Decode(&res)
	return fmt.Sprintf("%d", res.Data.MessageID)
}

type wechatBot struct {
	message Message
	msgMap  *WeMsgMap
}

func GetCQImage(file string) string {
	req, _ := http.NewRequest("POST", cqHost+"/get_image", strings.NewReader(fmt.Sprintf(`{"file": %q}`, file)))
	req.Header.Add("content-type", "application/json")
	do, _ := c.Do(req)
	defer do.Body.Close()
	all, _ := io.ReadAll(do.Body)
	fmt.Println(string(all))
	var res imageResponse
	json.NewDecoder(bytes.NewReader(all)).Decode(&res)
	return res.Data.Url
}

type imageResponse struct {
	Data struct {
		Size     int    `json:"size"`
		Filename string `json:"filename"`
		Url      string `json:"url"`
	} `json:"data"`
}

type WeMsgMap struct {
	sync.RWMutex
	m map[string]*openwechat.SentMessage
}

func NewWeMsgMap() *WeMsgMap {
	return &WeMsgMap{m: map[string]*openwechat.SentMessage{}}
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

func NewWechatBot(msg Message, msgMap *WeMsgMap) Bot {
	return &wechatBot{message: msg, msgMap: msgMap}
}

func (w *wechatBot) DeleteMsg(msgID string) {
	log.Println("DeleteMsgID: ", msgID)
	if message := w.msgMap.Get(msgID); message != nil {
		if message.CanRevoke() {
			message.Revoke()
		}
		w.msgMap.Delete(msgID)
	}
}

func (w *wechatBot) SendGroup(gid string, s string) string {
	return ""
}

func (w *wechatBot) SendToUser(uid string, s string) string {
	return ""
}

func (w *wechatBot) UserID() string {
	return w.message.SenderUserID
}

func (w *wechatBot) GroupID() string {
	return w.message.GroupID
}

func (w *wechatBot) From() string {
	return "Wechat"
}

func (w *wechatBot) IsFromAdmin() bool {
	return config.AdminIDs().Contains(w.UserID())
}

func (w *wechatBot) IsGroupMessage() bool {
	return w.message.IsSendByGroup
}

func (w *wechatBot) SendTextImage(text string) (string, error) {
	path := tmpPath()
	if err := text2png.Draw([]string{text}, path); err != nil {
		return "", err
	}
	open, _ := os.Open(path)
	defer open.Close()
	_, err := w.message.WeSendImg(open)
	return path, err
}

func (w *wechatBot) SendTextImageToUser(uid string, text string) (string, error) {
	return "", errors.New("不支持")
}

func (w *wechatBot) SendTextImageToGroup(gid string, text string) (string, error) {
	return "", errors.New("不支持")
}

func (w *wechatBot) Send(msg string) string {
	var text *openwechat.SentMessage
	var err error
	if err := retry.Times(3, func() error {
		text, err = w.message.WeReply(msg)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}); err != nil {
		return ""
	}
	//WeMessageMap.Add(text.MsgId, text)
	log.Println("text.MsgId: ", text)
	return text.MsgId
}

func (w *wechatBot) Message() *Message {
	return &w.message
}
