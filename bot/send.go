package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type botimp interface {
	DeleteMsg(msgID int)
	SendGroup(gid string, s string) int
	SendToUser(uid string, s string) int
}
type Bot interface {
	botimp
	UserID() string
	IsGroupMessage() bool
	Send(msg string) int
}

type CronBot interface {
	botimp
}

type dummyBot struct {
}

func NewDummyBot(message *Message) Bot {
	return &dummyBot{}
}

func (d *dummyBot) UserID() string {
	return ""
}

func (d *dummyBot) DeleteMsg(msgID int) {
	fmt.Printf("delete %d", msgID)
}

func (d *dummyBot) Send(msg string) int {
	fmt.Printf("Send:\n%s", msg)
	return 0
}

func (d *dummyBot) SendGroup(gid string, s string) int {
	fmt.Printf("Send:\ngid:%v\ncontent: %s", gid, s)
	return 0
}

func (d *dummyBot) SendToUser(uid string, s string) int {
	fmt.Printf("Send:\nuid:%v\ncontent: %s", uid, s)
	return 0
}

func (d *dummyBot) IsGroupMessage() bool {
	return false
}

type bot struct {
	msg *Message
}

func NewBot(msg *Message) Bot {
	return &bot{msg: msg}
}

func (m *bot) UserID() string {
	return fmt.Sprintf("%d", m.msg.UserID)
}

func (m *bot) IsGroupMessage() bool {
	return m.msg.MessageType == "group"
}

func (m *bot) DeleteMsg(msgID int) {
	deleteMsg(msgID)
}

func (m *bot) Send(msg string) int {
	fmt.Println("send: ", msg)
	return send(m.msg, msg)
}

func toInt(s string) int {
	atoi, _ := strconv.Atoi(s)
	return atoi
}

func (m *bot) SendGroup(gid string, s string) int {
	return send(&Message{GroupID: toInt(fmt.Sprintf("%v", gid))}, s)
}

func (m *bot) SendToUser(uid string, s string) int {
	return send(&Message{sender: sender{UserID: toInt(fmt.Sprintf("%v", uid))}}, s)
}

const cqHost = "http://127.0.0.1:5700"

var c = http.Client{}

type Message struct {
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
	    "raw_message":"????????????????????????",
	    "anonymous":null,
	    "message":"?????????????????????????????????",
	    "sender":{
	        "age":0,
	        "area":"",
	        "card":"??????",
	        "level":"",
	        "nickname":"????????????",
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
	ID   int64  `json:"id"`   //???????????? ID
	Name string `json:"name"` //??????????????????
	Flag string `json:"flag"` //???????????? flag, ??????????????? API ???????????????
}

type sendResponse struct {
	Data struct {
		MessageID int `json:"message_id"`
	} `json:"data"`
}

func deleteMsg(msgID int) {
	req, _ := http.NewRequest("POST", cqHost+"/delete_msg", strings.NewReader(fmt.Sprintf(`{"message_id": %d}`, msgID)))
	req.Header.Add("content-type", "application/json")
	do, _ := c.Do(req)
	defer do.Body.Close()
}

func send(message *Message, msg string) int {
	if message.GroupID == 0 && message.UserID == 0 {
		log.Println("GroupID == 0, UserID == 0")
		return 0
	}
	var req *http.Request
	if message.GroupID > 0 {
		req, _ = http.NewRequest("POST", cqHost+"/send_group_msg", strings.NewReader(fmt.Sprintf(`{"group_id": %d, "message": %q}`, message.GroupID, strings.Trim(msg, "\n"))))
	} else {
		req, _ = http.NewRequest("POST", cqHost+"/send_msg", strings.NewReader(fmt.Sprintf(`{"user_id": %d, "message": %q}`, message.UserID, strings.Trim(msg, "\n"))))
	}
	req.Header.Add("content-type", "application/json")
	do, _ := c.Do(req)
	defer do.Body.Close()
	var res sendResponse
	json.NewDecoder(do.Body).Decode(&res)
	return res.Data.MessageID
}
