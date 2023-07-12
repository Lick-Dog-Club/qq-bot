package comic

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"qq/bot"
	"qq/features"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func init() {
	features.AddKeyword("comic", "<+name: haizeiwang> 搜索漫画", func(bot bot.Bot, content string) error {
		url := fmt.Sprintf("http://www.yxtun.com/manhua/%s", content)
		if strings.HasPrefix(content, "http") {
			url = content
		}
		bot.Send(scrape(url).Render())
		return nil
	})
}

type comic struct {
	Name         string
	LastUrl      string
	LastTitle    string
	HeadImageUrl string
	UpdatedAt    time.Time
}

func dateStr(t time.Time) string {
	return t.Local().Format(time.DateTime)
}

var temp, _ = template.New("").Funcs(map[string]any{"datestr": dateStr}).Parse(`
动漫: {{.Name}}
更新时间：{{ .UpdatedAt | datestr }}
最新话: {{.LastTitle}}
最新话地址: {{.LastUrl}}

[CQ:image,file={{.HeadImageUrl}}]
`)

func (c *comic) Render() string {
	bf := bytes.Buffer{}
	temp.Execute(&bf, c)
	return bf.String()
}

func scrape(comicUrl string) *comic {
	var c = &comic{}
	resp, err := http.DefaultClient.Get(comicUrl)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return nil
	}

	find := htmlquery.Find(doc, `//span[@class="sj "]/text()`)
	for _, node := range find {
		parse, _ := time.ParseInLocation("2006-01-02", node.Data, time.Local)
		c.UpdatedAt = parse
	}
	for _, attribute := range htmlquery.Find(doc, `//p[@class="cover"]/img`)[0].Attr {
		if attribute.Key == "src" {
			c.HeadImageUrl = attribute.Val
		}
	}
	c.Name = htmlquery.Find(doc, `//div[@class="book-title"]/h1/span/text()`)[0].Data
	nodes := htmlquery.Find(doc, `//ul[@id="chapter-list-1"]/li/a@href]`)
	lastIndex := len(nodes) - 1
	if len(nodes) >= 2 {
		last := nodes[lastIndex]
		c.LastTitle = htmlquery.Find(last, "//span/text()")[0].Data
		c.LastUrl = fmt.Sprintf("http://www.yxtun.com%s", href(last.Attr))
	}
	return c
}

func href(attr []html.Attribute) string {
	for _, attribute := range attr {
		if attribute.Key == "href" {
			return attribute.Val
		}
	}
	return ""
}
