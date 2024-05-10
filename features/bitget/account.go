package bitget

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"qq/util"
	"strings"
	"time"
)

func MoneyTotal() (float64, error) {
	var (
		total float64
		v     float64
		err   error
	)
	if v, err = AccountFuturesMoney(); err != nil {
		return 0, err
	}
	total += v
	if v, err = AccountSpotMoney(); err != nil {
		return 0, err
	}
	total += v
	return total, nil
}

func AccountFuturesMoney() (usdt float64, err error) {
	cli := newClient()

	params := make(map[string]string)
	params["productType"] = "umcbl"

	resp, err := cli.DoGet("/api/mix/v1/account/accounts", params)
	if err != nil {
		return 0, err
	}
	var data AccountMoneyResp
	var totalUsdt float64
	_ = json.NewDecoder(strings.NewReader(resp)).Decode(&data)
	for _, item := range data.Data {
		totalUsdt += util.ToFloat64(item.UsdtEquity)
	}
	return totalUsdt, nil
}

func AccountSpotMoney() (usdt float64, err error) {
	cli := newClient()
	params := make(map[string]string)

	resp, err := cli.DoGet("/api/spot/v1/account/assets-lite", params)
	if err != nil {
		return 0, err
	}
	var data AccountMoneySpotResp
	json.NewDecoder(strings.NewReader(resp)).Decode(&data)
	var totalUsdt float64
	for _, item := range data.Data {
		var u float64 = 1
		if item.CoinName != "USDT" {
			u, err = TransUsdt(item.CoinName + "USDT_SPBL")
			if err != nil {
				return 0, err
			}
		}
		totalUsdt += (util.ToFloat64(item.Available) + util.ToFloat64(item.Frozen)) * u
	}
	return totalUsdt, nil
}

type CurrentUnFinishOrder struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		UserId           string      `json:"userId"`
		Symbol           string      `json:"symbol"`
		OrderId          string      `json:"orderId"`
		ClientOid        string      `json:"clientOid"`
		PriceAvg         string      `json:"priceAvg"`
		Size             string      `json:"size"`
		OrderType        string      `json:"orderType"`
		Side             string      `json:"side"`
		Status           string      `json:"status"`
		BasePrice        string      `json:"basePrice"`
		BaseVolume       string      `json:"baseVolume"`
		QuoteVolume      string      `json:"quoteVolume"`
		EnterPointSource string      `json:"enterPointSource"`
		TriggerPrice     interface{} `json:"triggerPrice"`
		TpslType         string      `json:"tpslType"`
		CTime            string      `json:"cTime"`
	} `json:"data"`
}

func CurrentUnFinishOrderList(symbol string) (usdt *CurrentUnFinishOrder, err error) {
	cli := newClient()
	params := make(map[string]string)
	if strings.HasSuffix(symbol, "_SPBL") {
		symbol = strings.TrimSuffix(symbol, "_SPBL")
	}
	if !strings.HasSuffix(symbol, "USDT") {
		symbol += "USDT"
	}
	params["symbol"] = symbol
	resp, err := cli.DoGet("/api/v2/spot/trade/unfilled-orders", params)
	if err != nil {
		return nil, err
	}
	var data CurrentUnFinishOrder
	json.NewDecoder(strings.NewReader(resp)).Decode(&data)
	if data.Code != "00000" {
		return nil, errors.New(data.Message)
	}
	return &data, nil
}

func BuySpot(symbol string, price string, totalUsdt float64) (usdt *BuySpotResponse, err error) {
	list, err := CurrentUnFinishOrderList(symbol)
	if err != nil {
		return nil, err
	}
	if len(list.Data) > 0 {
		var plist []string
		for _, datum := range list.Data {
			plist = append(plist, fmt.Sprintf("%v", datum.PriceAvg))
		}
		return nil, fmt.Errorf("%s 当前已有挂单, %v", symbol, plist)
	}
	if totalUsdt <= 0 {
		return nil, fmt.Errorf("totalUsdt must be greater than 0, current is %v", totalUsdt)
	}
	cli := newClient()
	params := make(map[string]string)
	if strings.HasSuffix(symbol, "USDT_SPBL") {
		symbol = strings.TrimSuffix(symbol, "USDT_SPBL")
	}
	params["symbol"] = symbol + "USDT"
	params["side"] = "buy"
	params["orderType"] = "limit"
	params["force"] = "gtc"
	p := util.ToFloat64(price)
	if p > 100 {
		p = math.Floor(p)
		price = fmt.Sprintf("%.0f", p)
	}

	params["price"] = price
	params["size"] = fmt.Sprintf("%.6f", totalUsdt/util.ToFloat64(price))
	if util.ToFloat64(price) <= 0 {
		return nil, fmt.Errorf("price must be greater than 0, current is %v", price)
	}

	toJson, _ := ToJson(params)
	resp, err := cli.DoPost("/api/v2/spot/trade/place-order", toJson)
	if err != nil {
		return nil, err
	}
	var data *BuySpotResponse
	json.NewDecoder(strings.NewReader(resp)).Decode(&data)
	return data, nil
}

type BuySpotResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		OrderId   string `json:"orderId"`
		ClientOid string `json:"clientOid"`
	} `json:"data"`
}

func TransUsdt(coin string) (usdt float64, err error) {
	cli := newClient()

	params := make(map[string]string)
	if !strings.HasSuffix(coin, "USDT_SPBL") {
		coin += "USDT_SPBL"
	}
	params["symbol"] = coin

	resp, err := cli.DoGet("/api/spot/v1/market/ticker", params)
	if err != nil {
		return 0, err
	}
	var data TransUsdtResp
	json.NewDecoder(strings.NewReader(resp)).Decode(&data)
	//indent, _ := json.MarshalIndent(data.Data, "", "  ")
	//fmt.Println(string(indent))
	return util.ToFloat64(data.Data.BuyOne), nil
}

type Line struct {
	Date  string
	Price float64
}

func KLine(coin string, after time.Time) (lines []Line, err error) {
	cli := newClient()

	params := make(map[string]string)
	params["symbol"] = coin
	params["period"] = "1h"
	params["limit"] = "1000"
	params["after"] = fmt.Sprintf("%d", after.UnixMilli())
	resp, err := cli.DoGet("/api/spot/v1/market/candles", params)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var data Kline
	json.NewDecoder(strings.NewReader(resp)).Decode(&data)
	var res []Line
	for _, datum := range data.Data {
		res = append(res, Line{
			Date:  time.UnixMilli(util.ToInt64(datum.Ts)).Local().Format("2006-01-02 15:04:05"),
			Price: util.ToFloat64(datum.Close),
		})
	}
	return res, nil
}

type Kline struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Open     string `json:"open"`
		High     string `json:"high"`
		Low      string `json:"low"`
		Close    string `json:"close"`
		QuoteVol string `json:"quoteVol"`
		BaseVol  string `json:"baseVol"`
		UsdtVol  string `json:"usdtVol"`
		Ts       string `json:"ts"`
	} `json:"data"`
}

type TransUsdtResp struct {
	Code        string `json:"code"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
	Data        struct {
		Symbol    string `json:"symbol"`
		High24H   string `json:"high24h"`
		Low24H    string `json:"low24h"`
		Close     string `json:"close"`
		QuoteVol  string `json:"quoteVol"`
		BaseVol   string `json:"baseVol"`
		UsdtVol   string `json:"usdtVol"`
		Ts        string `json:"ts"`
		BuyOne    string `json:"buyOne"`
		SellOne   string `json:"sellOne"`
		BidSz     string `json:"bidSz"`
		AskSz     string `json:"askSz"`
		OpenUtc0  string `json:"openUtc0"`
		ChangeUtc string `json:"changeUtc"`
		Change    string `json:"change"`
	} `json:"data"`
}

type AccountMoneySpotResp struct {
	Code        string `json:"code"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
	Data        []struct {
		CoinID    int    `json:"coinId"`
		CoinName  string `json:"coinName"`
		Available string `json:"available"`
		Frozen    string `json:"frozen"`
		Lock      string `json:"lock"`
		UTime     string `json:"uTime"`
	} `json:"data"`
}

type AccountMoneyResp struct {
	Code        string `json:"code"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
	Data        []struct {
		MarginCoin           string      `json:"marginCoin"`
		Locked               string      `json:"locked"`
		Available            string      `json:"available"`
		CrossMaxAvailable    string      `json:"crossMaxAvailable"`
		FixedMaxAvailable    string      `json:"fixedMaxAvailable"`
		MaxTransferOut       string      `json:"maxTransferOut"`
		Equity               string      `json:"equity"`
		UsdtEquity           string      `json:"usdtEquity"`
		BtcEquity            string      `json:"btcEquity"`
		CrossRiskRate        string      `json:"crossRiskRate"`
		UnrealizedPL         string      `json:"unrealizedPL"`
		Bonus                string      `json:"bonus"`
		CrossedUnrealizedPL  interface{} `json:"crossedUnrealizedPL"`
		IsolatedUnrealizedPL interface{} `json:"isolatedUnrealizedPL"`
	} `json:"data"`
}
