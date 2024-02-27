package holiday

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"qq/bot"
	"qq/features"
	"qq/util/proxy"
	"sort"
	"time"

	"github.com/sashabaranov/go-openai/jsonschema"

	"github.com/samber/lo"
)

func init() {
	features.AddKeyword("holiday", "获取节假日数据, 获取法定节假日数据, 返回节日名称和具体的放假时间", func(bot bot.Bot, content string) error {
		bot.SendTextImage(Get(time.Now().Year()))
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"year": {
				Type:        jsonschema.Integer,
				Description: "4位数的年份, 例如 2024, 2023",
			},
		},
		Call: func(args string) (string, error) {
			var input = struct {
				Year int `json:"year"`
			}{}
			json.Unmarshal([]byte(args), &input)
			return Get(input.Year), nil
		},
	}))
	features.AddKeyword("next-holiday", "获取下一个节假日, 获取下一个法定节假日, 返回节日名称和具体的放假时间", func(bot bot.Bot, content string) error {
		bot.SendTextImage(GetNextHolidays().Render())
		return nil
	})
}

type response struct {
	Year   int       `json:"year"`
	Papers []string  `json:"papers"`
	Days   []Holiday `json:"days"`
}

type Holiday struct {
	Date     string `json:"date"`
	Name     string `json:"name"`
	IsOffDay bool   `json:"isOffDay"`
}

func (h Holiday) Datetime() time.Time {
	parse, _ := time.Parse("2006-01-02", h.Date)
	return parse
}

type Holidays []Holiday

func (h Holidays) Len() int {
	return len(h)
}

func (h Holidays) Less(i, j int) bool {
	idate, _ := time.Parse("2006-01-02", h[i].Date)
	jDate, _ := time.Parse("2006-01-02", h[i].Date)
	return idate.Before(jDate)
}

func (h Holidays) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func Get(year int) string {
	resp, _ := proxy.NewHttpProxyClient().Get(fmt.Sprintf("https://raw.githubusercontent.com/NateScarlet/holiday-cn/master/%d.json", year))
	defer resp.Body.Close()

	var data response
	json.NewDecoder(resp.Body).Decode(&data)
	filter := lo.Filter(data.Days, func(item Holiday, index int) bool {
		return item.IsOffDay
	})
	bf := &bytes.Buffer{}
	temp.Execute(bf, map[string]any{
		"Days": filter,
	})
	return bf.String()
}

func GetItems(year int) []Holiday {
	resp, _ := proxy.NewHttpProxyClient().Get(fmt.Sprintf("https://raw.githubusercontent.com/NateScarlet/holiday-cn/master/%d.json", year))
	defer resp.Body.Close()

	var data response
	json.NewDecoder(resp.Body).Decode(&data)
	filter := lo.Filter(data.Days, func(item Holiday, index int) bool {
		return item.IsOffDay
	})
	return filter
}

var temp = template.Must(template.New("").Parse(`
{{ range $item := .Days }}
 节日: {{$item.Name}}, 日期: {{ $item.Date }}
{{- end }}
`))

func GetNextHolidays() Holis {
	items := GetItems(time.Now().Year())
	filter := lo.Filter(items, func(item Holiday, index int) bool {
		return item.Datetime().After(time.Now())
	})
	by := lo.GroupBy(filter, func(item Holiday) string {
		return item.Name
	})

	var holis Holis
	for name, holidays := range by {
		sort.Sort(Holidays(holidays))
		start := holidays[0].Datetime()
		end := holidays[len(holidays)-1].Datetime()
		holis = append(holis, Holi{
			StartDate:    start,
			EndDate:      end,
			StartWeekDay: WeekDays[start.Weekday()],
			EndWeekDay:   WeekDays[end.Weekday()],
			Name:         name,
			Days:         len(holidays),
			LeftDay:      int(math.Ceil(start.Sub(time.Now()).Hours() / 24)),
		})
	}
	sort.Sort(holis)
	return holis
}

var holidateTemp, _ = template.New("").Parse(`{{ range . }}
离 【{{.Name}}】还有 {{.LeftDay}} 天，从 {{.StartDate.Format "2006-01-02"}} ({{.StartWeekDay}}) 到 {{.EndDate.Format "2006-01-02"}} ({{.EndWeekDay}}) 共 {{.Days}} 天
{{- end -}}
`)

type Holi struct {
	StartDate    time.Time
	EndDate      time.Time
	StartWeekDay string
	EndWeekDay   string
	Name         string
	Days         int

	LeftDay int
}
type Holis []Holi

func (h Holis) Render() string {
	var bf = &bytes.Buffer{}
	holidateTemp.Execute(bf, h)
	return bf.String()
}

func (h Holis) Len() int {
	return len(h)
}

func (h Holis) Less(i, j int) bool {
	return h[i].StartDate.Before(h[j].StartDate)
}

func (h Holis) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

var WeekDays = map[time.Weekday]string{
	time.Sunday:    "星期日",
	time.Monday:    "星期一",
	time.Tuesday:   "星期二",
	time.Wednesday: "星期三",
	time.Thursday:  "星期四",
	time.Friday:    "星期五",
	time.Saturday:  "星期六",
}
