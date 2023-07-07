package imaotai

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	mrand "math/rand"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/features/geo"
	"qq/features/util/retry"
	"qq/util"
	"regexp"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/forgoer/openssl"
)

var allShops AllShopMap = getMap()

func init() {
	features.AddKeyword("mt", "<+phoneNum>: 自动预约茅台", func(bot bot.Bot, content string) error {
		res := Run(content)
		fmt.Println(res)
		bot.Send(res)
		return nil
	}, features.WithGroup("maotai"))
	features.AddKeyword("mt-del", "<+phoneNum>: 取消茅台自动预约", func(bot bot.Bot, content string) error {
		config.DelMaoTaiInfo(content)
		bot.Send("成功取消！")
		return nil
	}, features.WithGroup("maotai"))
	features.AddKeyword("mt-jwd", "+<phone> +<lat,lng> 设置经纬度", func(bot bot.Bot, content string) error {
		split := strings.Split(content, " ")
		if len(split) == 2 {
			phone := split[0]
			latlng := strings.Split(split[1], ",")
			if len(latlng) == 2 {
				if info, ok := config.MaoTaiInfoMap()[phone]; ok {
					info.Lat = util.ToFloat64(latlng[0])
					info.Lng = util.ToFloat64(latlng[1])
					config.AddMaoTaiInfo(info)
					bot.Send(fmt.Sprintf(`设置成功：
手机号：%s
Geo：%s
`, phone, split[1]))
					return nil
				}
				bot.Send(`请先登陆之后再设置经纬度：
登陆:
mt %s

设置经纬度:
mt-jwd %s <lat,lng>
`)
				return nil
			}
		}
		bot.Send("输入不合法: " + content)
		return nil
	}, features.WithGroup("maotai"))
	features.AddKeyword("mt-geo", "+<phone> +<地址,高德自动查询 geo> 设置经纬度", func(bot bot.Bot, content string) error {
		split := strings.Split(content, " ")
		if len(split) == 2 {
			phone := split[0]
			geoStr := geo.Geo(config.GeoKey(), split[1])
			latlng := strings.Split(geoStr, ",")
			if len(latlng) == 2 {
				if info, ok := config.MaoTaiInfoMap()[phone]; ok {
					info.Lat = util.ToFloat64(latlng[0])
					info.Lng = util.ToFloat64(latlng[1])
					config.AddMaoTaiInfo(info)
					bot.Send(fmt.Sprintf(`设置成功：
手机号：%s
地址：%s
Geo：%s
`, phone, split[1], geoStr))
					return nil
				}
				bot.Send(`请先登陆之后再设置经纬度：
登陆:
mt %s

设置经纬度:
mt-geo %s <地址>
`)
				return nil
			}
		}
		bot.Send("输入不合法: " + content)
		return nil
	}, features.WithGroup("maotai"))
	features.AddKeyword("mt-list", "当前用户以及过期时间", func(bot bot.Bot, content string) error {
		var res string
		for _, info := range config.MaoTaiInfoMap() {
			res += fmt.Sprintf(`
手机号码：%s
过期时间：%s
经纬度: %f,%f

`, info.Phone, info.ExpireAt.Format(time.DateTime), info.Lat, info.Lng)
		}
		bot.Send(res)
		return nil
	}, features.WithGroup("maotai"))
	features.AddKeyword("mt-login", "<+phone> <+code>: 自动预约茅台", func(bot bot.Bot, content string) error {
		split := strings.Split(content, " ")
		var phone, code string
		if len(split) >= 2 {
			phone = strings.TrimSpace(split[0])
			code = strings.TrimSpace(split[1])
		}
		uid, token := login(phone, code)
		info := config.MaoTaiInfo{
			Phone:    phone,
			Uid:      uid,
			Token:    token,
			ExpireAt: time.Time{},
		}

		if token != "" {
			decodeString, _ := base64.StdEncoding.DecodeString(strings.Split(token, ".")[1])
			var e exp
			json.Unmarshal([]byte(string(decodeString)+"}"), &e)
			info.ExpireAt = time.Unix(e.Exp, 0)
		}
		if info.ExpireAt.IsZero() {
			bot.Send("信息有误，添加失败")
			return nil
		}
		config.AddMaoTaiInfo(info)
		bot.Send(fmt.Sprintf(`
用户添加成功
过期时间是: %s
再次执行命令来申购茅台：

mt %s
`, info.ExpireAt.Format(time.DateTime), info.Phone))
		return nil
	}, features.WithGroup("maotai"))
}

