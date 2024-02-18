package trainticket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai/jsonschema"
)

func init() {
	features.AddKeyword("t", "查询高铁/火车票信息, 例如: 't 杭州东 绍兴北 20240216'", func(bot bot.Bot, content string) error {
		split := strings.Split(content, " ")
		from := GetStationCode(split[0])
		to := GetStationCode(split[1])
		if from == "" {
			bot.Send("出发地不正确")
			return nil
		}
		if to == "" {
			bot.Send("目的地不正确")
			return nil
		}
		date := time.Now()
		if len(split) > 2 {
			date, _ = time.Parse("20060102", split[2])
		}
		if _, err := bot.SendTextImage(Search(SearchInput{
			From:           from,
			To:             to,
			Date:           date.Format("2006-01-02"),
			OnlyShowTicket: false,
		}).FilterICanBuy().String()); err != nil {
			log.Println(err)
		}
		return nil
	}, features.WithGroup("train"))
	features.AddKeyword("to", "查询车票信息, 只显示有票的班次, 例如: 'to 杭州东 绍兴北 20240216'", func(bot bot.Bot, content string) error {
		split := strings.Split(content, " ")
		from := GetStationCode(split[0])
		to := GetStationCode(split[1])
		if from == "" {
			bot.Send("出发地不正确")
			return nil
		}
		if to == "" {
			bot.Send("目的地不正确")
			return nil
		}
		date := time.Now()
		if len(split) > 2 {
			date, _ = time.Parse("20060102", split[2])
		}
		if _, err := bot.SendTextImage(Search(SearchInput{
			From:           from,
			To:             to,
			Date:           date.Format("2006-01-02"),
			OnlyShowTicket: true,
		}).FilterICanBuy().String()); err != nil {
			log.Println(err)
		}
		return nil
	}, features.WithGroup("train"))

	features.AddKeyword("GetStationCodeByName", "<+name> 查询高铁/火车车站名称和 code 的对应关系表", func(bot bot.Bot, content string) error {
		bot.Send(GetStationCode(content))
		return nil
	}, features.WithHidden(), features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"name": {
				Type:        jsonschema.String,
				Description: "地点，例如 '杭州东' '绍兴北' 等",
			},
		},
		Call: func(args string) (string, error) {
			var a = struct {
				Name string `json:"name"`
			}{}
			json.Unmarshal([]byte(args), &a)
			return GetStationCode(a.Name), nil
		},
	}), features.WithGroup("train"))
	features.AddKeyword("Search12306", "", func(bot bot.Bot, content string) error {
		bot.Send("未实现该方法～")
		return nil
	}, features.WithHidden(), features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"from": {
				Type:        jsonschema.String,
				Description: "出发地, 需要通过 GetStationCodeByName 函数获取 code 值, 例如: 出发去杭州东, 需要根据 GetStationCodeByName 函数, 然后查到对应 from='HGH'",
			},
			"to": {
				Type:        jsonschema.String,
				Description: "目的地, 需要通过 GetStationCodeByName 函数获取 code 值, 例如: 出发去杭州东, 需要根据 GetStationCodeByName 函数, 然后查到对应 to='HGH'",
			},
			"date": {
				Type:        jsonschema.String,
				Description: "查询日期, 默认今天，日期格式: '2006-01-02', 例如: '2024-02-19'",
			},
			"only_show_ticket": {
				Type:        jsonschema.Boolean,
				Description: "是否只显示有票的班次",
			},
		},
		Call: func(args string) (string, error) {
			var input SearchInput
			json.Unmarshal([]byte(args), &input)
			return Search(input).String(), nil
		},
	}), features.WithGroup("train"))
}

var reg = regexp.MustCompile(`var station_names =\s?'(.*)';`)
var once sync.Once
var stations = make(map[string]Station)

type Station struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func GetStationCode(name string) string {
	if station, ok := stationNames()[name]; ok {
		fmt.Println(station.Code)
		return station.Code
	}
	return "未找到"
}

func StationNamesJson() string {
	stationNames()
	var res []Station
	for _, a := range stations {
		res = append(res, a)
	}
	marshal, _ := json.Marshal(res)
	return string(marshal)
}

func stationNames() map[string]Station {
	once.Do(func() {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://kyfw.12306.cn/otn/resources/js/framework/station_name.js?station_version=1.9298", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("sec-ch-ua", `"Not A(Brand";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
		req.Header.Set("Referer", "https://kyfw.12306.cn/otn/leftTicket/init?linktypeid=dc")
		req.Header.Set("sec-ch-ua-mobile", "?0")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
		req.Header.Set("sec-ch-ua-platform", `"macOS"`)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		submatch := reg.FindAllStringSubmatch(string(bodyText), -1)
		for _, d := range submatch {
			if len(d) == 2 {
				split := strings.Split(d[1], "|||")
				for _, str := range split {
					i := strings.Split(str, "|")
					if len(i) > 3 {
						stations[i[1]] = Station{
							Name: i[1],
							Code: i[2],
						}
					}
				}
			}
		}
	})
	return stations
}

type SearchResult []map[string]string

