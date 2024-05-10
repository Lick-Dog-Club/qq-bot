package pixiv

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"qq/config"
	"strings"
	"time"
)

func login(user, pass string) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"),
	)

	// 创建一个浏览器上下文环境
	ctx, cancel := chromedp.NewExecAllocator(context.TODO(), opts...)
	defer cancel()

	// 创建浏览器上下文
	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// 设置超时
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	// 要获取的网页地址
	var url = "https://accounts.pixiv.net/login?return_to=https%3A%2F%2Fwww.pixiv.net%2F&lang=zh&source=pc&view_type=page"

	loginBtn := "//*[@id=\"app-mount-point\"]/div/div/div[4]/div[1]/div[2]/div/div/div/form/fieldset[1]/label/input"
	passBtn := "//*[@id=\"app-mount-point\"]/div/div/div[4]/div[1]/div[2]/div/div/div/form/fieldset[2]/label/input"
	submitBtn := "//*[@id=\"app-mount-point\"]/div/div/div[4]/div[1]/div[2]/div/div/div/form/button"
	// 运行任务
	return chromedp.Run(ctx,
		chromedp.Sleep(time.Duration(rand.Intn(1000))*time.Millisecond),
		// 打开网页
		chromedp.Navigate(url),
		// 等待网页加载完成，淘宝会检查环境
		chromedp.Tasks{
			chromedp.WaitVisible(loginBtn),
			chromedp.Click(loginBtn),
			chromedp.SendKeys(loginBtn, user),
			chromedp.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond),
			chromedp.Click(passBtn),
			chromedp.SendKeys(passBtn, pass),
			chromedp.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond),
			chromedp.Click(submitBtn),
		},
		chromedp.ActionFunc(func(ctx context.Context) error {
			for {
				chromedp.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond).Do(ctx)
				var location string
				chromedp.Location(&location).Do(ctx)
				fmt.Println(location)
				if strings.HasPrefix(location, "https://www.pixiv.net/") {
					break
				}
				chromedp.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond).Do(ctx)
			}
			cookies, _ := network.GetCookies().Do(ctx)
			for _, cookie := range cookies {
				if cookie.Name == "PHPSESSID" {
					log.Printf("%v: %v\n", cookie.Name, cookie.Value)
					config.Set(map[string]string{"pixiv_session": cookie.Value})
					break
				}
			}

			return nil
		}),
	)
}
