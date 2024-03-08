package taobaoprice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"
	"strings"
)

func init() {
	features.AddKeyword("staobao", "<+id> 查询淘宝价格", func(bot bot.Bot, content string) error {
		search, err := Search(strings.Trim(content, ""))
		if err != nil {
			bot.Send(err.Error())
			return nil
		}
		bot.Send(search.String())
		return nil
	}, features.WithHidden())
}

func Search(numIid string) (config.Skus, error) {
	var (
		key    = config.OKey()
		secret = config.OSecret()
	)
	fmt.Println(fmt.Sprintf("https://api-gw.onebound.cn/taobao/item_get/?key=%s&&num_iid=%s&is_promotion=1&&lang=zh-CN&secret=%s", key, numIid, secret))
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api-gw.onebound.cn/taobao/item_get/?key=%s&&num_iid=%s&is_promotion=1&&lang=zh-CN&secret=%s", key, numIid, secret), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data response
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}
	if data.ErrorCode != "0000" {
		return nil, errors.New("onebound 接口查询失败" + numIid)
	}
	var skus = config.Skus{}
	result := data.Item.Skus.Sku
	for idx := range result {
		s := result[idx]
		var skumap = make(map[int64]config.Sku)
		if m, ok := skus[data.Item.NumIid]; ok {
			skumap = m
		}
		skumap[s.SkuID] = config.Sku{
			NumIID:        data.Item.NumIid,
			Title:         data.Item.Title,
			SkuID:         s.SkuID,
			Price:         s.Price,
			OriginalPrice: s.OrginalPrice,
		}
		skus[data.Item.NumIid] = skumap
	}

	return skus, nil
}

type response struct {
	Item struct {
		NumIid int64  `json:"num_iid"`
		Title  string `json:"title"`
		Skus   struct {
			Sku []struct {
				Price        float64 `json:"price"`
				TotalPrice   float64 `json:"total_price"`
				OrginalPrice float64 `json:"orginal_price"`
				SkuID        int64   `json:"sku_id"`
			} `json:"sku"`
		} `json:"skus"`
	} `json:"item"`
	Error           string `json:"error"`
	Secache         string `json:"secache"`
	SecacheTime     int    `json:"secache_time"`
	SecacheDate     string `json:"secache_date"`
	TranslateStatus string `json:"translate_status"`
	TranslateTime   int    `json:"translate_time"`
	Reason          string `json:"reason"`
	ErrorCode       string `json:"error_code"`
	Cache           int    `json:"cache"`
	APIInfo         string `json:"api_info"`
	ExecutionTime   string `json:"execution_time"`
	ServerTime      string `json:"server_time"`
	ClientIP        string `json:"client_ip"`
	CallArgs        struct {
		NumIid      string `json:"num_iid"`
		IsPromotion string `json:"is_promotion"`
	} `json:"call_args"`
	APIType           string `json:"api_type"`
	TranslateLanguage string `json:"translate_language"`
	TranslateEngine   string `json:"translate_engine"`
	ServerMemory      string `json:"server_memory"`
	RequestID         string `json:"request_id"`
	LastID            string `json:"last_id"`
}
