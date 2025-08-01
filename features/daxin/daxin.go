package daxin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qq/bot"
	"qq/features"
	"strings"
)

func init() {
	features.AddKeyword("dx", "股票新债/打新，查询今日是否有新债", func(bot bot.Bot, content string) error {
		get, _ := Get()
		bot.Send(get)
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: nil,
		Call: func(args string) (string, error) {
			get, _ := Get()
			return get, nil
		},
	}))
}

/*
[
	{
	      "code": "370258",
	      "name": "精锻发债",
	      "conversionPrice": "13.09",
	      "stockCode": "300258",
	      "stockName": "精锻科技",
	      "stockPrice": "13.08",
	      "bondPrice": "100.00",
	      "premium": "0.08"
	}
]
*/

func Get() (string, bool) {
	request, _ := http.NewRequest("GET", "https://eq.10jqka.com.cn/mobileuserinfo/app/purchaseIcloud/data/newBondList.json", nil)
	request.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	var data response
	json.NewDecoder(resp.Body).Decode(&data)
	if len(data) == 0 {
		return "今日无新债", false
	}
	var sli = make([]string, 0, len(data))
	for _, item := range data {
		sli = append(sli, item.Name)
	}

	return fmt.Sprintf("今日有新债:\n%s", strings.Join(sli, "\n")), true
}

type response []struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	ConversionPrice string `json:"conversionPrice"`
	StockCode       string `json:"stockCode"`
	StockName       string `json:"stockName"`
	StockPrice      string `json:"stockPrice"`
	BondPrice       string `json:"bondPrice"`
	Premium         string `json:"premium"`
}
