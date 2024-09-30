package x

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/features/stock/httpproxy"
	"qq/util/random"
	"qq/util/translate"
	"strings"
	"text/template"
	"time"

	"github.com/golang-module/carbon/v2"
	twitterscraper "github.com/imperatrona/twitter-scraper"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func init() {
	features.AddKeyword("x", "<+user> <+maxLimit: 1> 获取用户最近推文", func(bot bot.Bot, content string) error {
		split := strings.Split(content, " ")
		if len(split) < 1 {
			bot.Send("查询格式不正确, eg: x user 5")
			return nil
		}
		m := NewManager(config.XTokens(), config.HttpProxy())
		user := split[0]
		limit := 1
		if len(split) > 1 {
			limit = cast.ToInt(split[1])
		}
		tweets, err := m.GetTweets(context.Background(), user, limit)
		if err != nil {
			bot.Send(err.Error())
		}
		for _, tweet := range tweets {
			func() {
				result, f := RenderTweetResult(tweet)
				defer f()
				if result != "" {
					bot.Send(result)
				}
			}()
		}

		return nil
	})
}

type Account struct {
	Token     string
	CSRFToken string
}

type Manager interface {
	login(ctx context.Context, account *Account) (*twitterscraper.Scraper, error)
	GetTweets(ctx context.Context, user string, maxTweets int) ([]*twitterscraper.TweetResult, error)
}

var _ Manager = (*manager)(nil)

type manager struct {
	tokens []config.Token
	proxy  string
}

func downloadPic(Photos []string) []string {
	return lo.Map(Photos, func(item string, index int) string {
		client := httpproxy.NewHttpProxyClient(config.HttpProxy())
		resp, err := client.Get(item)
		if err != nil {
			return ""
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return ""
		}
		savePath := filepath.Join(config.ImageDir, fmt.Sprintf("tmp-%s-%s%s", time.Now().Format("2006-01-02"), random.String(10), filepath.Ext(item)))
		create, err := os.Create(savePath)
		defer create.Close()
		io.Copy(create, resp.Body)
		return savePath
	})
}

func RenderTweetResult(r *twitterscraper.TweetResult) (string, func()) {
	if r.Username == "" {
		return "", func() {}
	}
	var b strings.Builder
	var Quoted map[string]any
	var removeImages []string
	if r.QuotedStatus != nil {
		localImages := downloadPic(lo.Map(r.QuotedStatus.Photos, func(item twitterscraper.Photo, index int) string { return item.URL }))
		Quoted = map[string]any{
			"Name":   r.QuotedStatus.Name,
			"Text":   r.QuotedStatus.Text,
			"Photos": localImages,
		}
		removeImages = append(removeImages, localImages...)
	}
	p := downloadPic(lo.Map(r.Photos, func(item twitterscraper.Photo, index int) string { return item.URL }))
	removeImages = append(removeImages, p...)
	tweetTemplate.Execute(&b, map[string]any{
		"DateString": r.TimeParsed.Local().Format("2006-01-02 15:04:05"),
		"Name":       r.Name,
		"Text":       r.Text,
		"Photos":     p,
		"Quoted":     Quoted,
		"Link":       r.PermanentURL,
	})
	return b.String(), func() {
		for _, img := range removeImages {
			os.Remove(img)
			log.Println("[RenderTweetResult] 删除图片: ", img)
		}
	}
}

func NewManager(tokens []config.Token, proxy string) Manager {
	return &manager{tokens: tokens, proxy: proxy}
}

func (m *manager) login(ctx context.Context, account *Account) (*twitterscraper.Scraper, error) {
	scraper := twitterscraper.New()
	scraper.SetProxy(m.proxy)
	scraper.SetAuthToken(twitterscraper.AuthToken{
		Token:     account.Token,
		CSRFToken: account.CSRFToken,
	})
	if !scraper.IsLoggedIn() {
		return nil, errors.New("Invalid AuthToken")
	}
	return scraper, nil
}

func (m *manager) GetTweets(ctx context.Context, user string, maxTweets int) ([]*twitterscraper.TweetResult, error) {
	var aggerr = NewAggregateError()
	for _, token := range m.tokens {
		scraper, err := m.login(ctx, &Account{Token: token.Token, CSRFToken: token.CSRF})
		if err != nil {
			aggerr.Add(fmt.Errorf("token: %v, 登录失败: %v", token.Token, err))
			log.Printf("%s 登录失败: %v\n", token.Token, err)
			continue
		}
		var res []*twitterscraper.TweetResult
		for tweet := range scraper.GetTweets(context.Background(), user, maxTweets) {
			if tweet.Error != nil {
				aggerr.Add(tweet.Error)
			}
			res = append(res, tweet)
		}
		return res, aggerr.ToError()
	}
	return nil, errors.New("没设置token")
}

type AggregateError struct {
	e []error
}

func NewAggregateError() *AggregateError {
	return &AggregateError{}
}

func (a *AggregateError) ToError() error {
	if len(a.e) == 0 {
		return nil
	}
	var b strings.Builder
	for _, err := range a.e {
		b.WriteString(err.Error())
		b.WriteString("\n")
	}
	return errors.New(b.String())
}

func (a *AggregateError) Add(err error) {
	if err != nil {
		a.e = append(a.e, err)
	}
}

var tweetTemplate, _ = template.New("").Funcs(map[string]any{
	"humanize": func(s string) string {
		return carbon.Parse(s).DiffForHumans()
	},
	"translate": func(s string) string {
		if config.YDKey() == "" || config.YDSecret() == "" {
			return ""
		}
		return translate.EnToZh(config.YDKey(), config.YDSecret(), s)
	},
}).Parse(`
{{.Name}} 发推了！ {{ .DateString }} {{ humanize .DateString }}

👉 {{.Link}}

{{.Text}}

{{- range .Photos}}
[CQ:image,file=file://{{.}}]
{{- end}}
{{- if .Quoted}}

转发了 {{.Quoted.Name}}:
{{.Quoted.Text}}
{{- range .Quoted.Photos}}
[CQ:image,file=file://{{.}}]
{{- end}}
{{- end }}

翻译:
{{ translate .Text }}
{{- if .Quoted}}
===== 转发内容 =====
{{translate .Quoted.Text}}
{{- end }}
`)
