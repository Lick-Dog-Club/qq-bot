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

	"golang.org/x/exp/slices"

	"github.com/sashabaranov/go-openai/jsonschema"

	"github.com/samber/lo"
)

func init() {
	features.AddKeyword("holiday", "获取年份对应的法定节假日数据, 返回节日名称和具体的放假时间", func(bot bot.Bot, content string) error {
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
	}), features.WithGroup("holiday"))
	features.AddKeyword("next-holiday", "获取下一个法定节假日, 返回节日名称和具体的放假时间", func(bot bot.Bot, content string) error {
		bot.SendTextImage(GetNextHolidays(time.Now()).Render())
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Call: func(args string) (string, error) {
			return GetNextHolidays(time.Now()).Render(), nil
		},
	}), features.WithGroup("holiday"))
}

type response struct {
	Year   int        `json:"year"`
	Papers []string   `json:"papers"`
	Days   []*Holiday `json:"days"`
}

type Holiday struct {
	Date        string `json:"date"`
	Name        string `json:"name"`
	IsOffDay    bool   `json:"isOffDay"`
	WeekDayName string `json:"weekDayName"`
}

func (h Holiday) Datetime() time.Time {
	parse, _ := time.Parse("2006-01-02", h.Date)
	return parse
}

type Holidays []*Holiday

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
	for _, day := range data.Days {
		parse, _ := time.Parse("2006-01-02", day.Date)
		day.WeekDayName = toWeekDay(parse.Weekday())
	}
	bf := &bytes.Buffer{}
	temp.Execute(bf, map[string]any{
		"Days": data.Days,
	})
	return bf.String()
}

func toWeekDay(weekday time.Weekday) string {
	return WeekDays[weekday]
}

func GetItems(year int) []*Holiday {
	resp, _ := proxy.NewHttpProxyClient().Get(fmt.Sprintf("https://raw.githubusercontent.com/NateScarlet/holiday-cn/master/%d.json", year))
	defer resp.Body.Close()

	var data response
	json.NewDecoder(resp.Body).Decode(&data)
	for _, day := range data.Days {
		parse, _ := time.Parse("2006-01-02", day.Date)
		day.WeekDayName = toWeekDay(parse.Weekday())
	}
	return data.Days
}

var temp = template.Must(template.New("").Parse(`
{{ range $item := .Days }}
 节日: {{$item.Name}}, 日期: {{ $item.Date }} {{- if not $item.IsOffDay }}，{{$item.WeekDayName}} 要上班 ！{{- end }}
{{- end }}
`))

func GetNextHolidays(n time.Time) Holis {
	res := getNextHolidays(n.Year(), n)
	if len(res) < 1 {
		res = getNextHolidays(n.Year()+1, n)
	}
	return res
}

func getNextHolidays(year int, now time.Time) Holis {
	items := GetItems(year)
	filter := lo.Filter(items, func(item *Holiday, index int) bool {
		if item.Datetime().Format("2006-01-02") == now.Format("2006-01-02") {
			return true
		}
		return item.Datetime().After(now)
	})
	by := lo.GroupBy(filter, func(item *Holiday) string {
		return item.Name
	})

	var holis Holis
	for name, holidays := range by {
		sort.Sort(Holidays(holidays))
		find, foundStart := lo.Find(holidays, func(item *Holiday) bool {
			return item.IsOffDay
		})
		l, foundEnd := lo.Find(lo.Reverse(slices.Clone(holidays)), func(item *Holiday) bool {
			return item.IsOffDay
		})
		var (
			start    time.Time
			end      time.Time
			IsPassed bool = false
		)
		if foundStart {
			start = find.Datetime()
		}
		if foundEnd {
			end = l.Datetime()
		}
		if !foundStart {
			IsPassed = true
		}
		holis = append(holis, Holi{
			StartDate:    start,
			EndDate:      end,
			StartWeekDay: WeekDays[start.Weekday()],
			EndWeekDay:   WeekDays[end.Weekday()],
			Name:         name,
			Days: len(lo.Filter(holidays, func(item *Holiday, index int) bool {
				return item.IsOffDay
			})),
			IsPassed: IsPassed,
			WorkDays: lo.Filter(holidays, func(item *Holiday, index int) bool {
				return !item.IsOffDay
			}),
			LeftDay: int(math.Ceil(start.Sub(now).Hours() / 24)),
		})
	}
	sort.Sort(holis)
	return holis
}

var holidateTemp, _ = template.New("").Parse(`
{{define "workday"}}
{{- if gt (len .) 0}}
{{ range $d := . }}
{{$d.Name}}: {{$d.Date}} {{$d.WeekDayName}}，要上班 ！
{{- end -}}
{{- end }}
{{end}}

{{- range . }}
{{- template "workday" .WorkDays -}}
{{- if not .IsPassed -}}
离 【{{.Name}}】还有 {{.LeftDay}} 天，从 {{.StartDate.Format "2006-01-02"}} ({{.StartWeekDay}}) 到 {{.EndDate.Format "2006-01-02"}} ({{.EndWeekDay}}) 共 {{.Days}} 天
{{- end -}}
{{- end -}}
`)

type Holi struct {
	StartDate    time.Time
	EndDate      time.Time
	StartWeekDay string
	EndWeekDay   string
	Name         string
	Days         int
	IsPassed     bool

	WorkDays []*Holiday

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
