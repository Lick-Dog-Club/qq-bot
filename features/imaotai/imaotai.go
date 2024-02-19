package imaotai

import (
	"bytes"
	"encoding/base64"
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
	"qq/util"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai/jsonschema"

	"github.com/forgoer/openssl"
)

var once sync.Once
var shops AllShopMap
var allShops func() AllShopMap = func() AllShopMap {
	once.Do(func() {
		shops = getMap()
	})
	return shops
}

func init() {
	features.AddKeyword("mt", "<+phoneNum>: 通过手机号自动预约茅台", func(bot bot.Bot, content string) error {
		bot.Send(Run(content))
		return nil
	}, features.WithGroup("maotai"), features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"phone": {
				Type:        jsonschema.String,
				Description: "13位的手机号, 例如 18888888888",
			},
		},
		Call: func(args string) (string, error) {
			var s = struct {
				Phone string `json:"phone"`
			}{}
			json.Unmarshal([]byte(args), &s)
			return Run(s.Phone), nil
		},
	}))
	features.AddKeyword("mt-redo", "全部重新申购", func(bot bot.Bot, content string) error {
		bot.SendTextImage(ReservationAll())
		return nil
	}, features.WithGroup("maotai"), features.WithHidden())
	features.AddKeyword("mt-reward", "领取小游戏奖励", func(bot bot.Bot, content string) error {
		bot.SendTextImage(ReceiveAllReward())
		return nil
	}, features.WithGroup("maotai"))
	features.AddKeyword("mt-game-up", "加速小游戏", func(bot bot.Bot, content string) error {
		bot.SendTextImage(SpeedUpGames())
		return nil
	}, features.WithGroup("maotai"))
	features.AddKeyword("mt-del", "<+phoneNum>: 取消茅台自动预约", func(bot bot.Bot, content string) error {
		config.DelMaoTaiInfo(content)
		bot.Send("成功取消！")
		return nil
	}, features.WithGroup("maotai"))
	features.AddKeyword("mt-jwd", "<+phone> <+lat,lng> 设置经纬度", func(bot bot.Bot, content string) error {
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
	features.AddKeyword("mt-geo", "<+phone> <+地址,高德自动查询 geo> 设置经纬度", func(bot bot.Bot, content string) error {
		split := strings.Split(content, " ")
		if len(split) == 2 {
			phone := split[0]
			geoStr := geo.Geo(config.GeoKey(), split[1])
			latlng := strings.Split(geoStr, ",")
			if len(latlng) == 2 {
				if info, ok := config.MaoTaiInfoMap()[phone]; ok {
					info.Lat = util.ToFloat64(latlng[1])
					info.Lng = util.ToFloat64(latlng[0])
					config.AddMaoTaiInfo(info)
					bot.Send(fmt.Sprintf(`设置成功：
手机号：%s
地址：%s
lat: %v
lng: %v
`, phone, split[1], info.Lat, info.Lng))
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

`, util.FuzzyPhone(info.Phone), info.ExpireAt.Format(time.DateTime), info.Lat, info.Lng)
		}
		bot.SendTextImage(res)
		return nil
	}, features.WithGroup("maotai"))
	features.AddKeyword("mt-login", "<+phone> <+code>: 茅台登录，通过6位短信验证码登录用户", func(bot bot.Bot, content string) error {
		split := strings.Split(content, " ")
		var phone, code string
		if len(split) >= 2 {
			phone = strings.TrimSpace(split[0])
			code = strings.TrimSpace(split[1])
		}
		bot.Send(loginAndStore(phone, code))
		return nil
	}, features.WithGroup("maotai"), features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"phone": {
				Type:        jsonschema.String,
				Description: "13位的手机号, 例如 18888888888",
			},
			"code": {
				Type:        jsonschema.String,
				Description: "短信验证码, 例如 123456",
			},
		},
		Call: func(args string) (string, error) {
			var s = struct {
				Phone string `json:"phone"`
				Code  string `json:"code"`
			}{}
			json.Unmarshal([]byte(args), &s)
			return loginAndStore(s.Phone, s.Code), nil
		},
	}))
}

func loginAndStore(phone, code string) string {
	uid, token, cookie := login(phone, code)
	info := config.MaoTaiInfo{
		Phone:    phone,
		Uid:      uid,
		Token:    token,
		Cookie:   cookie,
		ExpireAt: time.Time{},
	}
	var geoSet bool
	if taiInfo, ok := config.MaoTaiInfoMap()[phone]; ok {
		info.Lat = taiInfo.Lat
		info.Lng = taiInfo.Lng
		geoSet = true
	}

	if token != "" {
		decodeString, _ := base64.StdEncoding.DecodeString(strings.Split(token, ".")[1])
		var e exp
		json.Unmarshal([]byte(string(decodeString)+"}"), &e)
		info.ExpireAt = time.Unix(e.Exp, 0)
	}
	if info.ExpireAt.IsZero() {
		return "信息有误，添加失败"
	}
	config.AddMaoTaiInfo(info)
	return fmt.Sprintf(`
用户添加成功
过期时间是: %s
设置 geo 信息请执行(当前用户是否已设置 Geo 信息：%t):

mt-geo %s <地址>

申购茅台请执行:

mt %s
`, info.ExpireAt.Format(time.DateTime), geoSet, info.Phone, info.Phone)
}

func ReservationAll() string {
	var res string
	for _, info := range config.MaoTaiInfoMap() {
		if info.Expired() {
			res += fmt.Sprintf("%s: token已过期，需要重新登陆\n\n", util.FuzzyPhone(info.Phone))
			continue
		}
		res += fmt.Sprintf("%s:\n%s\n", util.FuzzyPhone(info.Phone), Run(info.Phone))
	}
	return res
}

var gameReward = []gameFunc{
	receiveTravel,
	receiveReWardMw,
	getEnergyAward,
}

// ReceiveAllReward 领取小游戏奖励
func ReceiveAllReward() string {
	var res string
	for _, info := range config.MaoTaiInfoMap() {
		if info.Cookie != "" {
			var str = []string{fmt.Sprintf("手机: %s", util.FuzzyPhone(info.Phone))}
			for _, fn := range gameReward {
				if s, err := fn(info.Cookie); err == nil {
					str = append(str, s)
				}
			}

			res += strings.Join(str, "\n") + "\n\n"
		}
	}
	return res
}

var gameQuickUp = []gameFunc{
	quickMw,
	quickTravel,
}

// SpeedUpGames 加速小游戏奖励
func SpeedUpGames() string {
	var res string
	for _, info := range config.MaoTaiInfoMap() {
		if info.Cookie != "" {
			var str = []string{fmt.Sprintf("手机: %s", util.FuzzyPhone(info.Phone))}
			for _, fn := range gameQuickUp {
				if s, err := fn(info.Cookie); err == nil {
					str = append(str, s)
				}
			}

			res += strings.Join(str, "\n") + "\n\n"
		}
	}
	return res
}

type exp struct {
	Exp int64 `json:"exp"`
}

type gameFunc func(cookie string) (string, error)

var games = []gameFunc{
	getEnergyAward,
	goTravel,
	startMw,
}

func Run(m string) string {
	// 1. sessionID
	sessionID := GetCurrentSessionID()

	info, ok := config.MaoTaiInfoMap()[m]
	if (ok && info.Expired()) || !ok {
		if err := getCode(m); err != nil {
			return err.Error()
		}
		return fmt.Sprintf("用户未登陆，短信已发送，收到后执行：\n\nmt-login %s <code>", m)
	}

	// 申购
	res := doReservation(sessionID, info.Uid, info.Token, LatLng{
		lat: info.Lat,
		lng: info.Lng,
	})

	// 领取耐力值
	if info.Cookie != "" {
		var otherAwards []string
		for _, game := range games {
			if s, err := game(info.Cookie); err == nil {
				otherAwards = append(otherAwards, s)
			}
		}
		if len(otherAwards) > 0 {
			res += strings.Join(otherAwards, ", ") + "\n"
		}
	}

	return res
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
				if addr, ok := allShops()[shop.ShopID]; ok {
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

var ItemIDs = []int{10941, 10942}

// doReservation 申购
func doReservation(sessionID, uid int, token string, latLng LatLng) (res string) {
	fmt.Printf("申购：\nsessionID: %v\nuid: %v\ntoken: %v\nlatlng: %#v", sessionID, uid, token, latLng)
	items := map[int][]shopInfo{}
	for _, id := range ItemIDs {
		shop := getItemShop(fmt.Sprintf(`https://static.moutai519.com.cn/mt-backend/xhr/front/mall/shop/list/slim/v3/%d/浙江省/%d/%d`, sessionID, id, util.Today().UnixMilli()), id, latLng)
		items[id] = append(items[id], shop...)
	}
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
	return util.MD5(text)
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
		Cookie   string `json:"cookie"`
	} `json:"data"`
}

func login(mobile string, code string) (uid int, token, cookie string) {
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
	return data.Data.UserID, data.Data.Token, data.Data.Cookie
}

func getCode(mobile string) error {
	currentTimestamp := time.Now().UnixMilli()
	body := fmt.Sprintf(`{"mobile": "%s", "md5": "%s", "timestamp": "%d", "MT-APP-Version": "%s"}`, mobile, signature(map[string]string{"mobile": mobile}, currentTimestamp), currentTimestamp, version())
	request, _ := http.NewRequest("POST", "https://app.moutai519.com.cn/xhr/front/user/register/vcode", strings.NewReader(body))
	addHeaders(request)
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	type resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	var res resp
	json.NewDecoder(do.Body).Decode(&res)
	if res.Code != 2000 {
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

func GetCurrentSessionID() int {
	resp, _ := http.Get(fmt.Sprintf("https://static.moutai519.com.cn/mt-backend/xhr/front/mall/index/session/get/%d", util.Today().UnixMilli()))
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
	resp, _ := http.Get("https://static.moutai519.com.cn/mt-backend/xhr/front/mall/resource/get")
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&data)
	var shops AllShopMap
	get, _ := http.Get(data.Data.MtshopsPc.URL)
	defer get.Body.Close()
	json.NewDecoder(get.Body).Decode(&shops)
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

type energyAwardResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AwardRule []struct {
			GoodID   int    `json:"goodId"`
			GoodName string `json:"goodName"`
			Count    int    `json:"count"`
		} `json:"awardRule"`
	} `json:"data"`
}

// getEnergyAward 领取耐力值
func getEnergyAward(cookie string) (string, error) {
	request, _ := http.NewRequest("POST", "https://h5.moutai519.com.cn/game/isolationPage/getUserEnergyAward", nil)
	addHeaders(request)
	request.Header.Add("cookie", fmt.Sprintf("MT-Token-Wap=%s;MT-Device-ID-Wap=%s;", cookie, device))
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	var data energyAwardResp
	json.NewDecoder(do.Body).Decode(&data)
	if data.Code == 200 && len(data.Data.AwardRule) > 0 {
		var str []string
		for _, s := range data.Data.AwardRule {
			str = append(str, fmt.Sprintf("已领取 %d %s", s.Count, s.GoodName))
		}
		return strings.Join(str, ", "), nil
	}

	return "耐力值: " + data.Message, nil
}

// goTravel 开始旅行
func goTravel(cookie string) (string, error) {
	request, _ := http.NewRequest("POST", "https://h5.moutai519.com.cn/game/xmTravel/startTravel", nil)
	addHeaders(request)
	request.Header.Add("cookie", fmt.Sprintf("MT-Token-Wap=%s;MT-Device-ID-Wap=%s;", cookie, device))
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	var data = struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}
	json.NewDecoder(do.Body).Decode(&data)
	if data.Code == 2000 {
		return "旅行成功", nil
	}
	return "旅行: " + data.Message, nil
}

// quickTravel 加速旅行
func quickTravel(cookie string) (string, error) {
	request, _ := http.NewRequest("POST", "https://h5.moutai519.com.cn/game/xmTravel/quickenTravel", nil)
	addHeaders(request)
	request.Header.Add("cookie", fmt.Sprintf("MT-Token-Wap=%s;MT-Device-ID-Wap=%s;", cookie, device))
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	var data = struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}
	json.NewDecoder(do.Body).Decode(&data)
	if data.Code == 2000 {
		return "旅行加速成功", nil
	}
	return "旅行加速: " + data.Message, nil
}

// {"code":2000,"message":null,"data":"领取成功","error":null}
// receiveTravel 领取旅行奖励
func receiveTravel(cookie string) (string, error) {
	var title = "旅行: "
	request, _ := http.NewRequest("POST", "https://h5.moutai519.com.cn/game/xmTravel/receiveReward", strings.NewReader(`{}`))
	addHeaders(request)
	request.Header.Add("cookie", fmt.Sprintf("MT-Token-Wap=%s;MT-Device-ID-Wap=%s; ", cookie, device))
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	var data = struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}{}
	json.NewDecoder(do.Body).Decode(&data)
	if data.Code == 2000 {
		return title + data.Data, nil
	}
	return title + data.Message, nil
}

// startMw 开始酿酒
func startMw(cookie string) (string, error) {
	request, _ := http.NewRequest("POST", "https://h5.moutai519.com.cn/game/xmMw/startMw", nil)
	addHeaders(request)
	request.Header.Add("cookie", fmt.Sprintf("MT-Token-Wap=%s;MT-Device-ID-Wap=%s;", cookie, device))
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	var data = struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}
	json.NewDecoder(do.Body).Decode(&data)
	if data.Code == 2000 {
		return "酿酒进行中", nil
	}
	return "酿酒: " + data.Message, nil
}

