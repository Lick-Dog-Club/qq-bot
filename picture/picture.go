package picture

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
)

var (
	urls []string = []string{
		"https://api.btstu.cn/sjbz/?lx=dongman",
		"https://www.dmoe.cc/random.php",
	}

	client = http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Url() string {
	var (
		response *http.Response
		err      error
	)
	if err := backoff.Retry(func() error {
		weburl := urls[rand.Intn(len(urls))]
		response, err = client.Get(weburl)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode > 400 {
			return errors.New(weburl + ": status code > 400")
		}
		return nil
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Second), 10)); err != nil {
		return "没图了~"
	}
	url := fmt.Sprintf("[CQ:image,file=%s]", response.Header.Get("Location"))
	log.Println("图片url: ", url)
	return url
}
