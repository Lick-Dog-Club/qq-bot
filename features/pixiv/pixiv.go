package pixiv

import (
	"context"
	"errors"
	"math/rand"
	"qq/bot"
	"qq/features"
	"sync"
	"time"

	"github.com/NateScarlet/pixiv/pkg/artwork"
	"github.com/NateScarlet/pixiv/pkg/client"
)

var (
	session = ""
	mu      sync.RWMutex
)

func newClientCtx() (context.Context, error) {
	var s string
	func() {
		mu.RLock()
		defer mu.RUnlock()
		s = session
	}()
	if s == "" {
		return nil, errors.New("请先设置session: pixiv-session +<session>")
	}
	// 使用 PHPSESSID Cookie 登录 (推荐)。
	c := &client.Client{}
	c.SetDefaultHeader("User-Agent", client.DefaultUserAgent)
	c.SetPHPSESSID(s)

	// 所有查询从 context 获取客户端设置, 如未设置将使用默认客户端。
	var ctx = context.TODO()
	ctx = client.With(ctx, c)
	return ctx, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
	features.AddKeyword("pixiv-session", "设置 pixiv session", func(bot bot.Bot, content string) error {
		mu.Lock()
		defer mu.Unlock()
		session = content
		bot.Send("已设置 pixiv session")
		return nil
	})
	features.AddKeyword("pixiv", "+<ai/r18/r18_ai> pixiv top", func(bot bot.Bot, content string) error {
		//daily_r18_ai
		//daily_r18
		//daily
		//daily_ai
		switch content {
		case "ai":
			content = "daily_ai"
		case "r18":
			content = "daily_r18"
		case "r18_ai":
			content = "daily_r18_ai"
		default:
			content = "daily"
		}
		ctx, err := newClientCtx()
		if err != nil {
			bot.Send(err.Error())
			return nil
		}
		rank := &artwork.Rank{Mode: content, Page: rand.Intn(20)}
		if err = rank.Fetch(ctx); err != nil {
			bot.Send(err.Error())
			return nil
		}
		bot.Send(rank.Items[rand.Intn(len(rank.Items))].Image.Regular)
		return nil
	})
}
