package pixiv

import (
	"context"
	"crypto/tls"
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
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/NateScarlet/pixiv/pkg/artwork"
	"github.com/NateScarlet/pixiv/pkg/client"
	"github.com/cenkalti/backoff/v4"
)

var (
	httpClient = func() *http.Client {
		parse, _ := url.Parse(config.HttpProxy())
		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(parse),
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				MaxConnsPerHost: 1000,
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
	features.AddKeyword("pd", "pixiv_mode 设置成 daily", func(bot bot.Bot, content string) error {
		config.Set(map[string]string{"pixiv_mode": "daily"})
		bot.Send("pixiv_mode 已设置成 daily")
		return nil
	}, features.WithHidden())
	features.AddKeyword("pw", "pixiv_mode 设置成 weekly", func(bot bot.Bot, content string) error {
		config.Set(map[string]string{"pixiv_mode": "weekly"})
		bot.Send("pixiv_mode 已设置成 weekly")
		return nil
	}, features.WithHidden())
	features.AddKeyword("p", "+<n/r18/r18_ai> pixiv 热榜图片", func(bot bot.Bot, content string) error {
		//daily_r18_ai
		//daily_r18
		//daily
		//daily_ai
		var mode = config.PixivMode()
		switch content {
		case "n":
		case "r18":
			mode = mode + "_r18"
		case "r18_ai":
			mode = mode + "_r18_ai"
		default:
			mode = mode + "_ai"
		}
		ctx, err := newClientCtx()
		if err != nil {
			bot.Send(err.Error())
			log.Println(err)
			return nil
		}
		rank := &artwork.Rank{Mode: mode}
		err = retry(func() error {
			rank.Page = 1
			if config.PixivMode() != "daily" {
				rank.Page = rand.Intn(5) + 1
			}
			return rank.Fetch(ctx)
		})
		if err != nil {
			bot.Send(err.Error())
			log.Println(err)
			return nil
		}
		image := rank.Items[rand.Intn(len(rank.Items))]
		a := artwork.Artwork{
			ID: image.ID,
		}
		err = retry(func() error {
			return a.Fetch(ctx)
		})
		if err != nil {
			bot.Send(err.Error())
			log.Println(err)
			return nil
		}
		var get *http.Response
		c := httpClient()
		err = retry(func() error {
			var err error
			request, _ := http.NewRequest("GET", a.Image.Original, nil)
			request.Header.Add("Referer", "https://www.pixiv.net/")
			get, err = c.Do(request)
			return err
		})
		if err != nil {
			bot.Send(err.Error())
			log.Println(err)
			return nil
		}
		defer get.Body.Close()
		base := filepath.Base(a.Image.Original)
		fpath := filepath.Join("/data", "images", base)

		err = func() error {
			file, err := os.OpenFile(fpath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				log.Println(err)
				return err
			}
			defer file.Close()
			_, err = io.Copy(file, get.Body)
			return err
		}()
		if err != nil {
			log.Println(err)
			bot.Send(err.Error())
			return nil
		}
		msgID := bot.Send(fmt.Sprintf("[CQ:image,file=file://%s]", fpath))
		os.Remove(fpath)
		if bot.IsGroupMessage() {
			tID := bot.Send("图片即将在 30s 之后撤回，要保存的赶紧了~")
			time.Sleep(30 * time.Second)
			bot.DeleteMsg(msgID)
			bot.DeleteMsg(tID)
		}
		return nil
	})
}

func retry(fn func() error) error {
	return backoff.Retry(fn, backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Second), 20))
}
