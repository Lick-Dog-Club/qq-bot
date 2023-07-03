package ddys

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"qq/bot"
	"qq/features"
	"qq/features/util/proxy"
	"qq/features/util/retry"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("ddys", "<+dy/dm>, 获取更新的电影/动漫, 默认 +dy", func(bot bot.Bot, content string) error {
		for _, m := range Get(content, 3*24*time.Hour) {
			bot.Send(m.String())
		}
		return nil
	})
}

type movie struct {
	Url            string
	Title          string
	Director       string
	Year           string
	Kind           string
	HeadImageUrl   string
	Rating         string
	Country        string
	UpdateAt       time.Time
	ImageLocalPath string
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
更新时间: {{ .UpdateAt | datestr }}

[CQ:image,file=file://{{ .ImageLocalPath }}]
`)

func (m *movie) String() string {
	bf := bytes.Buffer{}
	temp.Execute(&bf, m)
	return bf.String()
}

func (m *movie) isNew(duration time.Duration) bool {
	now := time.Now()
	fromDate := time.Date(now.Year(), now.Month(), now.Day(), 9, 10, 0, 0, time.Local).Add(-duration)

	return m.UpdateAt.After(fromDate)
}

func buildRequest(url string) *http.Request {
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	return request
}

func doRequest(req *http.Request) (resp *http.Response, err error) {
	retry.Times(10, func() error {
		resp, err = proxy.NewHttpProxyClient().Do(req)
		return err
	})
	return
}

func Get(param string, duration time.Duration) (res []*movie) {
	url := "https://ddys.art/category/movie/"
	if param == "dm" {
		url = "https://ddys.art/category/anime/new-bangumi/"
	}
	do, err := doRequest(buildRequest(url))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer do.Body.Close()
	doc, err := htmlquery.Parse(do.Body)
	if err != nil {
		log.Fatal(err)
	}
	// 获取电影详情页的 url
	var (
		articleDetailUrlCh chan string = make(chan string, 20)
		wg                             = sync.WaitGroup{}
		resultCh           chan *movie = make(chan *movie, 20)
	)

	for i := 0; i < 5; i++ {
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

func writeImage(imageUrl string) string {
	base := filepath.Base(imageUrl)
	fpath := filepath.Join("/data", "ddys-images", base)

	os.MkdirAll(filepath.Join("/data", "ddys-images"), 0755)
	get, err := doRequest(buildRequest(imageUrl))
	if err != nil {
		return ""
	}
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
		return ""
	}

	return fpath
}

func fetchDetail(url string) (m *movie) {
	log.Println(url)
	do, err := doRequest(buildRequest(url))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer do.Body.Close()
	parse, err := htmlquery.Parse(do.Body)
	if err != nil {
		log.Fatal(err)
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
				m.UpdateAt = t
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
				m.HeadImageUrl = attribute.Val
			}
		}
	}
	m.ImageLocalPath = writeImage(m.HeadImageUrl)

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
