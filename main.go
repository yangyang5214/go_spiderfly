package main

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
	"net/url"
	"pvp_spiderfly/src/config"
	"pvp_spiderfly/src/engine"
	"pvp_spiderfly/src/logger"
	"pvp_spiderfly/src/model"
	"pvp_spiderfly/src/tools"
	"strings"
	"sync"
	"time"
)

const (
	ChromePath = "/Users/beer/java/chrome-mac/Chromium.app/Contents/MacOS/Chromium"
	Headless   = false
	TargetDir  = "/Users/beer/beer/go_spiderfly/tmp"
	//EntryPoint = "http://10.0.83.172:5004/general/index.php?isIE=0&modify_pwd=0"
	EntryPoint = "https://10.0.83.35/owa/"
	//Cookie     = "USER_NAME_COOKIE=admin; OA_USER_ID=admin; PHPSESSID=ugko5e6pf6bc47lps4okodqpq4; SID_1=8d93b584"
	Cookie = "X-BackEndCookie=S-1-5-21-3957220163-591661206-1592131018-500=u56Lnp2ejJqBx8nKnZ3Jz5nSzp6czdLLx8/K0sbGmZvSmsaezMvMxsmdyc6ZgYHNz83N0s/I0szOq8/NxczPxczH; PrivateComputer=true; ClientId=3186DDAB19A54C6289E26474B5A72EC9; X-OWA-JS-PSD=1; PBack=0; cadata=n7rgVF27JmXkSJtcbn5h8sByemJYRK4wuc0/tW5Qnyc/XaRR3soAoaB5yXX5DdTn3LdMwne6ydWN3PC7WoihEq9Yova0s6loRUZtpP6ovIIFiddxGz2coHMVbzLZih3Q; cadataTTL=050xk/3ZQ74/4KRg+1oR5Q==; cadataKey=Pgmgf8HgXnojhZB6hXH5bcb30RC32P2Xz4sHP+S28kwsFOKRkW4ho24vnzJ02XX09rIQlc95NEcKG/7ETlKVaOb2mzeCGUquZBLdrvm7MWf059v7AbBfWHJu6xaFJDmpg0aBeSbevVECLYNLT8TyUdn5SFwBmXCpI8vLda9DBYugkuk6veYgMRA1Sb/AbchFPAUal3OmywljZS6ko2QHFhfFpWcNONwmdKpyuicdpxMXLNUp28fk6VSMLygtxyOPDb62lKd3Id4R1PdlZVdG0ZGWkdv6NfSQcr3y/2IHWK+EYIMEIvrhx6EsQWN5fqSJABLkjnXomik54bFDL/WWbQ==; cadataIV=Hz0E0N5J70bijByfo3HU4iBDqCkaveEZAAtzQE1NM0+92ywLWBPwCb+g74HIMshXzbJk57IhpppSIKFY/qunlSWZRhljBt+OXKqtZ/q1hCB7F+RRv82ASjXPxyHcr3ddSGoWpLU274JdQ9Jp20j2W/Leiz10rWGh6oSAaMr/QcrYNvYw4IA8DnZ1a5MSPxfLJtwT5zEOFnCj3a4fwy+OhAF/d5afRBTEGopyq8Co7BFKeIBvgPg4AfUo57YTO3AX1hYEsoDp4yH06L3n958q7dFNh/HlP1Fdb0TPuglxbdziRrXXj3DTxgGYeQ19KezpzPsjhMcnFHW/cIdTHUuUCw==; cadataSig=KK21h7fxdjmkIN/tvTZG7uzRtZyQEObLRDJ507E3zHnojVyvTDulK4P7BqzwEeBRRcuJAoFvWtRgx7J59OT44aO+bVfIRilwxIZzL1ueUZSfzkPmsoPsTow7t7AR4fm+rij3j34h3P65Mor3S5S+Mmxvb5H/8hqqBeLG7dN9FTrB6uesJSEB9D4x5zO/tvMR52WNwsqAqGkyHjHSlylOVncgNrzTxXQosXkLh60fGBRpZ6osoMvtXYb409vI8D1L/UUz/xCOThjFyZzi260cY4WuxqU/+gmQMGxYgHa+KWBISVRRW9jlpzj5pt3l0X62s4mjn34FWuIiHYIYvBD2Tg==; UC=3cfc4a83ac234a479d29270ee0fa6ac0; X-OWA-CANARY=5wtFCRe9ckqpzpB-ff7NKTDReaAJW9oIz0DCLzp4AvwXn5yMANLpERb4OmHvdCOd6tIbOS1UdAA."
)