type exp struct {
	Exp int64 `json:"exp"`
}

func Run(m string) string {
	// 1. sessionID
	sessionID := GetCurrentSessionID()

	info, ok := config.MaoTaiInfoMap()[m]
	if (ok && info.Expired()) || !ok {
		getCode(m)
		return fmt.Sprintf("用户未登陆，短信已发送，收到后执行：\n\nmt-login %s <code>", m)
	}

	return doReservation(sessionID, info.Uid, info.Token, LatLng{
		lat: info.Lat,
		lng: info.Lng,
	})
}

type ItemShopResp struct {
	Code int `json:"code"`
	Data struct {
		Shops []struct {
			ShopID string `json:"shopId"`
			Items  []struct {
				ItemID    string `json:"itemId"`
				OwnerName string `json:"ownerName"`
			} `json:"items"`
		} `json:"shops"`
	} `json:"data"`
}

// 以闸弄口为中心
var zhalongkou = LatLng{
	lat: 30.27844,
	lng: 120.184013,
}

type shopInfo struct {
	ID   string
	Name string
}

func getItemShop(url string, itemID int, latLng LatLng) (shopIDs []shopInfo) {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	var data ItemShopResp
	json.NewDecoder(resp.Body).Decode(&data)

	oriLatLng := zhalongkou
	if latLng.lat != 0 && latLng.lng != 0 {
		oriLatLng = latLng
	}
	for _, shop := range data.Data.Shops {
		for _, item := range shop.Items {
			if item.ItemID == fmt.Sprintf("%d", itemID) {
				if addr, ok := allShops[shop.ShopID]; ok {
					dis := GetDistance(oriLatLng, LatLng{
						lat: addr.Lat,
						lng: addr.Lng,
					})
					//log.Printf("店铺: %s, 供应商: %s, 距离是 %v km, id: %v", addr.Name, item.OwnerName, dis, shop.ShopID)
					if dis < 15 {
						shopIDs = append(shopIDs, shopInfo{
							ID:   shop.ShopID,
							Name: addr.Name,
						})
						break
					}
				}
			}
		}
	}
	return shopIDs
}

func doReservation(sessionID, uid int, token string, latLng LatLng) (res string) {
	fmt.Printf("申购：\nsessionID: %v\nuid: %v\ntoken: %v\nlatlng: %v", sessionID, uid, token, latLng)
	// 4. reservation
	//10213 3%vol 500ml贵州茅台酒（癸卯兔年）
	//10214 53%vol 375ml×2贵州茅台酒（癸卯兔年）
	var urls = []string{
		fmt.Sprintf(`https://static.moutai519.com.cn/mt-backend/xhr/front/mall/shop/list/slim/v3/%d/浙江省/10213/%d`, sessionID, today().UnixMilli()),
		fmt.Sprintf(`https://static.moutai519.com.cn/mt-backend/xhr/front/mall/shop/list/slim/v3/%d/浙江省/10214/%d`, sessionID, today().UnixMilli()),
	}
	items := map[int][]shopInfo{
		10213: []shopInfo{},
		10214: []shopInfo{},
	}
	for _, url := range urls {
		shop213 := getItemShop(url, 10213, latLng)
		shop214 := getItemShop(url, 10214, latLng)
		items[10213] = append(items[10213], shop213...)
		items[10214] = append(items[10214], shop214...)
	}
	fmt.Println(items)
	for itemID, shopIDs := range items {
		if len(shopIDs) > 0 {
			shop := shopIDs[mrand.Intn(len(shopIDs))]
			res += reservation(itemID, shop, sessionID, uid, token) + "\n"
		}
	}
	return
}

var (
	AES_KEY = []byte("qbhajinldepmucsonaaaccgypwuvcjaa")
	AES_IV  = []byte("2018534749963515")
	SALT    = "2af72f100c356273d46284f6fd1dfc08"
)

