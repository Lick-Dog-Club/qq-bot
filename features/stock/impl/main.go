package impl

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"qq/features/stock/tools"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	log "github.com/sirupsen/logrus"
)

var (
	ToolGetStockPrice = tools.Tool{
		Name: "GetStockPrice",
		Type: tools.ToolTypeBuildIn,
		Define: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "GetStockPrice",
				Description: "获取特定股票价格信息，返回了：`日期`，`开盘`，`收盘`，`最高`，`最低`，`成交量` 的信息",
				Parameters: &jsonschema.Definition{
					Type:     jsonschema.Object,
					Required: []string{"adjust", "ticker", "start_date", "end_date"},
					Properties: map[string]jsonschema.Definition{
						"adjust": {
							Type:        jsonschema.String,
							Description: "前复权 (Forward Adjusted): `qfq`，后复权 (Backward Adjusted): `hfq`",
							Enum:        []string{"qfq", "hfq"},
						},
						"ticker": {
							Type:        jsonschema.String,
							Description: "A股股票代码，例如: 000001,000002",
						},
						"start_date": {
							Type:        jsonschema.String,
							Description: "开始时间, 格式: 2024-01-02",
						},
						"end_date": {
							Type:        jsonschema.String,
							Description: "结束时间, 格式: 2024-01-02",
						},
					},
				},
			},
		},
	}

	ToolsGetFinancialStatements = tools.Tool{
		Name: "GetFinancialStatements",
		Type: tools.ToolTypeBuildIn,
		Define: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "GetFinancialStatements",
				Description: "获取公司财务报表数据，返回了：`收入`, `净利润`, `总资产`, `总负债`, `股东权益` 的信息",
				Parameters: &jsonschema.Definition{
					Type:     jsonschema.Object,
					Required: []string{"ticker", "year", "quarter"},
					Properties: map[string]jsonschema.Definition{
						"ticker": {
							Type:        jsonschema.String,
							Description: "A股股票代码，例如: 000001,000002",
						},
						"year": {
							Type:        jsonschema.String,
							Description: "年份，比如: 2024",
						},
						"quarter": {
							Type:        jsonschema.String,
							Description: "季度",
						},
					},
				},
			},
		},
	}
	ToolsGetIndustryData = tools.Tool{
		Name: "GetIndustryData",
		Type: tools.ToolTypeBuildIn,
		Define: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "GetIndustryData",
				Description: "获取特定行业数据, 返回 行业代码,行业名称,股票数(只),市价总值(元),平均市盈率,平均价格(元)",
				Parameters: &jsonschema.Definition{
					Type: jsonschema.Object,
				},
			},
		},
	}
	ToolsGetMarketSentiment = tools.Tool{
		Name: "GetMarketSentiment",
		Type: tools.ToolTypeBuildIn,
		Define: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "GetMarketSentiment",
				Description: "获取市场信心指数，返回近一周/一个月/一年的指数",
				Parameters: &jsonschema.Definition{
					Type:     jsonschema.Object,
					Required: []string{"period"},
					Properties: map[string]jsonschema.Definition{
						"period": {
							Type:        jsonschema.String,
							Description: "周期, 'WEEK' 'MONTH' 'YEAR'",
							Enum:        []string{"WEEK", "MONTH", "YEAR"},
						},
					},
				},
			},
		},
	}
)

const ApiAddrPrefix = "http://localhost:8080/api/public"

func init() {
	tools.MustRegister(
		ToolGetStockPrice,
		ToolsGetIndustryData,
		ToolsGetMarketSentiment,
		//ToolsGetFinancialStatements,
	)
}

// GetStockPriceRequest 是获取特定股票价格信息的请求参数
type GetStockPriceRequest struct {
	Ticker    string `json:"ticker"` // 股票代码
	StartDate MyTime `json:"start_date"`
	EndDate   MyTime `json:"end_date"`
	Adjust    string `json:"adjust"`
}
type MyTime string

func (m MyTime) Format(s string) string {
	parse, _ := time.Parse("2006-01-02", string(m))
	return parse.Format(s)
}

// GetStockPriceResponse 是获取特定股票价格信息的响应数据
type GetStockPriceResponse struct {
	Date   string  `json:"日期"`  // 日期
	Open   float64 `json:"开盘"`  // 开盘价
	Close  float64 `json:"收盘"`  // 收盘价
	High   float64 `json:"最高"`  // 最高价
	Low    float64 `json:"最低"`  // 最低价
	Volume int64   `json:"成交量"` // 成交量
}

