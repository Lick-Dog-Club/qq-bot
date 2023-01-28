package lottery

type UserInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Mid      int    `json:"mid"`
		Uname    string `json:"uname"`
		Userid   string `json:"userid"`
		Sign     string `json:"sign"`
		Birthday string `json:"birthday"`
		Sex      string `json:"sex"`
		NickFree bool   `json:"nick_free"`
		Rank     string `json:"rank"`
	} `json:"data"`
}

type FeedAll struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		HasMore bool `json:"has_more"`
		Items   []struct {
			IDStr   string `json:"id_str"`
			Modules struct {
				ModuleAuthor struct {
					Mid             int    `json:"mid"`
					Name            string `json:"name"`
					PubAction       string `json:"pub_action"`
					PubLocationText string `json:"pub_location_text"`
					PubTime         string `json:"pub_time"`
					PubTs           int    `json:"pub_ts"`
					Type            string `json:"type"`
				} `json:"module_author"`
				ModuleDynamic struct {
					Additional interface{} `json:"additional"`
					Desc       struct {
						RichTextNodes []struct {
							OrigText string `json:"orig_text"`
							Text     string `json:"text"`
							Type     string `json:"type"`
						} `json:"rich_text_nodes"`
						Text string `json:"text"`
					} `json:"desc"`
					Major struct {
						Archive struct {
							Aid   string `json:"aid"`
							Badge struct {
								BgColor string `json:"bg_color"`
								Color   string `json:"color"`
								Text    string `json:"text"`
							} `json:"badge"`
							Bvid           string `json:"bvid"`
							Cover          string `json:"cover"`
							Desc           string `json:"desc"`
							DisablePreview int    `json:"disable_preview"`
							DurationText   string `json:"duration_text"`
							JumpURL        string `json:"jump_url"`
							Stat           struct {
								Danmaku string `json:"danmaku"`
								Play    string `json:"play"`
							} `json:"stat"`
							Title string `json:"title"`
							Type  int    `json:"type"`
						} `json:"archive"`
						Type string `json:"type"`
					} `json:"major"`
					Topic interface{} `json:"topic"`
				} `json:"module_dynamic"`
			} `json:"modules"`
			Type    string `json:"type"`
			Visible bool   `json:"visible"`
		} `json:"items"`
		Offset         string `json:"offset"`
		UpdateBaseline string `json:"update_baseline"`
		UpdateNum      int    `json:"update_num"`
	} `json:"data"`
}

type NoticeResp struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
	Data    struct {
		// lottery_time 开奖时间
		LotteryTime    int    `json:"lottery_time"`
		FirstPrize     int    `json:"first_prize"`
		SecondPrize    int    `json:"second_prize"`
		ThirdPrize     int    `json:"third_prize"`
		FirstPrizeCmt  string `json:"first_prize_cmt"`
		SecondPrizeCmt string `json:"second_prize_cmt"`
		ThirdPrizeCmt  string `json:"third_prize_cmt"`
		FirstPrizePic  string `json:"first_prize_pic"`
		SecondPrizePic string `json:"second_prize_pic"`
		ThirdPrizePic  string `json:"third_prize_pic"`
		NeedPost       int    `json:"need_post"`
		BusinessType   int    `json:"business_type"`
		SenderUID      int    `json:"sender_uid"`
		PayStatus      int    `json:"pay_status"`
		Ts             int    `json:"ts"`
		LotteryID      int    `json:"lottery_id"`
		Gt             int    `json:"_gt_"`
	} `json:"data"`
}

type MyForwardResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		HasMore bool `json:"has_more"`
		Items   []struct {
			IDStr string `json:"id_str"`
			Orig  struct {
				Basic struct {
					CommentIDStr string `json:"comment_id_str"`
					CommentType  int    `json:"comment_type"`
					LikeIcon     struct {
						ActionURL string `json:"action_url"`
						EndURL    string `json:"end_url"`
						ID        int    `json:"id"`
						StartURL  string `json:"start_url"`
					} `json:"like_icon"`
					RidStr string `json:"rid_str"`
				} `json:"basic"`
				IDStr   string `json:"id_str"`
				Modules struct {
					ModuleAuthor struct {
						Name string `json:"name"`
					} `json:"module_author"`
				} `json:"modules"`
				Type    string `json:"type"`
				Visible bool   `json:"visible"`
			} `json:"orig,omitempty"`
			Type string `json:"type"`
		} `json:"items"`
		Offset string `json:"offset"`
	} `json:"data"`
}
