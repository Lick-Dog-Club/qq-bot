package kfc

import (
	"crypto/tls"
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

var c = http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func Get() string {
	resp, _ := c.Get("https://raw.fastgit.org/Nthily/KFC-Crazy-Thursday/main/kfc.json")
	defer resp.Body.Close()

	var data []response
	json.NewDecoder(resp.Body).Decode(&data)
	return data[rand.Intn(len(data))].Text
}

type response struct {
	Text string `json:"text"`
}
