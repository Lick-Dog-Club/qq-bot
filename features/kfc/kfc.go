package kfc

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"qq/bot"
	"qq/features"
)

func init() {
	features.AddKeyword("kfc", "KFC 骚话", func(bot bot.Bot, content string) error {
		bot.Send(Get())
		return nil
	})
}

func Get() string {
	resp, _ := http.DefaultClient.Get("https://raw.fastgit.org/Nthily/KFC-Crazy-Thursday/main/kfc.json")
	defer resp.Body.Close()

	var data []response
	json.NewDecoder(resp.Body).Decode(&data)
	return data[rand.Intn(len(data))].Text
}

type response struct {
	Text string `json:"text"`
}
