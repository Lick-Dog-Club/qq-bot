package pixiv

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"qq/bot"
	"qq/config"
	"qq/features"
	"sync"
	"time"

	"github.com/NateScarlet/pixiv/pkg/artwork"
	"github.com/NateScarlet/pixiv/pkg/client"
	"github.com/cenkalti/backoff/v4"
)

var (
	mu sync.RWMutex

	httpClient = func() *http.Client {
		parse, _ := url.Parse(config.HttpProxy())
		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(parse),
			},
		}
	}
)

func newClientCtx() (context.Context, error) {
	var s string = config.PixivSession()
	if s == "" {
		return nil, errors.New("请先设置session: pixiv-session +<session>")
	}
	// 使用 PHPSESSID Cookie 登录 (推荐)。
	c := &client.Client{
		Client: *httpClient(),
	}
	c.SetDefaultHeader("User-Agent", client.DefaultUserAgent)
	c.SetPHPSESSID(s)

	// 所有查询从 context 获取客户端设置, 如未设置将使用默认客户端。
	var ctx = context.TODO()
	ctx = client.With(ctx, c)
	return ctx, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
	features.AddKeyword("pixiv", "+<ai/r18/r18_ai> pixiv top", func(bot bot.Bot, content string) error {
		//daily_r18_ai
		//daily_r18
		//daily
		//daily_ai
		var mode = "daily"
		switch content {
		case "ai":
			mode = "daily_ai"
		case "r18":
			mode = "daily_r18"
		case "r18_ai":
			mode = "daily_r18_ai"
		}
		ctx, err := newClientCtx()
		if err != nil {
			bot.Send(err.Error())
			return nil
		}
		rank := &artwork.Rank{Mode: mode}
		err = backoff.Retry(func() error {
			rank.Page = rand.Intn(2)
			return rank.Fetch(ctx)
		}, backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Second), 10))
		if err != nil {
			bot.Send(err.Error())
			return nil
		}
                u := rank.Items[rand.Intn(len(rank.Items))].Image.Original
		request, _ := http.NewRequest("GET",u, nil)
		request.Header.Add("Referer", "https://www.pixiv.net/")
		var get *http.Response
		err = backoff.Retry(func() error {
			var err error
			get, err = httpClient().Do(request)
			return err
		}, backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Second), 10))
		if err != nil {
			bot.Send(err.Error())
			return nil
		}
		defer get.Body.Close()
		base := filepath.Base(u)
		fpath := filepath.Join("/data", "images", base)
		all, _ := io.ReadAll(get.Body)
		os.WriteFile(fpath, all, 0644)
		msgID := bot.Send(fmt.Sprintf("[CQ:image,file=file://%s]", fpath))
		if bot.IsGroupMessage() {
			tID := bot.Send("图片即将在 30s 之后撤回，要保存的赶紧了~")
			time.Sleep(30 * time.Second)
			bot.DeleteMsg(msgID)
			bot.DeleteMsg(tID)
		}
		return nil
	})
}