func TaskActions(task model.Task) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// create cookie expiration
			expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
			// add cookies to chrome
			cookies := strings.Split(task.Cookie, ";")
			for i := 0; i < len(cookies); i++ {
				cookieArr := strings.Split(cookies[i], "=")
				err := network.SetCookie(strings.Trim(cookieArr[0], " "), strings.Trim(cookieArr[1], "")).
					WithExpires(&expr).
					WithDomain(task.EntryPointHost).
					WithHTTPOnly(false).
					Do(ctx)
				if err != nil {
					logger.Logger.Error(err)
					return err
				}
			}
			return nil
		}),
		network.Enable(),
		chromedp.Navigate(task.EntryPoint),
		network.SetExtraHTTPHeaders(task.ExtraHeaders),
		chromedp.Sleep(time.Second * 10),
	}
}

func main() {

	extraHeaders := map[string]interface{}{}
	extraHeaders["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"

	urlParse, _ := url.Parse(EntryPoint)
	task := &model.Task{
		EntryPoint:     EntryPoint,
		Domain:         urlParse.Scheme + "://" + urlParse.Host,
		EntryPointHost: urlParse.Host,
		TargetDir:      TargetDir,
		ExtraHeaders:   extraHeaders,
		Cookie:         Cookie,
	}

	logger.Logger.WithFields(logrus.Fields{
		"EntryPoint": EntryPoint,
	}).Info()

	var wg sync.WaitGroup

	browser := engine.InitBrowser(ChromePath, extraHeaders, Headless)

	//request_id => response_url
	//request_url => redirect_url
	urlMap := map[string]string{}

	chromedp.ListenTarget(browser.Ctx, func(event interface{}) {
		switch ev := event.(type) {
		case *page.EventJavascriptDialogOpening:
			go func() {
				_ = chromedp.Run(browser.Ctx, page.HandleJavaScriptDialog(false)) //主要为了屏蔽登出
			}()
		case *network.EventRequestWillBeSent:
			wg.Add(1)
			if ev.RedirectResponse != nil {
				urlMap[ev.DocumentURL] = ev.RedirectResponse.URL
			}
		case *network.EventLoadingFinished:
			logger.Logger.WithFields(logrus.Fields{
				"EventLoadingFinished-RequestID": ev.RequestID.String(),
			}).Debug()

			localUrl := urlMap[ev.RequestID.String()]
			delete(urlMap, ev.RequestID.String())

			go func() {
				c := chromedp.FromContext(browser.Ctx)
				body, err := network.GetResponseBody(ev.RequestID).Do(cdp.WithExecutor(browser.Ctx, c.Target))
				if err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"GetResponseBody": err.Error(),
					}).Error()
					defer wg.Done()
					return
				}
				finalPath := tools.GetFilePathByUrl(task, localUrl)

				tools.WriteFile(finalPath, body)

				redirectUrl := urlMap[localUrl]
				if redirectUrl != "" {
					tools.WriteFile(tools.GetFilePathByUrl(task, redirectUrl), body)
				}
				defer wg.Done()
			}()
		case *network.EventResponseReceived:
			if strings.HasPrefix(ev.Response.URL, "data") {
				return // skip local memory cache)
			}
			logger.Logger.WithFields(logrus.Fields{
				"EventResponseReceived-RequestID": ev.RequestID.String(),
				"EventResponseReceived-URL":       ev.Response.URL,
			}).Debug()
			urlMap[ev.RequestID.String()] = ev.Response.URL
		}
	})

	if err := chromedp.Run(browser.Ctx, TaskActions(*task)); err != nil {
		logger.Logger.Error(err)
	}

	var nodes []*cdp.Node

	_ = chromedp.Run(browser.Ctx, chromedp.Nodes("//*", &nodes))
	for _, itemNode := range nodes {
		if !config.AllowedClickNode.Contains(itemNode.LocalName) {
			continue
		}
		if tools.Contains(itemNode.Attributes, "logout") { //todo 待优化
			continue
		}
		_ = chromedp.Run(browser.Ctx,
			chromedp.MouseClickNode(itemNode),
			chromedp.Sleep(3*time.Second), //todo 有没有更优的方式
		)
	}

	wg.Wait()

	browser.Close()
}
