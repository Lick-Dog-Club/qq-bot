package holiday

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"qq/util/proxy"

	"github.com/samber/lo"
)

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
