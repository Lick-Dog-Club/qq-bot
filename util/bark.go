package util

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

func Bark(title, body string, barkUrls ...string) {
	for _, barkUrl := range barkUrls {
		bark(barkUrl, title, body)
	}
}

func bark(barkUrl, title, body string) {
	parse, err := url.Parse(barkUrl)
	if err != nil {
		log.Println(err)
		return
	}
	split := strings.Split(parse.Path, "/")
	if len(split) < 2 {
		log.Println("bark url error: ", split)
		return
	}
	resp, err := http.Get(fmt.Sprintf("https://api.day.app/%s/%s/%s", split[1], title, body))
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
}
