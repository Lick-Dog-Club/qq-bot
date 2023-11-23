package bitget

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/util/proxy"
	"qq/util/retry"
	"strconv"
	"strings"
	"sync/atomic"
)

var curr atomic.Value

func init() {
	curr.Store(CoinList{})
	features.AddKeyword("bitget", "获取当前开仓币", func(bot bot.Bot, content string) error {
		bot.Send(Get(true))
		return nil
	})
	features.AddKeyword("bg-money", "获取当前资金", func(bot bot.Bot, content string) error {
		var total float64
		retry.Times(3, func() error {
			var err error
			total, err = MoneyTotal()
			return err
		})
		bot.Send(fmt.Sprintf("money: %.2f", total))
		return nil
	})
}

type Coin struct {
	Name  string
	Total float64
}

type CoinList []Coin

func (l CoinList) String() (s string) {
	var slices []string
	for _, coin := range l {
		slices = append(slices, coin.Name)
	}

	return strings.Join(slices, ",")
}

func Get(all bool) string {
	cli := newClient()

	params := make(map[string]string)
	params["productType"] = "umcbl"
	params["marginCoin"] = "USDT"

	resp, err := cli.DoGet("/api/mix/v1/position/allPosition", params)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var res response
	json.NewDecoder(strings.NewReader(resp)).Decode(&res)
	var newList CoinList
	for _, datum := range res.Data {
		float, _ := strconv.ParseFloat(datum.Total, 64)
		if float > 0 {
			newList = append(newList, Coin{
				Name:  datum.Symbol,
				Total: float,
			})
		}
	}
	add := CoinList{}
	miss := CoinList{}
	list := curr.Load().(CoinList)
	for _, s := range newList {
		if !Has(list, s) {
			add = append(add, Coin{
				Name:  s.Name,
				Total: s.Total,
			})
		}
	}
	for _, s := range list {
		if !Has(newList, s) {
			miss = append(miss, s)
		}
	}

	result := ""
	if len(add) > 0 {
		result += fmt.Sprintf("新增: %s ", add)
	}
	if len(miss) > 0 {
		result += fmt.Sprintf("删除: %s", miss)
	}
	curr.Store(newList)
	if all {
		return newList.String()
	}
	return result
}

func newClient() *RestClient {
	cli := &RestClient{
		ApiKey:       config.BgApiKey(),
		ApiSecretKey: config.BgApiSecretKey(),
		Passphrase:   config.BgPassphrase(),
		BaseUrl:      "https://api.bitget.com",
		HttpClient:   *proxy.NewHttpProxyClient(),
		Signer:       new(Signer).Init(config.BgApiSecretKey()),
	}
	return cli
}

func Has(list []Coin, key Coin) bool {
	for _, s := range list {
		if s == key {
			return true
		}
	}
	return false
}

type Signer struct {
	secretKey []byte
}

func (p *Signer) Init(key string) *Signer {
	p.secretKey = []byte(key)
	return p
}

func (p *Signer) Sign(method string, requestPath string, body string, timesStamp string) string {
	var payload strings.Builder
	payload.WriteString(timesStamp)
	payload.WriteString(method)
	payload.WriteString(requestPath)
	if body != "" && body != "?" {
		payload.WriteString(body)
	}
	hash := hmac.New(sha256.New, p.secretKey)
	hash.Write([]byte(payload.String()))
	result := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return result
}

type response struct {
	Code        string `json:"code"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
	Data        []struct {
		MarginCoin        string      `json:"marginCoin"`
		Symbol            string      `json:"symbol"`
		HoldSide          string      `json:"holdSide"`
		OpenDelegateCount string      `json:"openDelegateCount"`
		Margin            string      `json:"margin"`
		Available         string      `json:"available"`
		Locked            string      `json:"locked"`
		Total             string      `json:"total"`
		Leverage          int         `json:"leverage"`
		AchievedProfits   string      `json:"achievedProfits"`
		AverageOpenPrice  string      `json:"averageOpenPrice"`
		MarginMode        string      `json:"marginMode"`
		HoldMode          string      `json:"holdMode"`
		UnrealizedPL      string      `json:"unrealizedPL"`
		LiquidationPrice  string      `json:"liquidationPrice"`
		KeepMarginRate    string      `json:"keepMarginRate"`
		MarketPrice       string      `json:"marketPrice"`
		MarginRatio       interface{} `json:"marginRatio"`
		AutoMargin        string      `json:"autoMargin"`
		CTime             string      `json:"cTime"`
	} `json:"data"`
}
