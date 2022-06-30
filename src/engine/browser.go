package engine

import (
	"context"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"pvp_spiderfly/src/logger"
	"strings"
)

type Browser struct {
	Ctx          context.Context
	Cancel       context.CancelFunc
	ExtraHeaders map[string]interface{}
}

func InitBrowser(chromiumPath string, extraHeaders map[string]interface{}, hasHeadless bool) *Browser {
	var bro Browser
	options := []chromedp.ExecAllocatorOption{
		//all config see allocate.go
		chromedp.ExecPath(chromiumPath),
		chromedp.Flag("headless", hasHeadless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("ignore-certificate-errors", true), //ignoreHTTPSErrors
		chromedp.Flag("no-sandbox", true),
		chromedp.WindowSize(1920, 1080),
	}
	bgCtx, _ := chromedp.NewExecAllocator(
		context.Background(), options...)
	ctx, cancel := chromedp.NewContext(bgCtx)
	bro.Cancel = cancel
	bro.Ctx = ctx
	bro.ExtraHeaders = extraHeaders
	return &bro
}

func (bro *Browser) Close() {
	logger.Logger.Info("closing browser.")
	(bro.Cancel)()
}

func (bro *Browser) GetCookies(cookie string) []*network.CookieParam {
	var cookies []*network.CookieParam
	if cookie == "" {
		return cookies
	}
	cookieArr := strings.Split(cookie, ";")
	for i := 0; i < len(cookieArr); i++ {
		cookieArrs := strings.Split(cookieArr[i], "=")
		cookies = append(cookies, &network.CookieParam{
			Name:  strings.Trim(cookieArrs[0], " "),
			Value: strings.Trim(cookieArrs[1], " "),
		})
	}
	return cookies
}
