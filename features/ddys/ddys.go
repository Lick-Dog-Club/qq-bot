package ddys

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/util/random"
	"qq/util/retry"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"

	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("ddys", "<+dy/dm>, 获取 ddys 网站的电影、动漫资讯, 默认 +dy", func(bot bot.Bot, content string) error {
		for _, m := range Get(content, 3*24*time.Hour) {
			bot.Send("正在获取数据中。。。")
			bot.Send(m.String())
		}
		return nil
	})
	// features.WithAIFunc(features.AIFuncDef{
	//		Properties: map[string]jsonschema.Definition{
	//			"param": {
	//				Type:        jsonschema.String,
	//				Description: "电影(dy) or 动漫(dm)",
	//				Enum:        []string{"dy", "dm"},
	//			},
	//		},
	//		Call: func(args string) (string, error) {
	//			var input = struct {
	//				Param string `json:"param"`
	//			}{}
	//			json.Unmarshal([]byte(args), &input)
	//			str := ""
	//			for _, m := range Get(input.Param, 3*24*time.Hour) {
	//				str += m.String()
	//			}
	//			return str, nil
	//		},
	//	}), features.WithAI()
}

type movie struct {
	Url          string
	Title        string
	Director     string
	Year         string
	Kind         string
	HeadImageUrl string
	Rating       string
	Country      string
	UpdatedAt    time.Time
}

func dateStr(t time.Time) string {
	return t.Local().Format(time.DateTime)
}

var temp, _ = template.New("").Funcs(map[string]any{"datestr": dateStr}).Parse(`
{{ .Title }} -- ({{ .Director }} {{ .Year }})
类型：{{ .Kind }}
评分：{{ .Rating }}
国家：{{ .Country }}
影片地址：{{ .Url }}
更新时间: {{ .UpdatedAt | datestr }}

[CQ:image,file=file://{{.HeadImageUrl}}]
`)

func (m *movie) String() string {
	bf := bytes.Buffer{}
	temp.Execute(&bf, m)
	return bf.String()
}

func (m *movie) isNew(duration time.Duration) bool {
	return m.UpdatedAt.After(time.Now().Add(-duration))
}

func buildRequest() *url.Request {
	req := url.NewRequest()
	headers := url.NewHeaders()
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	headers.Set("Referer", "https://ddys.art/")
	req.Headers = headers
	req.Proxies = config.HttpProxy()
	req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21-65037,29-23-24,0"
	return req
}

func doRequest(req *url.Request, url string) (resp *models.Response, err error) {
	retry.Times(10, func() error {
		resp, err = requests.Get(url, req)
		return err
	})
	return
}

func Get(param string, duration time.Duration) (res []*movie) {
	url := "https://ddys.art/category/movie/"
	if param == "dm" {
		url = "https://ddys.art/category/anime/new-bangumi/"
	}
	do, err := doRequest(buildRequest(), url)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer do.Body.Close()
	doc, err := htmlquery.Parse(do.Body)
	if err != nil {
		log.Println(err)
		return nil
	}
	// 获取电影详情页的 url
	var (
		articleDetailUrlCh chan string = make(chan string, 20)
		wg                             = sync.WaitGroup{}
		resultCh           chan *movie = make(chan *movie, 20)
	)

	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case path, ok := <-articleDetailUrlCh:
					if !ok {
						log.Println("articleDetailUrlCh done")
						return
					}
					detail := fetchDetail(path)
					if detail != nil {
						resultCh <- detail
					}
				}
			}
		}()
	}

	find := htmlquery.Find(doc, "//article")
	for _, node := range find {
		for _, attribute := range node.Attr {
			if attribute.Key == "data-href" {
				articleDetailUrlCh <- attribute.Val
			}
		}
	}
	close(articleDetailUrlCh)
	go func() {
		wg.Wait()
		close(resultCh)
	}()
	for ch := range resultCh {
		if ch.isNew(duration) {
			res = append(res, ch)
		}
	}

	return res
}

func fetchDetail(url string) (m *movie) {
	log.Println(url)
	do, err := doRequest(buildRequest(), url)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer do.Body.Close()
	parse, err := htmlquery.Parse(do.Body)
	if err != nil {
		log.Println(err)
		return nil
	}

	m = &movie{}
	m.Url = url

	// Title
	title := htmlquery.Find(parse, `//div[@class="post-content"]/h1/text()`)
	for _, node := range title {
		m.Title = node.Data
	}
	// updated
	updated := htmlquery.Find(parse, `//li[@class="meta_date"]/time`)
	for _, node := range updated {
		for _, attribute := range node.Attr {
			if attribute.Key == "datetime" {
				t, _ := time.Parse(time.RFC3339, attribute.Val)
				m.UpdatedAt = t
			}
		}
	}
	// abstract
	abstract := htmlquery.Find(parse, `//div[@class="abstract"]/text()`)
	for _, node := range abstract {
		if strings.HasPrefix(node.Data, "导演") {
			m.Director = content(node.Data)
		}
		if strings.HasPrefix(node.Data, "类型") {
			m.Kind = content(node.Data)
		}
		if strings.HasPrefix(node.Data, "制片国家") {
			m.Country = content(node.Data)
		}
		if strings.HasPrefix(node.Data, "年份") {
			m.Year = content(node.Data)
		}
	}

	// image
	image := htmlquery.Find(parse, `//div[@class="post"]/img`)
	for _, node := range image {
		for _, attribute := range node.Attr {
			if attribute.Key == "src" {
				m.HeadImageUrl = func() string {
					resp, err := doRequest(buildRequest(), attribute.Val)
					if err != nil {
						return ""
					}
					defer resp.Body.Close()
					if resp.StatusCode != 200 {
						return ""
					}
					savePath := filepath.Join("/data", "images", fmt.Sprintf("tmp-%s-%s%s", time.Now().Format("2006-01-02"), random.String(10), filepath.Ext(attribute.Val)))
					create, err := os.Create(savePath)
					defer create.Close()
					io.Copy(create, resp.Body)
					return savePath
				}()
			}
		}
	}

	// rating
	rating := htmlquery.Find(parse, `//div[@class="rating"]/span[@class="rating_nums"]/text()`)
	for _, node := range rating {
		m.Rating = node.Data
	}

	return
}

func content(data string) string {
	split := strings.Split(data, " ")
	if len(split) >= 2 {
		return split[1]
	}
	return ""
}
