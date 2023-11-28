package bitget

import (
	"encoding/json"
	"qq/util"
	"strings"
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

func TransUsdt(coin string) (usdt float64, err error) {
	cli := newClient()

	params := make(map[string]string)
	params["symbol"] = coin

	resp, err := cli.DoGet("/api/spot/v1/market/ticker", params)
	if err != nil {
		return 0, err
	}
	var data TransUsdtResp
	json.NewDecoder(strings.NewReader(resp)).Decode(&data)
	return util.ToFloat64(data.Data.SellOne), nil
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
