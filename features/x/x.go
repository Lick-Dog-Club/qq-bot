package x

import (
	"context"
	"errors"
	"log"
	"qq/bot"
	"qq/config"
	"text/template"

	"github.com/samber/lo"

	"qq/features"
	"strings"

	"github.com/golang-module/carbon/v2"
	twitterscraper "github.com/imperatrona/twitter-scraper"
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
			bot.Send(RenderTweetResult(tweet))
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

func RenderTweetResult(r *twitterscraper.TweetResult) string {
	var b strings.Builder
	var Quoted map[string]any
	if r.QuotedStatus != nil {
		Quoted = map[string]any{
			"Text": r.QuotedStatus.Text,
			"Photos": lo.Map(r.QuotedStatus.Photos, func(item twitterscraper.Photo, index int) string {
				return item.URL
			}),
		}
	}
	tweetTemplate.Execute(&b, map[string]any{
		"DateString": r.TimeParsed.Local().Format("2006-01-02 15:04:05"),
		"Name":       r.Username,
		"Text":       r.Text,
		"Photos": lo.Map(r.Photos, func(item twitterscraper.Photo, index int) string {
			return item.URL
		}),
		"Quoted": Quoted,
	})
	return b.String()
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
	var aggerr = newAggregateError()
	for _, token := range m.tokens {
		scraper, err := m.login(ctx, &Account{Token: token.Token, CSRFToken: token.CSRF})
		if err != nil {
			aggerr.Add(err)
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

type aggregateError struct {
	e []error
}

func newAggregateError() *aggregateError {
	return &aggregateError{}
}

func (a *aggregateError) ToError() error {
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

func (a *aggregateError) Add(err error) {
	if err != nil {
		a.e = append(a.e, err)
	}
}

var tweetTemplate, _ = template.New("").Funcs(map[string]any{
	"humanize": func(s string) string {
		return carbon.Parse(s).DiffForHumans()
	},
}).Parse(`
{{.Name}} 发推了！ {{ .DateString }} {{ humanize .DateString }}

{{.Text}}

{{- range .Photos}}
[CQ:image,file={{.}}]
{{- end}}
{{- if .Quoted}}

转发了:
{{.Quoted.Text}}
{{- range .Quoted.Photos}}
[CQ:image,file={{.}}]
{{- end}}
{{- end }}
`)
