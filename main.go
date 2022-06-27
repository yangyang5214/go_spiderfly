package main

import (
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
	"pvp_spiderfly/src/engine"
	"pvp_spiderfly/src/logger"
	"pvp_spiderfly/src/tools"
	"sync"
	"time"
)

const (
	ChromePath = "/Users/beer/java/chrome-mac/Chromium.app/Contents/MacOS/Chromium"
	Headless   = false
	TargetDir  = "./tmp"
	EntryPoint = "https://10.0.83.1/web/frame/login.html"
)

func main() {

	domain := tools.GetDomain(EntryPoint)

	logger.Logger.WithFields(logrus.Fields{
		"EntryPoint": EntryPoint,
	}).Info()

	var wg sync.WaitGroup
	extraHeaders := map[string]interface{}{}
	browser := engine.InitBrowser(ChromePath, extraHeaders, Headless)

	//request_id => response_url
	//request_url => redirect_url
	urlMap := map[string]string{}

	chromedp.ListenTarget(browser.Ctx, func(event interface{}) {
		switch ev := event.(type) {
		case *network.EventRequestWillBeSent:
			if ev.RedirectResponse != nil {
				urlMap[ev.DocumentURL] = ev.RedirectResponse.URL
			}
		case *network.EventLoadingFinished:
			wg.Add(1)
			logger.Logger.WithFields(logrus.Fields{
				"EventLoadingFinished-RequestID": ev.RequestID.String(),
			}).Debug()
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
				url := urlMap[ev.RequestID.String()]
				finalPath := tools.GetFilePathByUrl(domain, TargetDir, url)

				tools.WriteFile(finalPath, body)

				redirectUrl := urlMap[url]
				if redirectUrl != "" {
					tools.WriteFile(tools.GetFilePathByUrl(domain, TargetDir, redirectUrl), body)
				}
				defer wg.Done()
			}()

		case *network.EventResponseReceived:
			logger.Logger.WithFields(logrus.Fields{
				"EventResponseReceived-RequestID": ev.RequestID.String(),
				"EventResponseReceived-URL":       ev.Response.URL,
			}).Debug()
			urlMap[ev.RequestID.String()] = ev.Response.URL
		}
	})

	err := chromedp.Run(browser.Ctx,
		network.Enable(),
		chromedp.Navigate(EntryPoint),
		chromedp.Sleep(time.Second*10), //todo
	)
	if err != nil {
		logger.Logger.Error(err)
	}
	wg.Wait()
	browser.Close()
}
