package comic

import (
	"bytes"
	"fmt"
	"html/template"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"qq/bot"
	"qq/features"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/signintech/gopdf"

	"github.com/mozillazg/go-pinyin"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func init() {
	features.AddKeyword("comic", "<+name: haizeiwang/海贼王> 搜索漫画", func(bot bot.Bot, content string) error {
		c := Get(content)
		bot.Send(c.Render())
		bot.Send(fmt.Sprintf("[CQ:image,file=file://%s]", c.ToJPEG()))
		return nil
	})
}

type Comic struct {
	Name         string
	LastUrl      string
	LastTitle    string
	HeadImageUrl string
	UpdatedAt    time.Time
}

func (c *Comic) TodayUpdated() bool {
	if c == nil {
		return false
	}
	return c.UpdatedAt.Format("2006-01-02") == time.Now().Format("2006-01-02")
}

func dateStr(t time.Time) string {
	return t.Local().Format("2006-01-02")
}

var temp, _ = template.New("").Funcs(map[string]any{"datestr": dateStr}).Parse(`
动漫: {{.Name}}
更新时间：{{ .UpdatedAt | datestr }}
最新话: {{.LastTitle}}
最新话地址: {{.LastUrl}}

[CQ:image,file={{.HeadImageUrl}}]
`)

func (c *Comic) Render() string {
	if c == nil {
		return "未找到"
	}
	bf := bytes.Buffer{}
	temp.Execute(&bf, c)
	return bf.String()
}

func Get(titleOrUrl string) *Comic {
	titleOrUrl = strings.Join(pinyin.LazyConvert(titleOrUrl, &pinyin.Args{
		Fallback: func(r rune, a pinyin.Args) []string {
			return []string{string(r)}
		},
	}), "")
	url := fmt.Sprintf("http://www.yxtun.com/manhua/%s", titleOrUrl)
	if strings.HasPrefix(titleOrUrl, "http") {
		url = titleOrUrl
	}

	var c = &Comic{}
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode > 400 {
		return nil
	}
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
	for !strings.Contains(htmlquery.Find(nodes[lastIndex], "//span/text()")[0].Data, "话") {
		lastIndex -= 1
	}
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

type imageByte struct {
	index int
	path  string
	b     []byte
}

func (c *Comic) loadImages() [][]byte {
	// 获取所有图片的路径
	resp, _ := http.Get(c.LastUrl)
	defer resp.Body.Close()
	all, _ := io.ReadAll(resp.Body)
	compile := regexp.MustCompile(`chapterImages = \[(.*?)];`)
	submatch := compile.FindStringSubmatch(string(all))
	var picPaths []string
	for _, s := range strings.Split(submatch[1], ",") {
		path := strings.TrimRight(strings.TrimLeft(strings.ReplaceAll(s, `\/`, `/`), `"`), `"`)
		picPaths = append(picPaths, path)
		fmt.Println(path)
	}
	ch := make(chan *imageByte, 20)
	resultCh := make(chan *imageByte, 20)
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case s, ok := <-ch:
					if !ok {
						return
					}
					resultCh <- &imageByte{
						index: s.index,
						path:  s.path,
						b:     fetchImg(s.path),
					}
				}
			}
		}()
	}

	for i, path := range picPaths {
		ch <- &imageByte{
			index: i,
			path:  path,
		}
	}
	close(ch)
	go func() {
		wg.Wait()
		close(resultCh)
	}()
	var res = make([][]byte, len(picPaths))
	for v := range resultCh {
		res[v.index] = v.b
	}

	return res
}

func (c *Comic) ToJPEG() string {
	images := c.loadImages()
	var height, width int
	for _, imagePath := range images {
		w, y := imgWidthHeight(bytes.NewReader(imagePath))
		height += int(y)
		width = int(math.Max(float64(width), w))
	}
	// 创建新图片，大小为两张原图的宽度和高度之和
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	var y int
	for _, img := range images {
		decode, _, err := image.Decode(bytes.NewReader(img))
		if err == nil {
			decode.Bounds()
			// 将第一张图片绘制到新图片的顶部
			draw.Draw(newImg, image.Rect(0, y, width, y+decode.Bounds().Dy()), decode, image.Point{0, 0}, draw.Src)
			y += decode.Bounds().Dy()
			fmt.Println(y)
		}
	}

	// 将新图片保存到文件
	name := filepath.Join("/data", "images", c.LastTitle+".jpg")
	outFile, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()
	jpeg.Encode(outFile, newImg, &jpeg.Options{Quality: 100})
	return name
}

func (c *Comic) ToPDF() string {
	images := c.loadImages()
	var total float64
	for _, imagePath := range images {
		w, h := imgWidthHeight(bytes.NewReader(imagePath))
		iy := (595 / w) * h
		total += iy
	}
	fmt.Printf("total: %v", total)
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595, H: total}})
	pdf.AddPage()
	var y float64
	for _, imagePath := range images {
		decode, _, err := image.Decode(bytes.NewReader(imagePath))
		if err == nil {
			b := decode.Bounds()
			w := float64(b.Dx())
			h := float64(b.Dy())
			iy := (595 / w) * h

			reader, _ := gopdf.ImageHolderByReader(bytes.NewReader(imagePath))
			pdf.ImageByHolder(reader, 0, y, &gopdf.Rect{
				W: 595,
				H: iy,
			})
			y += iy
		}
	}
	path := filepath.Join("/data", "images", c.LastTitle+".pdf")
	pdf.WritePdf(path)
	return path
}

func fetchImg(path string) []byte {
	fmt.Println("fetch: " + path)
	request, _ := http.NewRequest("GET", path, nil)
	request.Header.Add("Referer", "https://m.yxtun.com/")
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	all, _ := io.ReadAll(do.Body)
	return all
}

func imgWidthHeight(reader io.Reader) (float64, float64) {
	decode, _, err := image.Decode(reader)
	if err != nil {
		return 0, 0
	}
	b := decode.Bounds()
	return float64(b.Dx()), float64(b.Dy())
}