var device = "MFGOYB7G-R5FO-UB1K-H4VN-BAHGQM0COZHU"

func headers() map[string]string {
	return map[string]string{
		"MT-Lat":          "28.499562",
		"MT-K":            fmt.Sprintf("%d", time.Now().Unix()),
		"MT-Lng":          "102.182314",
		"Host":            "app.moutai519.com.cn",
		"MT-User-Tag":     "0",
		"Accept":          "*/*",
		"MT-Network-Type": "WIFI",
		"MT-Token":        "",
		"MT-Team-ID":      "",
		"MT-Info":         "028e7f96f6369cafe1d105579c5b9377",
		"MT-Device-ID":    device,
		"MT-Bundle-ID":    "com.moutai.mall",
		"Accept-Language": "en-CN;q=1, zh-Hans-CN;q=0.9",
		"MT-Request-ID":   fmt.Sprintf("%d", time.Now().UnixMicro()*100),
		"MT-APP-Version":  version(),
		"User-Agent":      "iOS;16.3;Apple;?unrecognized?",
		"MT-R":            "clips_OlU6TmFRag5rCXwbNAQ/Tz1SKlN8THcecBp/HGhHdw==",
		"Content-Length":  "93",
		"Accept-Encoding": "gzip, deflate, br",
		"Connection":      "keep-alive",
		"Content-Type":    "application/json",
		"userId":          "",
	}
}

var reg = regexp.MustCompile(`版本 (\d+\.\d+\.\d+)`)

func signature(m map[string]string, t int64) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	text := SALT
	for _, key := range keys {
		text += m[key]
	}
	text += fmt.Sprintf("%d", t)
	return MD5(text)
}

// MD5 md5
func MD5(data string) string {
	hash := md5.New()
	hash.Write([]byte(data))

	return hex.EncodeToString(hash.Sum(nil))
}

func version() string {
	resp, _ := http.Get("https://apps.apple.com/cn/app/i%E8%8C%85%E5%8F%B0/id1600482450")
	defer resp.Body.Close()
	all, _ := io.ReadAll(resp.Body)
	submatch := reg.FindSubmatch(all)
	if len(submatch) == 2 {
		return string(submatch[1])
	}
	return ""
}

type loginResp struct {
	Code string `json:"code"`
	Data struct {
		UserID   int    `json:"userId"`
		UserName string `json:"userName"`
		Mobile   string `json:"mobile"`
		IDCode   string `json:"idCode"`
		Token    string `json:"token"`
	} `json:"data"`
}

func login(mobile string, code string) (uid int, token string) {
	currentTimestamp := time.Now().UnixMilli()
	var params = map[string]string{
		"mobile":  mobile,
		"vCode":   code,
		"ydToken": "",
		"ydLogId": "",
	}
	md5 := signature(params, currentTimestamp)
	var body = fmt.Sprintf(`{"vCode": %s, "ydToken": "", "ydLogId": "", "mobile": "%s", "md5": "%s", "timestamp": "%d", "MT-APP-Version": "%s"}`, code, mobile, md5, currentTimestamp, version())
	request, _ := http.NewRequest("POST", "https://app.moutai519.com.cn/xhr/front/user/register/login", strings.NewReader(body))
	addHeaders(request)
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	var data loginResp
	json.NewDecoder(do.Body).Decode(&data)
	return data.Data.UserID, data.Data.Token
}

func getCode(mobile string) error {
	currentTimestamp := time.Now().UnixMilli()
	body := fmt.Sprintf(`{"mobile": "%s", "md5": "%s", "timestamp": "%d", "MT-APP-Version": "%s"}`, mobile, signature(map[string]string{"mobile": mobile}, currentTimestamp), currentTimestamp, version())
	request, _ := http.NewRequest("POST", "https://app.moutai519.com.cn/xhr/front/user/register/vcode", strings.NewReader(body))
	addHeaders(request)
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	type resp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	var res resp
	json.NewDecoder(do.Body).Decode(&res)
	if res.Code != "2000" {
		return errors.New(res.Message)
	}
	return nil
}

func addHeaders(request *http.Request) map[string]string {
	h := headers()
	for k, v := range h {
		request.Header.Add(k, v)
	}
	return h
}