// quickMw 加速酿酒
func quickMw(cookie string) (string, error) {
	request, _ := http.NewRequest("POST", "https://h5.moutai519.com.cn/game/xmMw/quickenMw", nil)
	addHeaders(request)
	request.Header.Add("cookie", fmt.Sprintf("MT-Token-Wap=%s;MT-Device-ID-Wap=%s;", cookie, device))
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	var data = struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}
	json.NewDecoder(do.Body).Decode(&data)
	if data.Code == 2000 {
		return "酿酒加速成功", nil
	}
	return "酿酒加速: " + data.Message, nil
}

// receiveReWardMw 领取酿酒奖励
func receiveReWardMw(cookie string) (string, error) {
	var title = "酿酒: "
	request, _ := http.NewRequest("POST", "https://h5.moutai519.com.cn/game/xmMw/receiveReward", strings.NewReader(`{}`))
	addHeaders(request)
	request.Header.Add("cookie", fmt.Sprintf("MT-Token-Wap=%s;MT-Device-ID-Wap=%s;", cookie, device))
	do, _ := http.DefaultClient.Do(request)
	defer do.Body.Close()
	var data = struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}{}
	json.NewDecoder(do.Body).Decode(&data)
	if data.Code == 2000 {
		return title + data.Data, nil
	}
	return title + data.Message, nil
}