func (r SearchResult) String() string {
	bf := bytes.Buffer{}
	temp.Execute(&bf, map[string]any{
		"Data": r,
	})
	return bf.String()
}
func (r SearchResult) FilterICanBuy() SearchResult {
	var res SearchResult
	for _, item := range r {
		parse, _ := time.ParseInLocation("20060102 15:04", item["start_train_date"]+" "+item["start_time"], time.Local)
		if time.Now().Before(parse) {
			res = append(res, item)
		}
	}
	return res
}

type SearchInput struct {
	From           string `json:"from"`
	To             string `json:"to"`
	Date           string `json:"date"`
	OnlyShowTicket bool   `json:"only_show_ticket"`
}

func Search(input SearchInput) SearchResult {
	var from, to, date string = input.From, input.To, input.Date
	var onlyShowTicket bool = input.OnlyShowTicket
	fmt.Println(from, to, date)
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://kyfw.12306.cn/otn/leftTicket/queryE?leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=ADULT", date, from, to), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", fmt.Sprintf("JSESSIONID=%s;", config.T12306JSESSIONID()))
	req.Header.Set("If-Modified-Since", "0")
	req.Header.Set("Referer", "https://kyfw.12306.cn/otn/leftTicket/init?linktypeid=dc")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua", `"Not A(Brand";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var da = make([]map[string]string, 0)
	var data response
	json.NewDecoder(resp.Body).Decode(&data)
	for _, item := range data.Data.Result {
		split := strings.Split(item, "|")
		var m = make(map[string]any)
		var m2 = make(map[string]string)
		m["secretStr"] = split[0]
		m["buttonTextInfo"] = split[1]
		m2["train_no"] = split[2]
		m2["station_train_code"] = split[3]
		m2["start_station_telecode"] = split[4]
		m2["end_station_telecode"] = split[5]
		m2["from_station_telecode"] = split[6]
		m2["to_station_telecode"] = split[7]
		m2["start_time"] = split[8]
		m2["arrive_time"] = split[9]
		m2["lishi"] = split[10]
		m2["canWebBuy"] = split[11]
		m2["yp_info"] = split[12]
		m2["start_train_date"] = split[13]
		m2["train_seat_feature"] = split[14]
		m2["location_code"] = split[15]
		m2["from_station_no"] = split[16]
		m2["to_station_no"] = split[17]
		m2["is_support_card"] = split[18]
		m2["controlled_train_flag"] = split[19]
		m2["gg_num"] = split[20]
		m2["gr_num"] = split[21]
		m2["qt_num"] = split[22]
		m2["rw_num"] = split[23]
		m2["rz_num"] = split[24]
		m2["tz_num"] = split[25]
		m2["wz_num"] = split[26] // 无座
		m2["yb_num"] = split[27]
		m2["yw_num"] = split[28]
		m2["yz_num"] = split[29]
		m2["ze_num"] = split[30] // 二等座
		m2["zy_num"] = split[31] // 一等座
		m2["swz_num"] = split[32]
		m2["srrb_num"] = split[33]
		m2["yp_ex"] = split[34]
		m2["seat_types"] = split[35]
		m2["exchange_train_flag"] = split[36]
		m2["houbu_train_flag"] = split[37]
		m2["houbu_seat_limit"] = split[38]
		m2["yp_info_new"] = split[39]
		m2["dw_flag"] = split[40]
		m2["stopcheckTime"] = split[41]
		m2["country_flag"] = split[42]
		m2["local_arrive_time"] = split[43]
		m2["local_start_time"] = split[44]
		m2["bed_level_info"] = split[45]
		m2["seat_discount_info"] = split[46]
		m2["sale_time"] = split[47]
		m["queryLeftNewDTO"] = m2
		m2["from_station_name"] = data.Data.Map[split[6]]
		m2["to_station_name"] = data.Data.Map[split[7]]
		da = append(da, m2)
	}
	var res = SearchResult{}
	for idx, i := range da {
		if onlyShowTicket {
			if hasTicket(i) {
				res = append(res, da[idx])
			}
			continue
		}
		res = append(res, da[idx])
	}
	return res
}

func hasTicket(data map[string]string) bool {
	if data["wz_num"] != "无" && data["wz_num"] != "" {
		return true
	}
	if data["zy_num"] != "无" && data["zy_num"] != "" {
		return true
	}
	if data["ze_num"] != "无" && data["ze_num"] != "" {
		return true
	}
	return false
}

var temp, _ = template.New("").Parse(`
{{ range $idx, $data := .Data}}
日期: {{ $data.start_train_date }} 
班次: {{ $data.station_train_code }}
出发站-到达站: {{ $data.from_station_name }} - {{ $data.to_station_name }}
出发/到达时间: {{ $data.start_time }} - {{ $data.arrive_time }}, 历时: {{ $data.lishi }}
一等座: {{ if $data.zy_num }}{{$data.zy_num}}{{else}}无{{end}}
二等座: {{ if $data.ze_num }}{{$data.ze_num}}{{else}}无{{end}}
无座: {{ if $data.wz_num }}{{$data.wz_num}}{{else}}无{{end}}
{{"\n"}}
{{- end }}
`)

type response struct {
	Httpstatus int `json:"httpstatus"`
	Data       struct {
		Result []string          `json:"result"`
		Flag   string            `json:"flag"`
		Level  string            `json:"level"`
		Map    map[string]string `json:"map"`
	} `json:"data"`
	Messages string `json:"messages"`
	Status   bool   `json:"status"`
}
