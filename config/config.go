package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"os"
	"path/filepath"
	"qq/util"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/3bl3gamer/tgclient/mtproto"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

var c atomic.Value

var (
	DataDir = func() string {
		dir := "/data"
		en := os.Getenv("DATA_DIR")
		if en != "" {
			dir = en
		}
		return dir
	}()
	adminID          = os.Getenv("ADMIN_USER_ID")
	ForceStoreConfig = os.Getenv("FORCE_STORE_CONFIG") == "1"
)

var (
	ImageDir   string = filepath.Join(DataDir, "images")
	ConfigFile string = filepath.Join(DataDir, "qq-bot.json")
)

func init() {
	c.Store(mappingKV)
	if Pod() != "" || ForceStoreConfig {
		file, err := os.ReadFile(ConfigFile)
		if err == nil {
			v := KV{}
			json.Unmarshal(file, &v)
			Set(v)
		}
	}
}

func Configs() KV {
	return c.Load().(KV)
}

type KV map[string]string

func (k KV) Marshal() []byte {
	var kv = KV{}
	for key, v := range k {
		if v != "" {
			kv[key] = v
		}
	}
	marshal, _ := json.Marshal(kv)
	return marshal
}

func (k KV) String() string {
	var s string
	for key, value := range k {
		str := ""
		if len(value) <= 100 {
			str = value
		}
		for len(value) > 100 {
			str += value[0:100] + "\n"
			value = value[100:]
		}
		s += fmt.Sprintf("%s=%s\n", key, str)
	}
	return s
}

func PixivSession() string {
	return c.Load().(KV)["pixiv_session"]
}

func AiToken() string {
	return c.Load().(KV)["ai_token"]
}

func ChatGPTApiModel() string {
	return c.Load().(KV)["chatgpt_model"]
}

func PixivMode() string {
	return c.Load().(KV)["pixiv_mode"]
}

func GroupID() string {
	return c.Load().(KV)["group_id"]
}

type UID string

func (u UID) List() []string {
	return strings.Split(string(u), ",")
}

func (u UID) Contains(id string) bool {
	split := strings.Split(string(u), ",")
	for _, uid := range split {
		if uid == id {
			return true
		}
	}

	return false
}

func UserID() string {
	return c.Load().(KV)["user_id"]
}

func AdminIDs() UID {
	return UID(c.Load().(KV)["admin_id"])
}

func BiliCookie() string {
	return c.Load().(KV)["bili_cookie"]
}

func Namespace() string {
	return c.Load().(KV)["namespace"]
}

func TgInfo() (res *mtproto.SessionInfo) {
	if err := json.Unmarshal([]byte(c.Load().(KV)["tg_info"]), res); err != nil {
		log.Println(err)
	}
	return
}

func Pod() string {
	return c.Load().(KV)["pod_name"]
}

func WeatherKey() string {
	return c.Load().(KV)["weather_key"]
}

func GeoKey() string {
	return c.Load().(KV)["weather_key"]
}

func TianApiKey() string {
	return c.Load().(KV)["tian_api_key"]
}

func HttpProxy() string {
	return c.Load().(KV)["http_proxy"]
}
func GoogleSearchKey() string {
	return c.Load().(KV)["google_key"]
}
func GoogleSearchCX() string {
	return c.Load().(KV)["google_cx"]
}
func GPTOnlySearch() bool {
	return c.Load().(KV)["only_search"] == "1"
}

func BinanceKey() string {
	return c.Load().(KV)["binance_key"]
}

func BinanceSecret() string {
	return c.Load().(KV)["binance_secret"]
}

func BinanceDiff() string {
	return c.Load().(KV)["binance_diff"]
}

func TgPhone() string {
	return c.Load().(KV)["tg_phone"]
}

func TgCode() string {
	return c.Load().(KV)["tg_code"]
}

func TgAppHash() string {
	return c.Load().(KV)["tg_app_hash"]
}

func Birthday() string {
	return c.Load().(KV)["birthday"]
}
func OKey() string {
	return c.Load().(KV)["o_key"]
}

