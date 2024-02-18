package lottery

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var httpClient = &http.Client{}

type User struct {
	cookie   cookiePairs
	forwards map[string]struct{}
	me       userInfo
}

func (u *User) buildRequest(url string) *http.Request {
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("cookie", fmt.Sprintf("DedeUserID=%v; SESSDATA=%v; bili_jct=%v; DedeUserID__ckMd5=%v",
		u.cookie["DedeUserID"],
		u.cookie["SESSDATA"],
		u.cookie["bili_jct"],
		u.cookie["DedeUserID__ckMd5"]),
	)
	request.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36")
	request.Header.Add("referer", "https://www.bilibili.com/")
	return request
}

// info https://api.bilibili.com/x/member/web/account
func (u *User) info() (userInfo, error) {
	response, err := httpClient.Do(u.buildRequest("https://api.bilibili.com/x/member/web/account"))
	if err != nil {
		return userInfo{}, err
	}
	defer closeBody(response.Body)
	var info userInfo
	err = json.NewDecoder(response.Body).Decode(&info)
	if err != nil {
		return userInfo{}, err
	}
	if info.Code == 0 {
		log.Println(info.Data.Uname, info.Data.Mid)
	} else {
		return userInfo{}, errors.New(info.Message)
	}
	u.me = info
	return info, nil
}

// myForwards https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset=&host_mid=345516933&timezone_offset=-480
func (u *User) myForwards(mid int) map[string]struct{} {
	var res = make(map[string]struct{})
	curlFn := func(offset string, mid int) (string, bool) {
		resp, _ := httpClient.Do(u.buildRequest(fmt.Sprintf("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset=%v&host_mid=%v&timezone_offset=-480", offset, mid)))
		defer closeBody(resp.Body)
		var data myForwardResp
		json.NewDecoder(resp.Body).Decode(&data)
		for _, item := range data.Data.Items {
			if item.Type == "DYNAMIC_TYPE_FORWARD" {
				res[item.Orig.IDStr] = struct{}{}
			}
		}
		return data.Data.Offset, data.Data.HasMore
	}
	var (
		hasMore bool = true
		offset  string
	)

	for hasMore {
		offset, hasMore = curlFn(offset, mid)
	}
	return res
}

func (u *User) lotteryDynamics() (res []noticeBody) {
	curlFn := func(offset string, page int) (bool, string) {
		url := fmt.Sprintf("https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/all?timezone_offset=-480&type=all&pageSize=50&page=%v", page)
		if offset != "" {
			url += "&offset=" + offset
		}
		request := u.buildRequest(url)
		resp, _ := httpClient.Do(request)
		defer closeBody(resp.Body)
		var data feedAll
		json.NewDecoder(resp.Body).Decode(&data)
	Loop:
		for _, item := range data.Data.Items {
			if len(item.Modules.ModuleDynamic.Desc.RichTextNodes) > 0 {
				for _, node := range item.Modules.ModuleDynamic.Desc.RichTextNodes {
					if node.Type == "RICH_TEXT_NODE_TYPE_LOTTERY" {
						res = append(res, u.lotteryNotice(item.Modules.ModuleAuthor.Name, item.IDStr))
						continue Loop
					}
				}
			}
		}
		return data.Data.HasMore, data.Data.Offset
	}

	var hasMore bool = true
	var page int = 1
	var offset string
	for hasMore {
		hasMore, offset = curlFn(offset, page)
		page++
		time.Sleep(300 * time.Millisecond)
	}
	log.Println(page)
	return res
}

