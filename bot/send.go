package bot

import (
	"fmt"
	"net/http"
	"strings"
)

const CQHost = "http://127.0.0.1:5700"

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
	Anonymous     *Anonymous
	Message       string `json:"message"`
	Sender
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
type Sender struct {
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

type Anonymous struct {
	ID   int64  `json:"id"`   //匿名用户 ID
	Name string `json:"name"` //匿名用户名称
	Flag string `json:"flag"` //匿名用户 flag, 在调用禁言 API 时需要传入
}

func Send(message Message, msg string) {
	req, _ := http.NewRequest("POST", CQHost+"/send_group_msg", strings.NewReader(fmt.Sprintf(`{"group_id": %d, "message": %q}`, message.GroupID, strings.Trim(msg, "\n"))))
	req.Header.Add("content-type", "application/json")
	do, _ := c.Do(req)
	defer do.Body.Close()
}
