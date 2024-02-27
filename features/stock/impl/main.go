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
		Define: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
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
		Define: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "GetFinancialStatements",
				Description: "获取公司财务报表数据，返回了： '每股收益', '营业收入（元）', '营业收入去年同期（元）', '营业收入同比增长(%）', '营业收入季度环比增长(％）', '净利润（元）', '净利润去年同期（元）', '净利润同比增长(%）', '净利润季度环比增长(％）', '每股净资产（元）', '净资产收益率(%）' 的信息",
				Parameters: &jsonschema.Definition{
					Type:     jsonschema.Object,
					Required: []string{"ticker"},
					Properties: map[string]jsonschema.Definition{
						"ticker": {
							Type:        jsonschema.String,
							Description: "A股股票代码，例如: 000001,000002",
						},
					},
				},
			},
		},
	}
	ToolsGetIndustryData = tools.Tool{
		Name: "GetIndustryData",
		Define: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "GetIndustryData",
				Description: "获取特定行业数据, 返回 行业代码,行业名称,股票数(只),市价总值(元),平均市盈率,平均价格(元)",
				Parameters: &jsonschema.Definition{
					Type: jsonschema.Object,
				},
			},
		},
	}
	ToolsGetCashFlow = tools.Tool{
		Name: "GetCashFlow",
		Define: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "GetCashFlow",
				Description: "获取现金流数据, 返回 '净现金流（元）','净现金流同比(%）','经营性现金流量净额（元）','经营性现金流量净额占比(％)','客户及同业存款净增加额,金额（元）','客户及同业存款净增加额,占比(%）','投资性现金流量净额（元）','投资性现金流量净额净额占比(%)','贷款增加额金额（元）','贷款增加额占比(%）','取得投资收益收到的现金金额（元）','取得投资收益收到的现金占比(%）','融资性现金流量净额（元）','融资性现金流量净额占比(%）'",
				Parameters: &jsonschema.Definition{
					Type:     jsonschema.Object,
					Required: []string{"ticker"},
					Properties: map[string]jsonschema.Definition{
						"ticker": {
							Type:        jsonschema.String,
							Description: "A股股票代码，例如: 000001,000002",
						},
					},
				},
			},
		},
	}
	ToolsGetMarketSentiment = tools.Tool{
		Name: "GetMarketSentiment",
		Define: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
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
	ToolsGetCodeByName = tools.Tool{
		Name: "GetCodeByName",
		Define: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "GetCodeByName",
				Description: "根据名称获取股票代码",
				Parameters: &jsonschema.Definition{
					Type:     jsonschema.Object,
					Required: []string{"name"},
					Properties: map[string]jsonschema.Definition{
						"name": {
							Type:        jsonschema.String,
							Description: "股票名称, 例如 中国平安，浪潮信息",
						},
					},
				},
			},
		},
	}
)

const ApiAddrPrefix = "http://localhost:8080/api/public"

