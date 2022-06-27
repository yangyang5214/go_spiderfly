package engine

import (
	"context"
	"github.com/chromedp/chromedp"
	"pvp_spiderfly/src/logger"
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
