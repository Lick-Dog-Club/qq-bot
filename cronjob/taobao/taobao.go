package taobao

import (
	"encoding/json"
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/taobaoprice"
	"qq/util"
)

func init() {
	cronjob.NewCommand("taobao", func(bot bot.CronBot) error {
		watch, err := Watch()
		if err != nil {
			bot.SendToUser(config.UserID(), err.Error())
			return nil
		}
		if len(watch) > 0 {
			var to config.Skus
			for _, m := range watch {
				for _, sku := range m {
					if sku.Op == OpUpdate {
						to.Add(sku)
					}
				}
			}
			if len(to) > 0 {
				util.Bark("价格发生变化", watch.String(), append(config.TaobaoBarkUrls(), config.BarkUrls()...)...)
			}
			bot.SendToUser(config.UserID(), watch.String())
		}
		return nil
	}).DailyAt("0-3,8-23").HourlyAt([]int{15})
}

const (
	OpAdd    = "有新的sku添加"
	OpUpdate = "价格发生更新"
)

func Watch() (config.Skus, error) {
	var changes = config.Skus{}
	prev := config.TaobaoSkus()
	var newMap = config.Skus{}
	for _, id := range config.TaobaoIDs() {
		search, err := taobaoprice.Search(id)
		if err != nil {
			return nil, err
		}
		fmt.Println(search.String())

		for numiid, m := range search {
			newMap[numiid] = m
			for _, sku := range m {
				diff, err := prev.HasDiff(sku)
				if err != nil {
					sku.Op = OpAdd
					changes.Add(sku)
					continue
				}
				if diff {
					sku.Op = OpUpdate
					changes.Add(sku)
				}
			}
		}
	}
	marshal, _ := json.Marshal(newMap)
	config.Set(map[string]string{"taobao_skus": string(marshal)})
	return changes, nil
}