// lotteryNotice https://api.vc.bilibili.com/lottery_svr/v1/lottery_svr/lottery_notice?dynamic_id=753851500355125303
func (u *User) lotteryNotice(up, dynamicId string) noticeBody {
	resp, _ := httpClient.Do(u.buildRequest("https://api.vc.bilibili.com/lottery_svr/v1/lottery_svr/lottery_notice?dynamic_id=" + dynamicId))
	defer closeBody(resp.Body)
	var data noticeResp
	json.NewDecoder(resp.Body).Decode(&data)
	_, ok := u.forwards[dynamicId]
	deadline := time.Unix(int64(data.Data.LotteryTime), 0)
	atoi, _ := strconv.Atoi(dynamicId)

	in := noticeBody{
		DynamicId:      atoi,
		Up:             up,
		WebUrl:         fmt.Sprintf("https://t.bilibili.com/%v", dynamicId),
		Deadline:       deadline.Format("2006-01-02 15:04:05"),
		Past:           deadline.Before(time.Now()),
		FirstPrizeCmt:  data.Data.FirstPrizeCmt,
		FirstPrize:     data.Data.FirstPrize,
		SecondPrizeCmt: data.Data.SecondPrizeCmt,
		SecondPrize:    data.Data.SecondPrize,
		ThirdPrizeCmt:  data.Data.ThirdPrizeCmt,
		ThirdPrize:     data.Data.ThirdPrize,
		Forwarded:      ok,
	}
	bf := &bytes.Buffer{}
	lotteryNoticeTemplate.Execute(bf, in)
	log.Println(bf.String())

	return in
}

type noticeBodyList []noticeBody

func (l noticeBodyList) String() string {
	bf := &bytes.Buffer{}
	for _, body := range l {
		lotteryNoticeTemplate.Execute(bf, body)
	}
	s := bf.String()
	if s == "" {
		s = "当前暂无未转发的抽奖"
	}
	return s
}

type noticeBody struct {
	Up             string
	DynamicId      int
	Forwarded      bool
	Past           bool
	WebUrl         string
	Deadline       string
	FirstPrizeCmt  string
	FirstPrize     int
	SecondPrizeCmt string
	SecondPrize    int
	ThirdPrizeCmt  string
	ThirdPrize     int
}

var lotteryNoticeTemplate, _ = template.New("").Parse(`
up: {{ .Up }}
网页链接: {{ .WebUrl }}
抽奖截止时间: {{ .Deadline }} ({{ if .Past }}已开奖{{else}}未开奖{{end}})
一等奖: {{ .FirstPrizeCmt }}, {{ .FirstPrize }} 名
{{- if gt .SecondPrize 0 }}
二等奖: {{ .SecondPrizeCmt }}, {{ .SecondPrize }} 名
{{- end }}
{{- if gt .ThirdPrize 0 }}
三等奖: {{ .ThirdPrizeCmt }}, {{ .ThirdPrize }} 名
{{- end }}
是否转发: {{ if .Forwarded }}是{{ else }}否{{ end }}
`)

// dynaRepost 转发动态
//
// dyid 为转发的动态ID
func (u *User) dynaRepost(dyid int64, content string) error {
	req := u.buildRequest("https://api.vc.bilibili.com/dynamic_repost/v1/dynamic_repost/repost")
	req.Method = "POST"
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")
	v := url.Values{}
	v.Add("csrf", u.cookie["bili_jct"])
	v.Add("dynamic_id", strconv.FormatInt(dyid, 10))
	v.Add("content", content)
	v.Add("at_uids", "")
	v.Add("ctrl", "")
	req.Body = io.NopCloser(strings.NewReader(v.Encode()))

	do, _ := httpClient.Do(req)
	defer closeBody(do.Body)
	all, _ := io.ReadAll(do.Body)
	log.Println(string(all))
	return nil
}

func closeBody(rc io.ReadCloser) {
	io.Copy(io.Discard, rc)
	rc.Close()
}

// 判断 cookie 是否正确，输出当前用户名称: 校验 cookie 过期，提醒用户更新 cookie
// 关注转发
// ---- 1. 获取我关注的近两周的进行中的抽奖动态，转发没转发过的: 幂等
// ---- 2. 获取所有正在进行中的抽奖，关注 + 转发，关注后不想看他们的动态，还要研究下如何只关注，不接受动态
// 转发后返回转发详情
