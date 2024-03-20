package pixiv

import (
	"context"
	"encoding/base64"
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
	"qq/util/proxy"
	"qq/util/retry"
	"strings"
	"time"

	"github.com/NateScarlet/pixiv/pkg/image"

	"github.com/NateScarlet/pixiv/pkg/artwork"
	log "github.com/sirupsen/logrus"

	"github.com/NateScarlet/pixiv/pkg/client"
)

var (
	httpClient = proxy.NewHttpProxyClient
)

func newClientCtx() (context.Context, error) {
	var s string = config.PixivSession()
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
	features.AddKeyword("pd", "pixiv_mode 设置成 daily", func(bot bot.Bot, content string) error {
		config.Set(map[string]string{"pixiv_mode": "daily"})
		bot.Send("pixiv_mode 已设置成 daily")
		return nil
	}, features.WithHidden(), features.WithGroup("pixiv"))
	features.AddKeyword("pw", "pixiv_mode 设置成 weekly", func(bot bot.Bot, content string) error {
		config.Set(map[string]string{"pixiv_mode": "weekly"})
		bot.Send("pixiv_mode 已设置成 weekly")
		return nil
	}, features.WithHidden())
	features.AddKeyword("pm", "pixiv_mode 设置成 monthly", func(bot bot.Bot, content string) error {
		config.Set(map[string]string{"pixiv_mode": "monthly"})
		bot.Send("pixiv_mode 已设置成 monthly")
		return nil
	}, features.WithHidden(), features.WithGroup("pixiv"))
	features.AddKeyword("p", "<+n/r/rai> 返回 pixiv 热门图片", func(bot bot.Bot, content string) error {
		image, err := Image(content)
		if err != nil {
			bot.Send(err.Error())
			return nil
		}
		var msgID string
		if bot.Message().WeSendImg != nil {
			open, _ := os.Open(image)
			defer open.Close()
			img, _ := bot.Message().WeSendImg(open)
			msgID = img.MsgId
		} else {
			open, _ := os.Open(image)
			defer open.Close()
			all, _ := io.ReadAll(open)
			toString := base64.StdEncoding.EncodeToString(all)
			msgID = bot.Send(fmt.Sprintf("[CQ:image,file=base64://%s]", toString))
		}
		os.Remove(image)
		if bot.IsGroupMessage() {
			tID := bot.Send("图片即将在 30s 之后撤回，要保存的赶紧了~")
			time.Sleep(30 * time.Second)
			bot.DeleteMsg(msgID)
			bot.DeleteMsg(tID)
		}
		return nil
	}, features.WithGroup("pixiv"), features.WithAIFunc(features.AIFuncDef{
		Call: func(args string) (string, error) {
			return Image("n")
		},
	}))

	features.AddKeyword("lsp", "<+query> 搜索 pixiv 的图片", func(bot bot.Bot, content string) error {
		if content == "" {
			bot.Send("请输入搜索内容")
			return nil
		}
		bot.Send(fmt.Sprintf("正在搜索 %s", content))
		search, err := Search(content, true)
		if err != nil {
			bot.Send(err.Error())
		}
		open, _ := os.Open(search)
		defer open.Close()
		all, _ := io.ReadAll(open)
		toString := base64.StdEncoding.EncodeToString(all)
		msgID := bot.Send(fmt.Sprintf("[CQ:image,file=base64://%s]", toString))
		os.Remove(search)
		if bot.IsGroupMessage() {
			tID := bot.Send("图片即将在 30s 之后撤回，要保存的赶紧了~")
			time.Sleep(30 * time.Second)
			bot.DeleteMsg(msgID)
			bot.DeleteMsg(tID)
		}
		return nil
	}, features.WithGroup("pixiv"))
}

func Search(content string, yell bool) (string, error) {
	// 搜索画作
	var res artwork.SearchResult
	ctx, e := newClientCtx()
	if e != nil {
		return "", e
	}
	var opts = []artwork.SearchOption{
		artwork.SearchOptionOrder(artwork.OrderDateDSC),
	}
	if yell {
		opts = append(opts, artwork.SearchOptionContentRating(artwork.ContentRatingR18))
	}
	if err := retry.Times(3, func() error {
		var err error
		res, err = artwork.Search(
			ctx,
			content,
			opts...,
		)

		return err
	}); err != nil {
		return "", err
	}
	if len(res.Artworks()) > 0 {
		items := res.Artworks()
		log.Println(len(items))
		return downloadImage(ctx, items[rand.Intn(len(items))])
	}
	return "", errors.New("没搜索到相关图片")
}

func Image(content string) (string, error) {
	//daily_r18_ai
	//daily_r18
	//daily
	//daily_ai
	//monthly
	isDaily := func() bool {
		return strings.Contains(config.PixivMode(), "daily")
	}
	var mode = config.PixivMode()
	switch content {
	case "n":
	case "r":
		mode = mode + "_r18"
	case "rai":
		if isDaily() {
			mode = mode + "_r18_ai"
			break
		}
		mode = mode + "_r18"
	default:
		if isDaily() {
			mode = mode + "_ai"
		}
	}
	ctx, err := newClientCtx()
	if err != nil {
		log.Println(err)
		return "", err
	}
	rank := &artwork.Rank{Mode: mode}
	err = retry.Times(20, func() error {
		rank.Page = 1
		if !isDaily() {
			rank.Page = rand.Intn(5) + 1
		}
		return rank.Fetch(ctx)
	})
	if err != nil {
		log.Println(err)
		return "", err
	}
	return downloadImage(ctx, rank.Items[rand.Intn(len(rank.Items))].Artwork)
}

var DIR = config.ImageDir

func getImage(img image.URLs) string {
	if img.Original != "" {
		return img.Original
	}
	if img.Regular != "" {
		return img.Regular
	}
	if img.Small != "" {
		return img.Small
	}
	if img.Thumb != "" {
		return img.Thumb
	}
	if img.Mini != "" {
		return img.Mini
	}
	return ""
}

func downloadImage(ctx context.Context, a artwork.Artwork) (string, error) {
	i := &artwork.Artwork{ID: a.ID}
	retry.Times(3, func() error {
		return i.Fetch(ctx)
	})
	img := getImage(i.Image)
	if img == "" {
		return "", errors.New("图片地址为空")
	}
	var get *http.Response
	c := httpClient()
	if err := retry.Times(20, func() error {
		var err error
		request, _ := http.NewRequest("GET", img, nil)
		request.Header.Add("Referer", "https://www.pixiv.net/")
		get, err = c.Do(request)
		if err != nil {
			log.Println(err, img)
		}
		return err
	}); err != nil {
		log.Println(err)
		return "", err
	}
	defer get.Body.Close()
	parse, _ := url.Parse(img)
	base := filepath.Base(parse.Path)
	fpath := filepath.Join(DIR, base)

	os.MkdirAll(filepath.Join(DIR), 0755)
	if err := func() error {
		file, err := os.OpenFile(fpath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Println(err)
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, get.Body)
		return err
	}(); err != nil {
		log.Println(err)
		return "", err
	}

	return fpath, nil
}
