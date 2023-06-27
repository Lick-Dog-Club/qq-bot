package webot

import (
	"fmt"
	"io"
	"qq/bot"
	"qq/features"
	"qq/util"
	"strings"
	"sync"

	"github.com/eatmoreapple/openwechat"
	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("webot", "微信机器人扫码登录", func(bot bot.Bot, content string) error {
		RunWechat(bot)
		return nil
	})
}

func (sb *superBot) IsBotEnabledForThisMsg(msg *openwechat.Message) bool {
	log.Printf(`
msg.IsSendBySelf(): %v
msg.IsSendByGroup(): %v
msg.IsAt(): %v
msg.IsSendByFriend(): %v
msg.Owner().NickName: %v
`, msg.IsSendBySelf(), msg.IsSendByGroup(), msg.IsAt(), msg.IsSendByFriend(), msg.Owner().NickName)
	if msg.IsSendBySelf() {
		return true
	}

	if msg.IsSendByGroup() && msg.IsAt() && sb.um.exists(msg.Owner().NickName) {
		return true
	}

	if msg.IsSendByFriend() && sb.um.exists(msg.Owner().NickName) {
		return true
	}

	return false
}

type userMaps struct {
	sync.RWMutex
	users map[string]struct{}
}

func newUserMaps() *userMaps {
	return &userMaps{users: map[string]struct{}{}}
}

func (um *userMaps) add(nick string) {
	um.Lock()
	defer um.Unlock()
	um.users[nick] = struct{}{}
}

func (um *userMaps) del(nick string) {
	um.Lock()
	defer um.Unlock()
	delete(um.users, nick)
}

func (um *userMaps) exists(nick string) bool {
	um.RLock()
	defer um.RUnlock()
	_, ok := um.users[nick]
	return ok
}

func (um *userMaps) String() (res string) {
	um.RLock()
	defer um.RUnlock()
	var users []string
	for user, _ := range um.users {
		users = append(users, user)
	}

	return strings.Join(users, "\n")
}

type superBot struct {
	bot    *openwechat.Bot
	msgMap *bot.WeMsgMap
	um     *userMaps
}

func RunWechat(b bot.Bot) {
	webot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式
	var sb = &superBot{bot: webot, msgMap: bot.NewWeMsgMap(), um: newUserMaps()}

	// 注册消息处理函数
	webot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsText() && sb.IsBotEnabledForThisMsg(msg) {
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
			log.Printf("body: %v\n, key: %v\n,content: %v", body, keyword, content)
			if holdUp(sb, keyword, content) {
				if keyword == "list" {
					msg.ReplyText(sb.um.String())
					return
				}
				msg.ReplyText("done!")
				return
			}

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
					sb.msgMap.Add(image.MsgId, image)
					return image, err
				},
			}, sb.msgMap), keyword, content); err != nil {
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

func holdUp(sb *superBot, keyword string, content string) bool {
	switch keyword {
	case "add":
		log.Println("add: ", content)
		sb.um.add(content)
		return true
	case "del":
		log.Println("del: ", content)
		sb.um.del(content)
		return true
	case "list":
		return true
	default:
		return false
	}
}
