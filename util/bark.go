package util

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

func Bark(title, body string, barkUrls ...string) {
	for _, barkUrl := range barkUrls {
		bark(barkUrl, title, body)
	}
}

func bark(barkUrl, title, body string) {
	resp, err := http.Get(fmt.Sprintf("https://api.day.app/%s/%s/%s", barkUrl, url.QueryEscape(title), url.QueryEscape(body)))
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
}