func CallTool(name string, args string) (string, error) {
	switch name {
	case "GetMarketSentiment":
		var input GetMarketSentimentRequest
		json.NewDecoder(strings.NewReader(args)).Decode(&input)
		return GetMarketSentiment(input), nil
	case "GetFinancialStatements":
		var input GetFinancialStatementsRequest
		json.NewDecoder(strings.NewReader(args)).Decode(&input)
		return GetFinancialStatements(input), nil
	case "GetCashFlow":
		var input GetCashFlowRequest
		json.NewDecoder(strings.NewReader(args)).Decode(&input)
		return GetCashFlow(input), nil
	case "GetIndustryData":
		return GetIndustryData(), nil
	case "GetStockPrice":
		var input GetStockPriceRequest
		json.NewDecoder(strings.NewReader(args)).Decode(&input)
		price := GetStockPrice(input)
		marshal, _ := json.Marshal(price)
		return string(marshal), nil
	case "GetCodeByName":
		var s = struct {
			Name string `json:"name"`
		}{}
		json.Unmarshal([]byte(args), &s)
		return GetCodeByName(s.Name), nil
	case tools.BuildInPluginCurrentDatetime.Name:
		return time.Now().Format(time.DateTime), nil
	}
	return "", fmt.Errorf("plugin call '%s' not impl", name)
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

// 获取特定股票价格信息
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
	Ticker string `json:"ticker"` // 股票代码
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
type FData struct {
	SECURITYCODE         string  `json:"SECURITY_CODE"`
	SECURITYNAMEABBR     string  `json:"SECURITY_NAME_ABBR"`
	TRADEMARKET          string  `json:"TRADE_MARKET"`
	TRADEMARKETCODE      string  `json:"TRADE_MARKET_CODE"`
	SECURITYTYPE         string  `json:"SECURITY_TYPE"`
	SECURITYTYPECODE     string  `json:"SECURITY_TYPE_CODE"`
	UPDATEDATE           string  `json:"UPDATE_DATE"`
	REPORTDATE           string  `json:"REPORT_DATE"`
	BASICEPS             float64 `json:"BASIC_EPS"`
	TOTALOPERATEINCOME   int64   `json:"TOTAL_OPERATE_INCOME"`
	TOTALOPERATEINCOMESQ int64   `json:"TOTAL_OPERATE_INCOME_SQ"`
	PARENTNETPROFIT      int64   `json:"PARENT_NETPROFIT"`
	PARENTNETPROFITSQ    int64   `json:"PARENT_NETPROFIT_SQ"`
	PARENTBVPS           float64 `json:"PARENT_BVPS"`
	WEIGHTAVGROE         float64 `json:"WEIGHTAVG_ROE"`
	YSTZ                 float64 `json:"YSTZ"`
	JLRTBZCL             float64 `json:"JLRTBZCL"`
	DJDYSHZ              float64 `json:"DJDYSHZ"`
	DJDJLHZ              float64 `json:"DJDJLHZ"`
	PUBLISHNAME          string  `json:"PUBLISHNAME"`
	ORGCODE              string  `json:"ORG_CODE"`
	NOTICEDATE           string  `json:"NOTICE_DATE"`
	QDATE                string  `json:"QDATE"`
	DATATYPE             string  `json:"DATATYPE"`
	MARKET               string  `json:"MARKET"`
	ISNEW                string  `json:"ISNEW"`
	EITIME               string  `json:"EITIME"`
	SECUCODE             string  `json:"SECUCODE"`
}
type FinancialResp struct {
	Version string `json:"version"`
	Result  struct {
		Pages int     `json:"pages"`
		Data  []FData `json:"data"`
		Count int     `json:"count"`
	} `json:"result"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// 获取公司财务报表数据
func GetFinancialStatements(input GetFinancialStatementsRequest) string {
	client := &http.Client{}
	uv := url.Values{}

	uv.Set("callback", "")
	uv.Set("sortColumns", "REPORT_DATE")
	uv.Set("sortTypes", "-1")
	uv.Set("pageSize", "50")
	uv.Set("pageNumber", "1")
	uv.Set("columns", "ALL")
	uv.Set("filter", fmt.Sprintf(`(SECURITY_CODE="%v")`, input.Ticker))
	uv.Set("reportName", "RPT_FCI_PERFORMANCEE")
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?%s", uv.Encode()), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", fmt.Sprintf("https://data.eastmoney.com/bbsj/yjbb/%v.html", input.Ticker))
	req.Header.Set("Sec-Fetch-Dest", "script")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
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
	var data FinancialResp
	json.NewDecoder(strings.NewReader(string(bodyText))).Decode(&data)
	i := lo.Map(data.Result.Data, func(item FData, index int) map[string]any {
		return map[string]any{
			"截止日期":          item.REPORTDATE,
			"每股收益":          item.BASICEPS,
			"营业收入（元）":       item.TOTALOPERATEINCOME,
			"营业收入去年同期（元）":   item.TOTALOPERATEINCOMESQ,
			"营业收入同比增长(%）":   item.YSTZ,
			"营业收入季度环比增长(％）": item.DJDYSHZ,
			"净利润（元）":        item.PARENTNETPROFIT,
			"净利润去年同期（元）":    item.PARENTNETPROFITSQ,
			"净利润同比增长(%）":    item.JLRTBZCL,
			"净利润季度环比增长(％）":  item.DJDJLHZ,
			"每股净资产（元）":      item.PARENTBVPS,
			"净资产收益率(%）":     item.WEIGHTAVGROE,
		}
	})
	marshal, _ := json.Marshal(&i)
	return string(marshal)
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

// 获取股票成交量数据
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

// 获取市场数据
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

// 股票基本面数据
func GetStockFundamentals(GetStockFundamentalsRequest) GetStockFundamentalsResponse {
	panic("")
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

type GetCashFlowRequest struct {
	Ticker string `json:"ticker"`
}

// 现金流
func GetCashFlow(input GetCashFlowRequest) string {
	client := &http.Client{}

	uv := url.Values{}

	uv.Set("callback", ``)
	uv.Set("sortColumns", `REPORT_DATE`)
	uv.Set("sortTypes", `-1`)
	uv.Set("pageSize", `50`)
	uv.Set("pageNumber", `1`)
	uv.Set("columns", `ALL`)
	uv.Set("filter", fmt.Sprintf(`(SECURITY_CODE="%v")`, input.Ticker))
	uv.Set("reportName", `RPT_DMSK_FN_CASHFLOW`)

	req, err := http.NewRequest("GET", fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?%s", uv.Encode()), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://data.eastmoney.com/bbsj/yjbb/000001.html")
	req.Header.Set("Sec-Fetch-Dest", "script")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
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
	var data GetCashFlowResponse
	json.NewDecoder(strings.NewReader(string(bodyText))).Decode(&data)
	i := lo.Map(data.Result.Data, func(item CashFlowData, index int) map[string]any {
		return map[string]any{
			"净现金流（元）":           item.CCEADD,
			"净现金流同比(%）":         item.CCEADDRATIO,
			"经营性现金流量净额（元）":      item.NETCASHOPERATE,
			"经营性现金流量净额占比(％)":    item.NETCASHOPERATERATIO,
			"客户及同业存款净增加额,金额（元）": item.DEPOSITIOFIOTHER,
			"客户及同业存款净增加额,占比(%）": item.DIORATIO,
			"投资性现金流量净额（元）":      item.NETCASHINVEST,
			"投资性现金流量净额净额占比(%)":  item.NETCASHFINANCERATIO,
			"贷款增加额金额（元）":        item.LOANADVANCEADD,
			"贷款增加额占比(%）":        item.LAARATIO,
			"取得投资收益收到的现金金额（元）":  item.RECEIVEINVESTINCOME,
			"取得投资收益收到的现金占比(%）":  item.RIIRATIO,
			"融资性现金流量净额（元）":      item.NETCASHFINANCE,
			"融资性现金流量净额占比(%）":    item.NETCASHINVESTRATIO,
		}
	})
	marshal, _ := json.Marshal(&i)
	return string(marshal)
}

type CashFlowData struct {
	SECUCODE                  string      `json:"SECUCODE"`
	SECURITYCODE              string      `json:"SECURITY_CODE"`
	INDUSTRYCODE              string      `json:"INDUSTRY_CODE"`
	ORGCODE                   string      `json:"ORG_CODE"`
	SECURITYNAMEABBR          string      `json:"SECURITY_NAME_ABBR"`
	INDUSTRYNAME              string      `json:"INDUSTRY_NAME"`
	MARKET                    string      `json:"MARKET"`
	SECURITYTYPECODE          string      `json:"SECURITY_TYPE_CODE"`
	TRADEMARKETCODE           string      `json:"TRADE_MARKET_CODE"`
	DATETYPECODE              string      `json:"DATE_TYPE_CODE"`
	REPORTTYPECODE            string      `json:"REPORT_TYPE_CODE"`
	DATASTATE                 string      `json:"DATA_STATE"`
	NOTICEDATE                string      `json:"NOTICE_DATE"`
	REPORTDATE                string      `json:"REPORT_DATE"`
	NETCASHOPERATE            int64       `json:"NETCASH_OPERATE"`
	NETCASHOPERATERATIO       float64     `json:"NETCASH_OPERATE_RATIO"`
	SALESSERVICES             interface{} `json:"SALES_SERVICES"`
	SALESSERVICESRATIO        interface{} `json:"SALES_SERVICES_RATIO"`
	PAYSTAFFCASH              int64       `json:"PAY_STAFF_CASH"`
	PSCRATIO                  float64     `json:"PSC_RATIO"`
	NETCASHINVEST             int64       `json:"NETCASH_INVEST"`
	NETCASHINVESTRATIO        float64     `json:"NETCASH_INVEST_RATIO"`
	RECEIVEINVESTINCOME       int64       `json:"RECEIVE_INVEST_INCOME"`
	RIIRATIO                  float64     `json:"RII_RATIO"`
	CONSTRUCTLONGASSET        int64       `json:"CONSTRUCT_LONG_ASSET"`
	CLARATIO                  float64     `json:"CLA_RATIO"`
	NETCASHFINANCE            int64       `json:"NETCASH_FINANCE"`
	NETCASHFINANCERATIO       float64     `json:"NETCASH_FINANCE_RATIO"`
	CCEADD                    int64       `json:"CCE_ADD"`
	CCEADDRATIO               float64     `json:"CCE_ADD_RATIO"`
	CUSTOMERDEPOSITADD        *int        `json:"CUSTOMER_DEPOSIT_ADD"`
	CDARATIO                  *int        `json:"CDA_RATIO"`
	DEPOSITIOFIOTHER          *int64      `json:"DEPOSIT_IOFI_OTHER"`
	DIORATIO                  *float64    `json:"DIO_RATIO"`
	LOANADVANCEADD            int64       `json:"LOAN_ADVANCE_ADD"`
	LAARATIO                  float64     `json:"LAA_RATIO"`
	RECEIVEINTERESTCOMMISSION interface{} `json:"RECEIVE_INTEREST_COMMISSION"`
	RICRATIO                  interface{} `json:"RIC_RATIO"`
	INVESTPAYCASH             interface{} `json:"INVEST_PAY_CASH"`
	IPCRATIO                  interface{} `json:"IPC_RATIO"`
	BEGINCCE                  interface{} `json:"BEGIN_CCE"`
	BEGINCCERATIO             interface{} `json:"BEGIN_CCE_RATIO"`
	ENDCCE                    interface{} `json:"END_CCE"`
	ENDCCERATIO               interface{} `json:"END_CCE_RATIO"`
	RECEIVEORIGICPREMIUM      interface{} `json:"RECEIVE_ORIGIC_PREMIUM"`
	ROPRATIO                  interface{} `json:"ROP_RATIO"`
	PAYORIGICCOMPENSATE       interface{} `json:"PAY_ORIGIC_COMPENSATE"`
	POCRATIO                  interface{} `json:"POC_RATIO"`
}
type GetCashFlowResponse struct {
	Version string `json:"version"`
	Result  struct {
		Pages int            `json:"pages"`
		Data  []CashFlowData `json:"data"`
		Count int            `json:"count"`
	} `json:"result"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func GetCodeByName(name string) string {
	client := &http.Client{}
	uv := url.Values{}
	uv.Set("client", "web")
	uv.Set("clientType", "webSuggest")
	uv.Set("clientVersion", "lastest")
	uv.Set("keyword", name)
	uv.Set("pageIndex", "1")
	uv.Set("pageSize", "10")
	uv.Set("securityFilter", "")
	uv.Set("_", fmt.Sprintf("%d", time.Now().UnixMilli()))
	req, _ := http.NewRequest("GET", "https://search-codetable.eastmoney.com/codetable/search/web?"+uv.Encode(), nil)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://www.eastmoney.com/")
	req.Header.Set("Sec-Fetch-Dest", "script")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not A(Brand";v="99", "Google Chrome";v="121", "Chromium";v="121"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	bodyText, _ := io.ReadAll(resp.Body)
	var data response
	if err := json.NewDecoder(strings.NewReader(string(bodyText))).Decode(&data); err != nil {
		log.Fatal(err)
	}
	var result []struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}
	for _, datum := range data.Result {
		result = append(result, struct {
			Name string `json:"name"`
			Code string `json:"code"`
		}{Name: datum.ShortName, Code: datum.Code})
	}
	marshal, _ := json.Marshal(&result)
	return string(marshal)
}

type response struct {
	Code      string `json:"code"`
	Msg       string `json:"msg"`
	PageIndex int    `json:"pageIndex"`
	PageSize  int    `json:"pageSize"`
	Result    []struct {
		Code             string `json:"code"`
		InnerCode        string `json:"innerCode"`
		ShortName        string `json:"shortName"`
		Market           int    `json:"market"`
		Pinyin           string `json:"pinyin"`
		SecurityType     []int  `json:"securityType"`
		SecurityTypeName string `json:"securityTypeName"`
		SmallType        int    `json:"smallType"`
		Status           int    `json:"status"`
		Flag             int    `json:"flag"`
		ExtSmallType     int    `json:"extSmallType"`
	} `json:"result"`
	SearchId string `json:"searchId"`
}
