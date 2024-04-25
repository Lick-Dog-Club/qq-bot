package gold

import (
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/vicanso/go-charts/v2"
	"io"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/util"
	"qq/util/chart"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"
)

func init() {
	features.AddKeyword("gold", "金价", func(bot bot.Bot, content string) error {
		bot.SendTextImage(Get("JO_52683", 10).Render())
		return nil
	}, features.WithGroup("gold"))
	features.AddKeyword("gxx", "<+name: 模糊搜索> 金价", func(bot bot.Bot, content string) error {
		var code string
		for c, name := range m {
			if strings.Contains(name, content) {
				code = c
				break
			}
		}
		if code == "" {
			bot.Send("未找到该店铺")
			return nil
		}
		bot.SendTextImage(Get("code", 10).Render())
		return nil
	}, features.WithGroup("gold"))
	features.AddKeyword("glist", "金价店铺列表", func(bot bot.Bot, content string) error {
		var names []string
		for _, s := range m {
			names = append(names, s)
		}
		bot.SendTextImage(strings.Join(names, "\n"))
		return nil
	}, features.WithGroup("gold"))
	features.AddKeyword("gtop", "<+num: 数量> 前 {num} 便宜的店铺的金价", func(bot bot.Bot, content string) error {
		toInt64 := util.ToInt64(content)
		if toInt64 == 0 {
			toInt64 = 1
		}
		all := All(int(util.ToInt64(config.GoldCount())))
		lineChart := lineChartByLimit(all[:toInt64])
		bot.Send(fmt.Sprintf("[CQ:image,file=base64://%s]", lineChart))

		return nil
	}, features.WithGroup("gold"))
	features.AddKeyword("gx", "<+name: 模糊匹配店铺> 店铺金价", func(bot bot.Bot, content string) error {
		var code string
		for c, name := range m {
			if strings.Contains(name, content) {
				code = c
				break
			}
		}
		if code == "" {
			bot.Send("未找到该店铺")
			return nil
		}

		lineChart := lineChartByLimit(GoldList{Get(code, int(util.ToInt64(config.GoldCount())))})
		bot.Send(fmt.Sprintf("[CQ:image,file=base64://%s]", lineChart))

		return nil
	}, features.WithGroup("gold"))
}

func lineChartByLimit(all GoldList) string {
	var lines = map[string][]chart.XY{}
	// 日期 => 店铺 => 金价
	var allStores = map[string]struct{}{}
	var mm = map[int]map[string]chart.XY{}
	for _, g := range all {
		for _, datum := range g.Data {
			mmv, ok := mm[int(datum.Time)]
			if ok {
				allStores[g.CNName] = struct{}{}
				mmv[g.CNName] = chart.XY{
					X: time.UnixMilli(datum.Time).Format("06/01/02"),
					Y: datum.Q1,
				}
			} else {
				allStores[g.CNName] = struct{}{}
				mm[int(datum.Time)] = map[string]chart.XY{
					g.CNName: {
						X: time.UnixMilli(datum.Time).Format("06/01/02"),
						Y: datum.Q1,
					},
				}
			}
		}
	}
	keys := lo.Keys(mm)
	sort.Ints(keys)
	for _, key := range keys {
		mdata, _ := mm[key]
		for name := range allStores {
			xy, ok := mdata[name]
			if ok {
				_, ok := lines[name]
				if ok {
					lines[name] = append(lines[name], xy)
				} else {
					lines[name] = []chart.XY{xy}
				}
			} else {
				_, ok := lines[name]
				xy = chart.XY{
					X: time.UnixMilli(int64(key)).Format("06/01/02"),
					Y: charts.GetNullValue(),
				}
				if ok {
					lines[name] = append(lines[name], xy)
				} else {
					lines[name] = []chart.XY{xy}
				}
			}
		}
	}
	var showLabel = false
	if len(lines) == 1 {
		showLabel = config.GoldShowLabel()
	}
	lineChart := chart.DrawLineChart(chart.LineChartInput{
		Width:     1500,
		Height:    500,
		ShowLabel: showLabel,
		Lines:     lines,
		Base64:    true,
	})
	return lineChart
}

func allTodays() string {
	wg := sync.WaitGroup{}
	ch := make(chan Data, len(m))
	for code, name := range m {
		wg.Add(1)
		go func(code, name string) {
			defer wg.Done()
			get := Get(code, 1)
			if len(get.Data) > 0 && get.Data[0].IsToday() {
				ch <- Data{
					Name:  name,
					Code:  code,
					Price: get.Data[0].Q1,
					Diff:  get.Data[0].Q70,
				}
			}
		}(code, name)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	var items []Data
	for data := range ch {
		items = append(items, data)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Price > items[j].Price
	})
	var bf strings.Builder
	temp2.Execute(&bf, items)
	return bf.String()
}

type Data struct {
	Name  string
	Code  string
	Price float64
	Diff  float64
}

type GoldList []*Gold

func (g GoldList) Len() int {
	return len(g)
}

func (g GoldList) Less(i, j int) bool {
	return g[i].Data[0].Q1 < g[j].Data[0].Q1
}

func (g GoldList) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