func GetStockPrice(req GetStockPriceRequest) []GetStockPriceResponse {
	uv := url.Values{}
	uv.Set("symbol", req.Ticker)
	uv.Set("period", "daily")
	uv.Set("start_date", req.StartDate.Format("20060102"))
	uv.Set("end_date", req.EndDate.Format("20060102"))
	uv.Set("adjust", req.Adjust)
	fmt.Println(uv.Encode())
	resp, _ := http.Get(ApiAddrPrefix + "/stock_zh_a_hist?" + uv.Encode())
	defer resp.Body.Close()
	var data []GetStockPriceResponse
	json.NewDecoder(resp.Body).Decode(&data)
	return data
}

// GetFinancialStatementsRequest 是获取公司财务报表数据的请求参数
type GetFinancialStatementsRequest struct {
	Ticker  string `json:"ticker"`  // 股票代码
	Year    int    `json:"year"`    // 年份
	Quarter int    `json:"quarter"` // 季度
}

// FinancialStatement 包含了财务报表的关键数据
type FinancialStatement struct {
	Revenue          float64 `json:"revenue"`          // 收入
	NetIncome        float64 `json:"netIncome"`        // 净利润
	TotalAssets      float64 `json:"totalAssets"`      // 总资产
	TotalLiabilities float64 `json:"totalLiabilities"` // 总负债
	Equity           float64 `json:"equity"`           // 股东权益
}

// GetFinancialStatementsResponse 是财务报表数据的响应结构
type GetFinancialStatementsResponse struct {
	Ticker             string             `json:"ticker"`             // 股票代码
	FinancialStatement FinancialStatement `json:"financialStatement"` // 财务报表数据
}

func GetFinancialStatements(GetFinancialStatementsRequest) GetFinancialStatementsResponse {
	// TODO
	return GetFinancialStatementsResponse{}
}

// GetVolumeDataRequest 是获取股票成交量数据的请求参数
type GetVolumeDataRequest struct {
	Ticker string `json:"ticker"` // 股票代码
	Date   string `json:"date"`   // 查询日期
}

// GetVolumeDataResponse 是股票成交量数据的响应结构
type GetVolumeDataResponse struct {
	Ticker string `json:"ticker"` // 股票代码
	Date   string `json:"date"`   // 日期
	Volume int64  `json:"volume"` // 成交量
}

func GetVolumeData(GetVolumeDataRequest) GetVolumeDataResponse {
	// TODO
	return GetVolumeDataResponse{}
}

// GetMarketDataRequest 是获取市场数据的请求参数
type GetMarketDataRequest struct {
	Date string `json:"date"` // 查询日期
}

// IndexData 描述了一个市场指数的数据
type IndexData struct {
	IndexName string  `json:"indexName"` // 指数名称
	Close     float64 `json:"close"`     // 收盘点数
	Change    float64 `json:"change"`    // 变动百分比
}

// GetMarketDataResponse 是市场数据的响应结构
type GetMarketDataResponse struct {
	Date    string      `json:"date"`    // 日期
	Indices []IndexData `json:"indices"` // 指数数据列表
}

func GetMarketData(GetMarketDataRequest) GetMarketDataResponse {
	// TODO
	return GetMarketDataResponse{}
}

// GetStockFundamentalsRequest 是获取股票基本面数据的请求参数
type GetStockFundamentalsRequest struct {
	Ticker string `json:"ticker"` // 股票代码
}

// GetStockFundamentalsResponse 是股票基本面数据的响应结构
type GetStockFundamentalsResponse struct {
	Ticker        string  `json:"ticker"`        // 股票代码
	PE            float64 `json:"pe"`            // 市盈率
	PB            float64 `json:"pb"`            // 市净率
	DividendYield float64 `json:"dividendYield"` // 股息率
}

func GetStockFundamentals(GetStockFundamentalsRequest) GetStockFundamentalsResponse {
	// TODO
	return GetStockFundamentalsResponse{}
}

// GetHistoricalDataRequest 是获取股票历史价格数据的请求参数
type GetHistoricalDataRequest struct {
	Ticker string `json:"ticker"` // 股票代码
	From   string `json:"from"`   // 开始日期
	To     string `json:"to"`     // 结束日期
}