func OSecret() string {
	return c.Load().(KV)["o_secret"]
}

type Task struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	RunAt   string `json:"run_at"`
	Content string `json:"content"`
	UserID  string `json:"user_id"`
	GroupID string `json:"group_id"`
}

func Tasks() []Task {
	res := make([]Task, 0)
	json.Unmarshal([]byte(c.Load().(KV)["tasks"]), &res)
	return res
}

func SyncTasks(res []Task) {
	marshal, _ := json.Marshal(&res)
	Set(map[string]string{"tasks": string(marshal)})
	return
}

func WebotUsers() map[string]struct{} {
	m := make(map[string]struct{})
	for _, s := range strings.Split(c.Load().(KV)["webot_users"], ",") {
		m[s] = struct{}{}
	}
	return m
}

func TgAppID() int32 {
	atoi, _ := strconv.Atoi(c.Load().(KV)["tg_app_id"])
	return int32(atoi)
}

func BarkUrls() []string {
	return strings.Split(c.Load().(KV)["bark_token"], ",")
}
func TaobaoBarkUrls() []string {
	return strings.Split(c.Load().(KV)["taobao_bark_url"], ",")
}

func TaobaoSkus() Skus {
	data := Skus{}
	json.Unmarshal([]byte(c.Load().(KV)["taobao_skus"]), &data)
	return data
}

type Skus map[int64]map[int64]Sku

func (s Skus) Add(sku Sku) {
	var mm = make(map[int64]Sku)
	if m, ok := s[sku.NumIID]; ok {
		mm = m
	}
	mm[sku.SkuID] = sku
	s[sku.NumIID] = mm
}

func (s Skus) HasDiff(sku Sku) (bool, error) {
	if m, ok := s[sku.NumIID]; ok {
		if v, ok := m[sku.SkuID]; ok {
			return !(v.Price == sku.Price && v.OriginalPrice == sku.OriginalPrice), nil
		}
	}
	return false, errors.New("sku 不存在")
}

func (s Skus) String() string {
	parse, err := template.New("").Parse(`
{{- range $k,$v := .}}
numIID: {{ $k }}
{{- range $kk,$vv := $v }}
{{$vv.Title}} sku({{ $kk }}): 价格是 ¥{{ $vv.Price }} ，原价是 ¥{{ $vv.OriginalPrice }} {{$vv.Op}}
{{- end }}
{{"\n"}}
{{- end }}
`)
	if err != nil {
		panic(err)
	}
	bf := bytes.Buffer{}
	parse.Execute(&bf, s)
	return bf.String()
}

