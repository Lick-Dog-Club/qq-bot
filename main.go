package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"qq/bot"
	"qq/cronjob"
	"qq/features"
	"qq/util"
	"strings"

	"github.com/eatmoreapple/openwechat"
	log "github.com/sirupsen/logrus"

	// _ "qq/cronjob/lifetip"
	_ "qq/cronjob/dx"
	//_ "qq/cronjob/kfc"
	_ "qq/cronjob/lottery"
	_ "qq/cronjob/maotai"

	_ "qq/cronjob/picture"
	_ "qq/cronjob/tianqi"
	_ "qq/cronjob/xiaofeiquan"

	_ "qq/features/ai"
	_ "qq/features/bili-lottery"
	_ "qq/features/config"
	_ "qq/features/daxin"
	_ "qq/features/help"
	_ "qq/features/kfc"
	_ "qq/features/lifetip"
	_ "qq/features/picture"
	_ "qq/features/pixiv"
	_ "qq/features/raokouling"
	_ "qq/features/sysupdate"
	_ "qq/features/task"
	_ "qq/features/version"
	_ "qq/features/weather"
	_ "qq/features/weibo"
	_ "qq/features/zhihu"
)

func init() {
	log.SetReportCaller(true)
}

func main() {
	cm := cronjob.Manager()
	cm.Run(context.TODO())
	defer cm.Shutdown(context.TODO())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var message *bot.QQMessage
		json.NewDecoder(r.Body).Decode(&message)
		if message.PostType == "meta_event" {
			return
		}
		fmt.Printf("receive %d: %v\n", message.UserID, message.Message)
		atMsg := fmt.Sprintf("[CQ:at,qq=%v]", message.SelfID)
		if (strings.Contains(message.Message, atMsg) && message.MessageType == "group") || message.MessageType == "private" {
			msg := strings.ReplaceAll(message.Message, atMsg, "")
			keyword, content := util.GetKeywordAndContent(msg)
			if err := features.Run(bot.NewQQBot(&bot.Message{
				SenderUserID:  fmt.Sprintf("%d", message.UserID),
				Message:       content,
				IsSendByGroup: message.MessageType == "group",
				GroupID:       fmt.Sprintf("%d", message.GroupID),
			}), keyword, content); err != nil {
				log.Println(err)
			}
		}
	})
	go RunWechat()
	log.Println("[HTTP]: start...")
	log.Println(http.ListenAndServe(":5701", nil))
}

func RunWechat() {
	webot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式

	// 注册消息处理函数
	webot.MessageHandler = func(msg *openwechat.Message) {
		log.Printf(`
msg.IsSendByGroup: %v
msg.IsSendByFriend: %v
msg.IsAt: %v
msg: %v
FromUserName: %v
msg.IsFriendAdd(): %v
!msg.IsSendBySelf(): %v
msg.ToUserName: %v
`,
			msg.IsSendByGroup(),
			msg.IsSendByFriend(),
			msg.IsAt(),
			msg.Content,
			msg.FromUserName,
			msg.IsFriendAdd(),
			!msg.IsSendBySelf(),
			msg.ToUserName,
		)
		if msg.IsText() && ((msg.IsSendByGroup() && msg.IsAt()) || msg.IsSendByFriend()) && !msg.IsSendBySelf() {
			gid := ""
			receiver, _ := msg.Receiver()
			senderID := receiver.ID()
			if msg.IsComeFromGroup() {
				gid = receiver.ID()
				log.Println("receiver.NickName: ", receiver.NickName, receiver.DisplayName)
				sender, _ := msg.SenderInGroup()
				senderID = sender.ID()
				log.Println("sender.NickName: ", sender.NickName, sender.DisplayName)
			}
			log.Println("msg.Owner().NickName", msg.Owner().NickName, msg.Owner().DisplayName)

			log.Printf(`
UserName: %v
NickName: %v
DisplayName: %v
gid: %v,
msg.IsComeFromGroup(): %v"
%#v
`, msg.Owner().UserName, msg.Owner().NickName, msg.Owner().DisplayName, gid, msg.IsComeFromGroup(), bot.Message{
				SenderUserID:  senderID,
				Message:       msg.Content,
				IsSendByGroup: msg.IsComeFromGroup(),
				GroupID:       gid,
				WeReply:       msg.ReplyText,
				WeSendImg:     msg.ReplyImage,
			})

			atMsg := fmt.Sprintf("@%s", msg.Owner().NickName)
			body := strings.ReplaceAll(msg.Content, atMsg, "")
			keyword, content := util.GetKeywordAndContent(body)

			if err := features.Run(bot.NewWechatBot(bot.Message{
				SenderUserID:  senderID,
				Message:       msg.Content,
				IsSendByGroup: msg.IsComeFromGroup(),
				GroupID:       gid,
				WeReply:       msg.ReplyText,
				WeSendImg: func(file io.Reader) (*openwechat.SentMessage, error) {
					image, err := msg.ReplyImage(file)
					if err != nil {
						return nil, err
					}
					bot.WeMessageMap.Add(image.MsgId, image)
					return image, err
				},
			}), keyword, content); err != nil {
				log.Println(err)
			}
		}
	}
	// 注册登陆二维码回调
	webot.UUIDCallback = func(uuid string) {
		println("访问下面网址扫描二维码登录")
		qrcodeUrl := openwechat.GetQrcodeUrl(uuid)
		println(qrcodeUrl)
	}

	// 登陆
	if err := webot.Login(); err != nil {
		fmt.Println(err)
		return
	}

	// 获取登陆的用户
	self, err := webot.GetCurrentUser()
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println("当前用户是：", self.DisplayName)
	bot.WeFriends, _ = self.Friends()
	bot.WeGroups, _ = self.Groups()
	log.Println(bot.WeFriends)
	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	webot.Block()
}