// HistoricalPrice 描述了股票在某一日期的收盘价
type HistoricalPrice struct {
	Date  string  `json:"date"`  // 日期
	Close float64 `json:"close"` // 收盘价
}

// GetHistoricalDataResponse 是股票历史价格数据的响应结构
type GetHistoricalDataResponse struct {
	Ticker string            `json:"ticker"` // 股票代码
	Prices []HistoricalPrice `json:"prices"` // 历史价格列表
}

func GetHistoricalData(GetHistoricalDataRequest) GetHistoricalDataResponse {
	// TODO
	return GetHistoricalDataResponse{}
}

type IndustryData struct {
	CSRCCODE   string `json:"CSRC_CODE"`
	TOTALVALUE string `json:"TOTAL_VALUE"`
	CSRCNAME   string `json:"CSRC_NAME"`
	AVGPERATE  string `json:"AVG_PE_RATE"`
	AVGPRICE   string `json:"AVG_PRICE"`
	TRADENUM   string `json:"TRADE_NUM"`
	TRADEDATE  string `json:"TRADE_DATE"`
	LISTNUM    string `json:"LIST_NUM"`
}

type SseResp struct {
	QueryDate string         `json:"queryDate"`
	Result    []IndustryData `json:"result"`
}

func GetIndustryData() string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://query.sse.com.cn/commonQuery.do?&jsonCallBack=jsonpCallback56970893&isPagination=false&sqlId=COMMON_SSE_CP_GPJCTPZ_DQHYFL_HYFL_L&_=%d", time.Now().UnixMilli()), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Proxy-Connection", "keep-alive")
	req.Header.Set("Referer", "http://www.sse.com.cn/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, _ := io.ReadAll(resp.Body)
	index := strings.Index(string(bodyText), "{")
	var data SseResp
	json.NewDecoder(strings.NewReader(string(bodyText)[index:])).Decode(&data)
	i := lo.Map(data.Result, func(item IndustryData, index int) map[string]any {
		return map[string]any{
			"行业代码":    item.CSRCCODE,
			"行业名称":    item.CSRCNAME,
			"股票数(只)":  item.TRADENUM,
			"市价总值(元)": item.TOTALVALUE,
			"平均市盈率":   item.AVGPERATE,
			"平均价格(元)": item.AVGPRICE,
		}
	})
	marshal, _ := json.Marshal(&i)
	return string(marshal)
}

// GetMarketSentimentRequest 是获取市场情绪数据的请求参数
type GetMarketSentimentRequest struct {
	// YEAR WEEK MONTH
	Period string `json:"period"`
}

type GetMarketSentimentData struct {
	Date           string  `json:"tradeDate"`
	SentimentIndex float64 `json:"maIndex1"` // 市场情绪指数
}

func GetMarketSentiment(input GetMarketSentimentRequest) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://sentiment.chinascope.com/inews/senti/index?period=%s&contentType=0&_v=%d", input.Period, time.Now().UnixMilli()), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	//req.Header.Set("Cookie", "inews_session_product=ruscrd71gvmdmpd5bq4ghqeaon")
	req.Header.Set("Referer", "https://sentiment.chinascope.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not A(Brand";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var data []GetMarketSentimentData
	json.NewDecoder(resp.Body).Decode(&data)
	i := lo.Map(data, func(item GetMarketSentimentData, index int) map[string]any {
		return map[string]any{
			"信心指数": item.SentimentIndex,
			"日期":   item.Date,
		}
	})
	marshal, _ := json.Marshal(&i)
	return string(marshal)
}

// GetRegulatoryAnnouncementsRequest 是获取监管公告的请求参数
type GetRegulatoryAnnouncementsRequest struct {
	Ticker string `json:"ticker"` // 股票代码
	Date   string `json:"date"`   // 查询日期
}

// Announcement 描述了一个特定的监管公告
type Announcement struct {
	Date         string `json:"date"`         // 公告日期
	Announcement string `json:"announcement"` // 公告内容
}

// GetRegulatoryAnnouncementsResponse 是监管公告数据的响应结构
type GetRegulatoryAnnouncementsResponse struct {
	Ticker        string         `json:"ticker"`        // 股票代码
	Announcements []Announcement `json:"announcements"` // 监管公告列表
}

func GetRegulatoryAnnouncements(GetRegulatoryAnnouncementsRequest) GetRegulatoryAnnouncementsResponse {
	// TODO
	return GetRegulatoryAnnouncementsResponse{}
}