type Sku struct {
	NumIID        int64   `json:"num_iid"`
	Title         string  `json:"title"`
	SkuID         int64   `json:"sku_id"`
	Price         float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`

	Op string `json:"-"`
}

func TaobaoIDs() []string {
	return strings.Split(c.Load().(KV)["taobao_ids"], ",")
}
func DisabledCrons() Cmd {
	return Cmd(c.Load().(KV)["disabled_crons"])
}
func RunWebotOnSysStart() bool {
	return c.Load().(KV)["run_webot"] == "1"
}

func CronEnabled() bool {
	return c.Load().(KV)["cron_enabled"] == "1"
}
func GoldCount() string {
	return c.Load().(KV)["gold_count"]
}

func BgOneHandUSDT() float64 {
	return util.ToFloat64(c.Load().(KV)["bg_one_hand"])
}

func GoldShowLabel() bool {
	return c.Load().(KV)["gold_show_label"] == "1"
}

type BgBuy struct {
	Coin       string
	PriceBelow float64
}

func BgBuyCoin() []BgBuy {
	s := c.Load().(KV)["bg_buy"]
	split := strings.Split(s, ";")
	buys := make([]BgBuy, 0, len(split))
	for _, s2 := range split {
		i := strings.Split(s2, ",")
		if len(i) != 2 {
			continue
		}
		buys = append(buys, BgBuy{
			Coin:       i[0],
			PriceBelow: util.ToFloat64(i[1]),
		})
	}
	return buys
}

var kvHelp = map[string]string{
	"bg_coin_watch": "XRP8USDT_SPBL,-15;BTCUSDT_SPBL,-0.01 # 监控币种价格变化率，支持百分比 ';' 分隔",
	"bg_buy":        "XRP8USDT_SPBL,0.003; # 买入币种价格低于多少时买入, ';' 分隔",
	"birthday":      "2021-01-01 # 生日格式",
	"bark_token":    "xxa,aaa # bark 推送 token，',' 分隔",
	"x_tokens":      "token,csrf; # x 认证令牌，';' 分隔",
}

func GetHelp(key string) string {
	return kvHelp[key]
}

var mappingKV = KV{
	"bg_one_hand": "",
	"bg_buy":      "",
	// onebound
	"o_key":           "",
	"cron_enabled":    "1",
	"gold_show_label": "0",
	// onebound
	"o_secret":    "",
	"taobao_skus": "",
	"taobao_ids":  "",
	// https://api.day.app/xxxxxx/标题/内容
	"bark_url":        "",
	"bark_token":      "",
	"taobao_bark_url": "",
	"bili_cookie":     "",
	"disabled_crons":  "",
	"user_id":         "",
	"run_webot":       "0",
	// QQ 号码，"," 分隔，无法使用 config 设置
	"admin_id":       "",
	"ai_token":       "",
	"chatgpt_model":  openai.GPT3Dot5Turbo16K0613,
	"pixiv_mode":     "daily",
	"pixiv_session":  "",
	"tasks":          "",
	"webot_users":    "",
	"group_id":       os.Getenv("GROUP_ID"),
	"namespace":      os.Getenv("APP_NAMESPACE"),
	"pod_name":       os.Getenv("POD_NAME"),
	"weather_key":    os.Getenv("WEATHER_KEY"),
	"tian_api_key":   os.Getenv("TIAN_API_KEY"),
	"http_proxy":     os.Getenv("HTTP_PROXY"),
	"binance_key":    "",
	"google_key":     "",
	"google_cx":      "",
	"only_search":    "",
	"birthday":       "",
	"ai_max_token":   "128000",
	"binance_secret": "",
	"binance_diff":   "100",
	"maotai":         "",
	"tg_info":        "",
	"tg_app_id":      "",
	"bg_coin_watch":  "",
	"bg_goal":        "",
	"tg_app_hash":    "",
	"tg_phone":       "",
	"tg_code":        "",

	"bg_money_diff":     "30",
	"bg_api_key":        "",
	"bg_passphrase":     "",
	"bg_api_secret_key": "",
	"gold_count":        "40",

	"12306_JSESSIONID": "",

	"disabled_cmds": "",

	"x_tokens": "",
	// x 推文发送到哪个群里
	"x_group_id": "",
	"x_users":    "",

	// 有道翻译
	"yd_secret": "",
	"yd_key":    "",
}

type Token struct {
	Token string
	CSRF  string
}

func XGroupID() string {
	return c.Load().(KV)["x_group_id"]
}

func XUsers() []string {
	return strings.Split(c.Load().(KV)["x_users"], ",")
}

func XTokens() (res []Token) {
	split := strings.Split(c.Load().(KV)["x_tokens"], ";")
	for _, s := range split {
		i := strings.Split(s, ",")
		if len(i) == 2 {
			res = append(res, Token{
				Token: i[0],
				CSRF:  i[1],
			})
		}
	}
	rand.Shuffle(len(res), func(i, j int) {
		res[i], res[j] = res[j], res[i]
	})

	return
}

type Cmd string

func (c Cmd) Contains(keyword string) bool {
	split := strings.Split(string(c), ",")
	for _, cmd := range split {
		if cmd == keyword {
			return true
		}
	}

	return false
}

func DisabledCmds() Cmd {
	return Cmd(c.Load().(KV)["disabled_cmds"])
}

func T12306JSESSIONID() string {
	return c.Load().(KV)["12306_JSESSIONID"]
}

func BgApiKey() string {
	return c.Load().(KV)["bg_api_key"]
}

type WatchCoin struct {
	Name string
	// 价格变化率
	Rate  []float64
	Price float64
}

func BgGoal() []WatchCoin {
	s := c.Load().(KV)["bg_goal"]
	split := strings.Split(s, ";")
	watchCoins := make([]WatchCoin, 0, len(split))
	for _, s2 := range split {
		i := strings.Split(s2, ",")
		if len(i) == 2 {
			watchCoins = append(watchCoins, WatchCoin{
				Name:  i[0],
				Price: util.ToFloat64(i[1]),
			})
		}
	}
	return watchCoins
}

func YDKey() string {
	return c.Load().(KV)["yd_key"]
}
func YDSecret() string {
	return c.Load().(KV)["yd_secret"]
}

func BgCoinWatch() []WatchCoin {
	s := c.Load().(KV)["bg_coin_watch"]
	split := strings.Split(s, ";")
	watchCoins := make([]WatchCoin, 0, len(split))
	for _, s2 := range split {
		i := strings.Split(s2, ",")
		var rates []float64
		for _, s3 := range i[1:] {
			f := util.ToFloat64(s3)
			if f > 1 || f < -1 {
				f = f / 100
			}
			rates = append(rates, f)
		}
		watchCoins = append(watchCoins, WatchCoin{
			Name: i[0],
			Rate: rates,
		})
	}
	return watchCoins
}

func BgMoneyDiff() string {
	return c.Load().(KV)["bg_money_diff"]
}

func BgPassphrase() string {
	return c.Load().(KV)["bg_passphrase"]
}

func BgApiSecretKey() string {
	return c.Load().(KV)["bg_api_secret_key"]
}

func AIMaxToken() int64 {
	return util.ToInt64(c.Load().(KV)["ai_max_token"])
}

type MTInfos map[string]MaoTaiInfo

func AddMaoTaiInfo(info MaoTaiInfo) {
	var infos MTInfos
	if err := json.Unmarshal([]byte(c.Load().(KV)["maotai"]), &infos); err != nil {
		infos = MTInfos{}
	}
	infos[info.Phone] = info
	marshal, _ := json.Marshal(&infos)
	Set(map[string]string{"maotai": string(marshal)})
}

func DelMaoTaiInfo(phone string) {
	var infos MTInfos
	if err := json.Unmarshal([]byte(c.Load().(KV)["maotai"]), &infos); err != nil {
		infos = MTInfos{}
	}
	delete(infos, phone)
	marshal, _ := json.Marshal(&infos)
	Set(map[string]string{"maotai": string(marshal)})
}

type MaoTaiInfo struct {
	Phone    string    `json:"phone"`
	Uid      int       `json:"uid"`
	Token    string    `json:"token"`
	Cookie   string    `json:"cookie"`
	Lat      float64   `json:"lat"`
	Lng      float64   `json:"lng"`
	ExpireAt time.Time `json:"expire_at"`
}

func (m *MaoTaiInfo) Expired() bool {
	return m.ExpireAt.Before(time.Now())
}

func MaoTaiInfoMap() map[string]MaoTaiInfo {
	var infos MTInfos
	err := json.Unmarshal([]byte(c.Load().(KV)["maotai"]), &infos)
	if err != nil {
		infos = MTInfos{}
	}

	return infos
}

func Set(m map[string]string) (sets KV) {
	var newKv = KV{}
	sets = KV{}
	for k, v := range c.Load().(KV) {
		newv := v
		if s, ok := m[k]; ok && !(k == "pod_name" || k == "namespace" || k == "admin_id") {
			newv = s
			sets[k] = s
		}
		newKv[k] = newv
	}
	newKv["admin_id"] = adminID
	c.Store(newKv)
	if Pod() != "" || ForceStoreConfig {
		os.WriteFile(ConfigFile, newKv.Marshal(), 0644)
	}
	return sets
}
