package webot

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/features/ai"
	"qq/util"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/eatmoreapple/openwechat"
	log "github.com/sirupsen/logrus"
)

//1. 发送指令 "webot"
//2. 扫码登陆
//3. 微信需要对自己发送 "add 群组昵称" 才能开启机器人模式，不发就不会变成机器人
//4. 需要对用户开启的话也是 "add 用户昵称"
//5. 对自己发送的特殊指令：
//      list 列出所有开启机器人的用户或者群组
//      add 用户昵称/群组昵称
//      del 用户昵称/群组昵称

var o sync.Once

func init() {
	features.AddKeyword("webot", "微信机器人扫码登录", func(bot bot.Bot, content string) error {
		Run(bot)
		return nil
	})
}

func Run(bot bot.Bot) {
	o.Do(func() {
		runWechat(bot)
	})
}

func runWechat(b bot.Bot) {
	webot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式
	var sb = &superBot{bot: webot, msgMap: bot.NewWeMsgMap(), um: newUserMaps()}

	dispatcher := openwechat.NewMessageMatchDispatcher()
	dispatcher.OnImage(func(ctx *openwechat.MessageContext) {
		unix := time.Unix(ctx.Message.CreateTime, 0)
		if unix.Add(8 * time.Second).Before(time.Now()) {
			log.Println("过滤之前的消息", ctx.Message)
			return
		}
		log.Printf("%#v\n\n", ctx.Message)
		picture, _ := ctx.GetPicture()
		all, _ := io.ReadAll(picture.Body)
		b64 := ai.SeeB64(base64.StdEncoding.EncodeToString(all), http.DetectContentType(all))
		go func() {
			if sb.IsBotEnabledForThisMsg(ctx.Message) {
				msg := ctx.Message
				log.Printf("%#v\n\n", ctx.Message)
				gid := ""
				receiver, err := msg.Receiver()
				if err != nil {
					log.Println(err)
					return
				}

				var senderID string
				if msg.IsComeFromGroup() {
					gid = receiver.AvatarID()
					sender, _ := msg.SenderInGroup()
					senderID = receiver.UserName + sender.UserName
				} else {
					sender, _ := msg.Sender()
					senderID = receiver.UserName + sender.UserName
				}
				bot.NewWechatBot(bot.Message{
					SenderUserID:  senderID,
					Message:       msg.Content,
					IsSendByGroup: msg.IsComeFromGroup(),
					GroupID:       gid,
					WeReply:       replyText(msg),
					WeSendImg: func(file io.Reader) (*openwechat.SentMessage, error) {
						image, err := replyImg(msg)(file)
						if err != nil {
							log.Println(err)
							return nil, err
						}
						sb.msgMap.Add(image.MsgId, image)
						return image, err
					},
				}, sb.msgMap).Send(b64)
			}
		}()
	})
	dispatcher.OnText(func(ctx *openwechat.MessageContext) {
		unix := time.Unix(ctx.Message.CreateTime, 0)
		if unix.Add(8 * time.Second).Before(time.Now()) {
			log.Println("过滤之前的消息", ctx.Message)
			return
		}
		go func() {
			if sb.IsBotEnabledForThisMsg(ctx.Message) {
				msg := ctx.Message
				log.Printf("%#v\n\n", ctx.Message)
				gid := ""
				receiver, err := msg.Receiver()
				if err != nil {
					log.Println(err)
					return
				}

				var senderID string
				if msg.IsComeFromGroup() {
					gid = receiver.AvatarID()
					sender, _ := msg.SenderInGroup()
					senderID = receiver.UserName + sender.UserName
				} else {
					sender, _ := msg.Sender()
					senderID = receiver.UserName + sender.UserName
				}

				log.Printf(`
msg.Owner():
msg.Owner().Alias %v
msg.Owner().DisplayName %v
msg.Owner().NickName %v
msg.Owner().UserName %v
msg.Owner().RemarkName %v
`, msg.Owner().Alias, msg.Owner().DisplayName, msg.Owner().NickName, msg.Owner().UserName, msg.Owner().RemarkName)
				body := msg.Content
				if msg.IsComeFromGroup() {
					body = strings.ReplaceAll(body, "\u2005", " ")
					compile := regexp.MustCompile(`(@\S+)`)
					body = compile.ReplaceAllString(body, "")
				}
				keyword, content := util.GetKeywordAndContent(body)
				log.Printf("body: %v\n, key: %v\n,content: %v", body, keyword, content)

				if holdUp(sb, keyword, content) && msg.IsSendBySelf() {
					send := func(text string) {
						replyText(msg)(text)
					}
					if keyword == "list" {
						send(sb.um.String())
						return
					}
					send("done!")
					return
				}

				if err := features.Run(bot.NewWechatBot(bot.Message{
					SenderUserID:  senderID,
					Message:       msg.Content,
					IsSendByGroup: msg.IsComeFromGroup(),
					GroupID:       gid,
					WeReply:       replyText(msg),
					WeSendImg: func(file io.Reader) (*openwechat.SentMessage, error) {
						image, err := replyImg(msg)(file)
						if err != nil {
							log.Println(err)
							return nil, err
						}
						sb.msgMap.Add(image.MsgId, image)
						return image, err
					},
				}, sb.msgMap), keyword, content); err != nil {
					log.Println(err)
				}
			}
		}()
	})
	webot.MessageHandler = dispatcher.Dispatch

	// 注册登陆二维码回调
	webot.UUIDCallback = func(uuid string) {
		b.Send("访问下面网址扫描二维码登录")
		qrcodeUrl := openwechat.GetQrcodeUrl(uuid)
		log.Println(qrcodeUrl)
		b.Send(qrcodeUrl)
	}

	// 创建热存储容器对象
	f := "/data/webot-storage.json"
	if _, err2 := os.Stat(f); os.IsNotExist(err2) {
		create, _ := os.Create(f)
		create.Close()
	}
	reloadStorage := openwechat.NewFileHotReloadStorage(f)
	defer reloadStorage.Close()
	// 执行热登录
	if err := webot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		log.Println(err)
		return
	}

	// 获取登陆的用户
	self, err := webot.GetCurrentUser()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("当前用户是：", self.DisplayName)
	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	webot.Block()
}