type sessionResp struct {
	Code int `json:"code"`
	Data struct {
		SessionID int `json:"sessionId"`
	} `json:"data"`
}

func today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
}

func GetCurrentSessionID() int {
	resp, _ := http.Get(fmt.Sprintf("https://static.moutai519.com.cn/mt-backend/xhr/front/mall/index/session/get/%d", today().UnixMilli()))
	defer resp.Body.Close()
	var data sessionResp
	json.NewDecoder(resp.Body).Decode(&data)
	return data.Data.SessionID
}

type list struct {
	Count  int `json:"count"`
	ItemID int `json:"itemId"`
}

type ActParams struct {
	ActParam     string `json:"actParam,omitempty"`
	ItemInfoList []list `json:"itemInfoList"`
	ShopID       string `json:"shopId"`
	SessionID    int    `json:"sessionId"`
}

func reservation(itemID int, shop shopInfo, sessionID, userID int, token string) string {
	p := &ActParams{
		ItemInfoList: []list{
			{
				Count:  1,
				ItemID: itemID,
			},
		},
		ShopID:    shop.ID,
		SessionID: sessionID,
	}
	marshal, _ := json.Marshal(p)
	p.ActParam = encrypt(marshal)
	b, _ := json.Marshal(p)
	request, _ := http.NewRequest("POST", "https://app.moutai519.com.cn/xhr/front/mall/reservation/add", bytes.NewReader(b))
	addHeaders(request)
	request.Header.Del("userId")
	request.Header.Del("MT-Token")
	request.Header.Add("userId", fmt.Sprintf("%d", userID))
	request.Header.Add("MT-Token", token)
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	type resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	var data resp
	json.NewDecoder(do.Body).Decode(&data)
	if data.Code == 2000 {
		return fmt.Sprintf("申购成功：%d, 店铺：%s", itemID, shop.Name)
	}
	return fmt.Sprintf("itemID: %d, 店铺: %s, %s", itemID, shop.Name, data.Message)
}

func encrypt[T string | []byte](text T) string {
	dst, _ := openssl.AesCBCEncrypt([]byte(text), AES_KEY, AES_IV, openssl.PKCS7_PADDING)
	return base64.StdEncoding.EncodeToString(dst)
}

func decrypt[T string | []byte](text T) string {
	dst, _ := openssl.AesCBCDecrypt([]byte(text), AES_KEY, AES_IV, openssl.PKCS7_PADDING)
	return string(dst)
}

type resourceMap struct {
	Data struct {
		MtshopsPc struct {
			Md5 string `json:"md5"`
			URL string `json:"url"`
		} `json:"mtshops_pc"`
	} `json:"data"`
}

func getMap() AllShopMap {
	var data resourceMap

	err := retry.Times(3, func() error {
		resp, err := http.Get("https://static.moutai519.com.cn/mt-backend/xhr/front/mall/resource/get")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return json.NewDecoder(resp.Body).Decode(&data)
	})
	if err != nil {
		log.Println(err)
		return nil
	}
	var shops AllShopMap
	err = retry.Times(3, func() error {
		get, err := http.Get(data.Data.MtshopsPc.URL)
		if err != nil {
			log.Println(err)
			return err
		}
		defer get.Body.Close()
		return json.NewDecoder(get.Body).Decode(&shops)
	})
	if err != nil {
		log.Println(err)
		return nil
	}
	return shops
}

type ShopAddr struct {
	Name       string  `json:"name"`
	TenantName string  `json:"tenant_name"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
}

type AllShopMap map[string]ShopAddr

type LatLng struct {
	lat float64
	lng float64
}

func GetDistance(d1, d2 LatLng) float64 {
	radius := 6371000.0
	rad := math.Pi / 180.0
	lat1 := d1.lat * rad
	lng1 := d1.lng * rad
	lat2 := d2.lat * rad
	lng2 := d2.lng * rad
	theta := lng2 - lng1
	dist := math.Acos(math.Sin(lat1)*math.Sin(lat2) + math.Cos(lat1)*math.Cos(lat2)*math.Cos(theta))
	return dist * radius / 1000
}
