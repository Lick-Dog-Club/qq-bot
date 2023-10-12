package telegram

import (
	"encoding/json"
	"net/url"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/util"
	"runtime"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/3bl3gamer/tgclient"
	"github.com/3bl3gamer/tgclient/mtproto"
	"golang.org/x/net/proxy"
)

var done = make(chan struct{})
var stopped = atomic.Bool{}

func init() {
	stopped.Store(true)
	features.AddKeyword("tg-stop", "暂停 telegram 监控", func(bot bot.Bot, content string) error {
		if !stopped.Swap(true) {
			close(done)
			done = make(chan struct{})
			bot.Send("已暂停")
		}
		return nil
	}, features.WithHidden())
	features.AddKeyword("tg-start", "telegram 消息监控并且发送到 bark", func(bot bot.Bot, content string) error {
		if config.UserID() == "" {
			bot.Send("user_id 为空")
			return nil
		}
		if !stopped.Swap(false) {
			bot.Send("tg 已经启动")
			return nil
		}
		proxyAddr := config.HttpProxy()
		if proxyAddr != "" {
			parse, _ := url.Parse(proxyAddr)
			proxyAddr = parse.Host
		}
		fn := func(s string) {
			bot.SendToUser(config.UserID(), s)
		}
		Run(config.TgAppID(), config.TgAppHash(), proxyAddr, &MyStore{}, fn, mtproto.ScanfAuthDataProvider{})
		//Run(config.TgAppID(), config.TgAppHash(), proxyAddr, &MyStore{}, fn, &BotAuthDataProvider{send: fn})
		return nil
	}, features.WithHidden())
}

type MyStore struct{}

func (m *MyStore) Save(info *mtproto.SessionInfo) error {
	marshal, _ := json.Marshal(info)
	config.Set(map[string]string{
		"tg_info": string(marshal),
	})
	return nil
}

func (m *MyStore) Load(info *mtproto.SessionInfo) error {
	info = config.TgInfo()
	return nil
}

func Run(appID int32, appHash string, proxyAddr string, store mtproto.SessionStore, sendToUser func(string), authData mtproto.AuthDataProvider) {
	cfg := &mtproto.AppConfig{
		AppID:          appID,
		AppHash:        appHash,
		AppVersion:     "0.0.1",
		DeviceModel:    "Unknown",
		SystemVersion:  runtime.GOOS + "/" + runtime.GOARCH,
		SystemLangCode: "en",
		LangPack:       "",
		LangCode:       "en",
	}
	var dialer proxy.Dialer
	if proxyAddr != "" {
		dialer, _ = proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
		log.Println("use proxyAddr: ", proxyAddr)
	}

	if store == nil {
		store = &MyStore{}
	}

	tg := tgclient.NewTGClientExt(cfg, store, &mtproto.SimpleLogHandler{}, dialer)
	tg.SetUpdateHandler(func(tl mtproto.TL) {
		if channel, ok := tl.(mtproto.TL_updateNewChannelMessage); ok {
			if message, ok := channel.Message.(mtproto.TL_message); ok {
				if id, ok := message.FromID.(mtproto.TL_peerUser); ok {
					user := tg.FindExtraUser(id.UserID)
					log.Println(message.Message, user.FirstName, user.ID)
					util.Bark(user.FirstName, message.Message, config.BarkUrls()...)
				} else {
					log.Println(message.Message)
				}
			}
		}
	})
	tg.InitAndConnect()
	if sendToUser == nil {
		sendToUser = func(s string) {
			log.Println(s)
		}
	}
	log.Println("start AuthExt...")
	tg.AuthExt(authData, mtproto.TL_users_getUsers{ID: []mtproto.TL{mtproto.TL_inputUserSelf{}}})
	defer tg.Stop()
	config.Set(map[string]string{"tg_code": ""})
	sendToUser("tg started")
	<-done
	sendToUser("tg exit")
}

type BotAuthDataProvider struct {
	send func(string)
}

func (b *BotAuthDataProvider) PhoneNumber() (string, error) {
	b.send("获取 tg_phone")
	if config.TgPhone() == "" {
		b.send("phone 未设置, 请设置 tg_phone")
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
LABEL:
	for {
		select {
		case <-ticker.C:
			if config.TgPhone() != "" {
				break LABEL
			}
		}
	}

	return config.TgPhone(), nil
}

func (b *BotAuthDataProvider) Code() (string, error) {
	b.send("获取 tg_code")
	if config.TgCode() == "" {
		b.send("code 未设置, 请设置 tg_code")
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
LABEL:
	for {
		select {
		case <-ticker.C:
			if config.TgCode() != "" {
				break LABEL
			}
		}
	}

	return config.TgCode(), nil
}

func (b *BotAuthDataProvider) Password() (string, error) {
	return "", nil
}