func (sb *superBot) IsBotEnabledForThisMsg(msg *openwechat.Message) bool {
	log.Printf(`
msg.IsSendBySelf(): %v
msg.IsSendByGroup(): %v
msg.IsAt(): %v
msg.IsSendByFriend(): %v
msg.Owner().NickName: %v
`, msg.IsSendBySelf(), msg.IsSendByGroup(), msg.IsAt(), msg.IsSendByFriend(), msg.Owner().NickName)
	sender, _ := msg.Sender()
	receiver, _ := msg.Receiver()
	// 自己给自己发消息
	if msg.IsSendBySelf() && sender.NickName == receiver.NickName {
		return true
	}

	if (msg.IsSendByGroup() && msg.IsAt()) || msg.IsSendByFriend() && !msg.IsSendBySelf() {
		if sb.um.exists(sender.NickName) {
			return true
		}
	}

	return false
}

type userMaps struct {
	sync.RWMutex
	users map[string]struct{}
}

func newUserMaps() *userMaps {
	return &userMaps{users: config.WebotUsers()}
}

func (um *userMaps) add(nick string) {
	um.Lock()
	defer um.Unlock()
	users := config.WebotUsers()
	users[nick] = struct{}{}
	var s []string
	for name := range users {
		s = append(s, name)
	}
	config.Set(map[string]string{"webot_users": strings.Join(s, ",")})
}

func (um *userMaps) del(nick string) {
	um.Lock()
	defer um.Unlock()
	m := config.WebotUsers()
	delete(m, nick)
	var s []string
	for name := range m {
		s = append(s, name)
	}
	config.Set(map[string]string{"webot_users": strings.Join(s, ",")})
}

func (um *userMaps) exists(nick string) bool {
	um.RLock()
	defer um.RUnlock()
	_, ok := config.WebotUsers()[nick]
	return ok
}

func (um *userMaps) String() (res string) {
	um.RLock()
	defer um.RUnlock()
	var users []string
	for user := range config.WebotUsers() {
		users = append(users, user)
	}

	return strings.Join(users, "\n")
}

type superBot struct {
	bot    *openwechat.Bot
	msgMap *bot.WeMsgMap
	um     *userMaps
}

func replyText(msg *openwechat.Message) func(content string) (*openwechat.SentMessage, error) {
	if msg.IsSendBySelf() {
		if msg.IsSendByGroup() {
			return func(content string) (*openwechat.SentMessage, error) {
				return nil, errors.New("群里面不能自己用机器人")
			}
		}
	}
	return msg.ReplyText
}

func replyImg(msg *openwechat.Message) func(file io.Reader) (*openwechat.SentMessage, error) {
	if msg.IsSendBySelf() {
		if msg.IsSendByGroup() {
			return func(file io.Reader) (*openwechat.SentMessage, error) {
				return nil, errors.New("群里面不能自己用机器人")
			}
		}
	}
	return msg.ReplyImage
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
