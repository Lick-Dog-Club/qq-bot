package lpr

import (
	"bytes"
	"fmt"
	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"qq/bot"
	"qq/features"
	"qq/util"
	"strings"
	"time"
)

func init() {
	features.AddKeyword("lpr", "获取当前贷款市场报价利率(lpr)数值", func(bot bot.Bot, content string) error {
		bot.SendTextImage(Get().String())
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Call: func(args string) (string, error) {
			return Get().String(), nil
		},
	}))
}

type LPRs []LPR

func (r LPRs) String() string {
	bf := bytes.Buffer{}
	for idx := range r {
		lpr := r[idx]
		bf.WriteString(fmt.Sprintf("%s %v%% %v%%\n", lpr.Date.Format("2006-01-02"), lpr.OneYear, lpr.FiveYear))
	}
	return bf.String()
}

type LPR struct {
	Date     time.Time
	OneYear  float64
	FiveYear float64
}

func Get() LPRs {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.bankofchina.com/fimarkets/lilv/fd32/201310/t20131031_2591219.html", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://www.bankofchina.com/fimarkets/lilv/")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not A(Brand";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
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
	parse, _ := htmlquery.Parse(bytes.NewReader(bodyText))
	all, _ := htmlquery.QueryAll(parse, "//tbody/tr")
	l := LPRs{}
	for idx, node := range all {
		expr := "//td"
		if idx == 0 {
			expr = "//th"
			continue
		}
		queryAll, _ := htmlquery.QueryAll(node, expr)
		date, _ := time.ParseInLocation("2006-01-02", htmlquery.InnerText(queryAll[0]), time.Local)
		oneYear := strings.TrimRight(htmlquery.InnerText(queryAll[1]), "%")
		fiveYear := strings.TrimRight(htmlquery.InnerText(queryAll[2]), "%")
		l = append(l, LPR{
			Date:     date,
			OneYear:  util.ToFloat64(oneYear),
			FiveYear: util.ToFloat64(fiveYear),
		})
	}
	return l
}
