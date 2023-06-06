package webot

import (
	"fmt"
	"io"
	"qq/bot"
	"qq/features"
	"qq/util"
	"strings"

	"github.com/eatmoreapple/openwechat"
	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("webot", "微信机器人扫码登录", func(bot bot.Bot, content string) error {
		RunWechat(bot)
		return nil
	})
}

func RunWechat(b bot.Bot) {
	webot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式

	// 注册消息处理函数
	webot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsText() && ((msg.IsSendByGroup() && msg.IsAt()) || msg.IsSendByFriend()) && !msg.IsSendBySelf() {
			gid := ""
			receiver, _ := msg.Receiver()
			senderID := receiver.ID()
			if msg.IsComeFromGroup() {
				gid = receiver.ID()
				sender, _ := msg.SenderInGroup()
				senderID = sender.ID()
			}

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
		b.Send("访问下面网址扫描二维码登录")
		qrcodeUrl := openwechat.GetQrcodeUrl(uuid)
		log.Println(qrcodeUrl)
		b.Send(qrcodeUrl)
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
	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	webot.Block()
}
