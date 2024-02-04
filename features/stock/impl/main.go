package impl

import (
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"net/http"
	"net/url"
	"qq/features/stock/tools"
	"time"
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
)

const ApiAddrPrefix = "http://localhost:8080/api/public"

func init() {
	tools.MustRegister(
		ToolGetStockPrice,
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

// GetIndustryDataRequest 是获取特定行业数据的请求参数
type GetIndustryDataRequest struct {
	IndustryName string `json:"industryName"` // 行业名称
}

// IndustryData 描述了一个特定行业的数据
type IndustryData struct {
	IndustryName string  `json:"industryName"` // 行业名称
	PE           float64 `json:"pe"`           // 行业平均市盈率
	PB           float64 `json:"pb"`           // 行业平均市净率
}

// GetIndustryDataResponse 是行业数据的响应结构
type GetIndustryDataResponse struct {
	Industries []IndustryData `json:"industries"` // 行业数据列表
}

func GetIndustryData(GetIndustryDataRequest) GetIndustryDataResponse {
	// TODO
	return GetIndustryDataResponse{}
}

// GetMarketSentimentRequest 是获取市场情绪数据的请求参数
type GetMarketSentimentRequest struct {
	Date string `json:"date"` // 查询日期
}

// GetMarketSentimentResponse 是市场情绪数据的响应结构
type GetMarketSentimentResponse struct {
	SentimentIndex float64 `json:"sentimentIndex"` // 市场情绪指数
}

func GetMarketSentiment(GetMarketSentimentRequest) GetMarketSentimentResponse {
	// TODO
	return GetMarketSentimentResponse{}
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