func All(pageSize int) GoldList {
	wg := sync.WaitGroup{}
	ch := make(chan *Gold, len(m))
	for code, name := range m {
		wg.Add(1)
		go func(code, name string) {
			defer wg.Done()
			ch <- Get(code, pageSize)
		}(code, name)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	var items GoldList
	for data := range ch {
		items = append(items, data)
	}
	sort.Sort(items)
	return items
}

var temp2, _ = template.New("").Funcs(map[string]any{
	"datetime": func() string {
		return time.Now().Format("2006-01-02")
	},
}).Parse(`
{{ datetime }} 日，各店铺金价
{{ range .}}
{{ .Name}}: {{.Price}} 元/克，
{{- if gt .Diff 0.0 -}}
涨了 {{.Diff}} 元/克
{{- else if eq .Diff 0.0  -}}
平 {{.Diff}} 元/克
{{- else -}}
跌了 {{.Diff}} 元/克
{{- end -}}
{{end}}
`)

type Gold struct {
	Flag        bool   `json:"flag"`
	TotalPage   int    `json:"totalPage"`
	TotalCount  int    `json:"totalCount"`
	NowPage     int    `json:"nowPage"`
	Code        string `json:"code"`
	Style       int    `json:"style"`
	Digits      int    `json:"digits"`
	Status      int    `json:"status"`
	Unit        string `json:"unit"`
	ProductName string `json:"productName"`
	NextEndTime string `json:"nextEndTime"`
	Data        Items  `json:"data"`
	CNName      string
}

type Items []Item

type Item struct {
	Q1   float64 `json:"q1"`
	Q2   float64 `json:"q2"`
	Q3   float64 `json:"q3"`
	Q4   float64 `json:"q4"`
	Q60  float64 `json:"q60"`
	Q62  float64 `json:"q62"`
	Q128 float64 `json:"q128"`
	Q129 float64 `json:"q129"`
	Q70  float64 `json:"q70"`
	Time int64   `json:"time"`
}

func (i Item) IsToday() bool {
	return time.UnixMilli(i.Time).Format("2006-01-02") == time.Now().Format("2006-01-02")
}

func (g Items) Len() int {
	return len(g)
}

func (g Items) Less(i, j int) bool {
	return time.UnixMilli(g[i].Time).After(time.UnixMilli(g[j].Time))
}

func (g Items) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

var m = map[string]string{
	"JO_52683":  "上海黄金交易所",
	"JO_42660":  "周大福",
	"JO_42657":  "老凤祥",
	"JO_42653":  "周六福",
	"JO_42625":  "周生生",
	"JO_42646":  "六福珠宝",
	"JO_42638":  "菜百",
	"JO_42632":  "金至尊",
	"JO_42634":  "老庙",
	"JO_52670":  "潮宏基",
	"JO_52678":  "周大生",
	"JO_52672":  "亚一金店",
	"JO_52674":  "宝庆银楼",
	"JO_52676":  "太阳金店",
	"JO_52680":  "齐鲁金店",
	"JO_52686":  "千禧之星",
	"JO_52689":  "吉盟珠宝",
	"JO_52692":  "东祥金店",
	"JO_52694":  "萃华金店",
	"JO_52696":  "百泰黄金",
	"JO_52698":  "金象珠宝",
	"JO_52699":  "常州金店",
	"JO_52702":  "扬州金店",
	"JO_52703":  "嘉华珠宝",
	"JO_52705":  "福泰珠宝",
	"JO_52707":  "城隍珠宝",
	"JO_52709":  "星光达珠宝",
	"JO_52711":  "金兰首饰",
	"JO_92438":  "富艺珠宝",
	"JO_321446": "莱音珠宝",
	"JO_335546": "九龙福珠宝",
}

func Get(code string, size int) *Gold {
	if size <= 0 {
		size = 10
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.jijinhao.com/quoteCenter/history.htm?code=%s&style=3&pageSize=%d&needField=128,129,70&currentPage=1&_=%d",
			code,
			size,
			time.Now().UnixMilli(),
		),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("referer", "https://quote.cngold.org/gjs/swhj_zghj.html")
	req.Header.Set("sec-ch-ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("sec-fetch-dest", "script")
	req.Header.Set("sec-fetch-mode", "no-cors")
	req.Header.Set("sec-fetch-site", "cross-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, _ := io.ReadAll(resp.Body)
	index := strings.Index(string(bodyText), "{")
	var data *Gold
	json.NewDecoder(strings.NewReader(string(bodyText)[index:])).Decode(&data)
	data.CNName = m[code]
	sort.Sort(data.Data)

	return data
}

func (g Gold) Render() string {
	var bf strings.Builder
	temp.Execute(&bf, g)
	return bf.String()
}

var temp, _ = template.New("").Funcs(map[string]any{
	"millTime": func(t int64) string {
		return time.UnixMilli(t).Format("2006-01-02")
	},
}).Parse(`
{{.ProductName}}

{{range .Data}}
{{- millTime .Time -}}: {{.Q1}} 元/克，
{{- if gt .Q70 0.0 -}}
涨了 {{.Q70}} 元/克
{{- else if eq .Q70 0.0  -}}
平 {{.Q70}} 元/克
{{- else -}}
跌了 {{.Q70}} 元/克
{{- end -}}
{{"\n"}}
{{- end}}
`)
