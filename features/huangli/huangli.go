package huangli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qq/config"
	"time"
)

// Get 获取阳历
func Get(Day time.Time) Result {
	return getByType(Day, 0)
}

// GetLunar 获取阴历
func GetLunar(Day time.Time) Result {
	return getByType(Day, 1)
}

func getByType(Day time.Time, t int) Result {
	day := Day.Format("2006-01-02")
	if t == 1 {
		day = fmt.Sprintf("%d-%v-%v", Day.Year(), int64(Day.Month()), Day.Day())
	}
	get, _ := http.Get(fmt.Sprintf("https://apis.tianapi.com/lunar/index?key=%s&date=%s&type=%v", config.TianApiKey(), day, t))
	defer get.Body.Close()
	var data response
	json.NewDecoder(get.Body).Decode(&data)
	fmt.Println(data)
	return data.Result
}

type Result struct {
	Gregoriandate     string `json:"gregoriandate"`
	Lunardate         string `json:"lunardate"`
	LunarFestival     string `json:"lunar_festival"`
	Festival          string `json:"festival"`
	Fitness           string `json:"fitness"`
	Taboo             string `json:"taboo"`
	Shenwei           string `json:"shenwei"`
	Taishen           string `json:"taishen"`
	Chongsha          string `json:"chongsha"`
	Suisha            string `json:"suisha"`
	Wuxingjiazi       string `json:"wuxingjiazi"`
	Wuxingnayear      string `json:"wuxingnayear"`
	Wuxingnamonth     string `json:"wuxingnamonth"`
	Xingsu            string `json:"xingsu"`
	Pengzu            string `json:"pengzu"`
	Jianshen          string `json:"jianshen"`
	Tiangandizhiyear  string `json:"tiangandizhiyear"`
	Tiangandizhimonth string `json:"tiangandizhimonth"`
	Tiangandizhiday   string `json:"tiangandizhiday"`
	Lmonthname        string `json:"lmonthname"`
	Shengxiao         string `json:"shengxiao"`
	Lubarmonth        string `json:"lubarmonth"`
	Lunarday          string `json:"lunarday"`
	Jieqi             string `json:"jieqi"`
}

func (r Result) Tldr() string {
	s := fmt.Sprintf("%s%s\n适宜: %s\n禁忌: %s", r.Lubarmonth, r.Lunarday, r.Fitness, r.Taboo)
	if r.LunarFestival != "" {
		s += "\n农历节日: " + r.LunarFestival
	}
	return s
}

type response struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Result Result `json:"result"`
}
