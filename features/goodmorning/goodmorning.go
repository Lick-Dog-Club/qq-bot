package goodmorning

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/features/holiday"
	"qq/features/huangli"
	"qq/features/star"
	"qq/features/weather"
	"qq/util"
	"strings"
	"time"

	"github.com/samber/lo"
)

func init() {
	features.AddKeyword("gm", "早上好", func(bot bot.Bot, content string) error {
		bot.SendTextImage(Get(time.Now()))
		return nil
	})
}

func Get(n time.Time) string {
	get := holiday.GetItems(time.Now().Year())
	var ap string
	for _, i := range get {
		if i.Date == n.Format("2006-01-02") {
			s := "好好休息吧"
			if !i.IsOffDay {
				s = "需要上班"
			}
			ap = fmt.Sprintf("【%s】, %s", i.Name, s)
			break
		}
	}
	content := fmt.Sprintf(`今天是 %s

%s
=======================
%s
=======================
%s
=======================
%s
%s
`, n.Format("2006-01-02")+" "+holiday.WeekDays[n.Weekday()]+" "+ap, huangli.Get(n).Tldr(), weather.Get("杭州"), strings.Join(lo.ChunkString(star.Get(config.Birthday()), 40), "\n"), GetBirthDayInfo(), holiday.GetNextHolidays(n).Render())
	return content
}

func GetBirthDayInfo() string {
	if config.Birthday() == "" {
		return "未设置生日信息"
	}
	from := GetDayFrom(config.Birthday())
	parse, _ := time.Parse("2006-01-02", config.Birthday())
	split := strings.Split(huangli.Get(parse).Lunardate, "-")
	yli, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-%.2d-%.2d", time.Now().Year(), util.ToInt64(split[1]), util.ToInt64(split[2])))
	lunar := huangli.GetLunar(time.Date(time.Now().Year(), yli.Month(), yli.Day(), 0, 0, 0, 0, time.Local))
	lfrom := GetDayFrom(lunar.Gregoriandate)
	bf := bytes.Buffer{}
	btemp.Execute(&bf, map[string]interface{}{
		"IsGBirthday": from == 0,
		"GDay":        fmt.Sprintf("%d%s", time.Now().Year(), parse.Format("-01-02")),
		"GFrom":       from,
		"IsLBirthday": lfrom == 0,
		"LFrom":       lfrom,
		"LDay":        lunar.Gregoriandate,
	})
	return bf.String()
}

var btemp, _ = template.New("").Parse(`
{{- if .IsLBirthday}}
农历生日快乐🎉🎊
{{ else}}
距离农历生日【{{.LDay}}】还有 {{.LFrom}} 天
{{- end}}
{{- if .IsGBirthday}}
阳历生日快乐🎉🎊
{{ else}}
距离阳历生日【{{.GDay}}】还有 {{.GFrom}} 天
{{- end -}}
`)

func GetDayFrom(day string) int {
	parse, _ := time.Parse("2006-01-02", day)
	birthday := time.Date(time.Now().Year(), parse.Month(), parse.Day(), 23, 59, 59, 0, time.Local)
	var dayG float64
	if birthday.After(time.Now()) {
		dayG = birthday.Sub(time.Now()).Hours() / 24
	} else {
		dayG = birthday.AddDate(1, 0, 0).Sub(time.Now()).Hours() / 24
	}
	return int(math.Floor(dayG))
}
