package holiday

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai/jsonschema"
	"html/template"
	"qq/bot"
	"qq/features"
	"qq/util/proxy"
	"time"

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

var temp = template.Must(template.New("").Parse(`
{{ range $item := .Days }}
 节日: {{$item.Name}}, 日期: {{ $item.Date }}
{{- end }}
`))
