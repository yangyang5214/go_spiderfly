package main

import (
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
	"net/url"
	"pvp_spiderfly/src/engine"
	"pvp_spiderfly/src/logger"
	"pvp_spiderfly/src/model"
	"pvp_spiderfly/src/tools"
	"strings"
	"sync"
)

const (
	ChromePath = "/Users/beer/java/chrome-mac/Chromium.app/Contents/MacOS/Chromium"
	Headless   = true
	TargetDir  = "./tmp"
	EntryPoint = "https://www.baidu.com"
)

func main() {

	urlParse, _ := url.Parse(EntryPoint)
	task := &model.Task{
		EntryPoint:       EntryPoint,
		EntryPointDomain: urlParse.Scheme + "://" + urlParse.Host,
		EntryPointHost:   urlParse.Host,
		TargetDir:        TargetDir,
	}

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
			wg.Add(1)
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

	err := chromedp.Run(browser.Ctx,
		network.Enable(),
		chromedp.Navigate(EntryPoint),
		network.SetExtraHTTPHeaders(browser.ExtraHeaders),
	)
	if err != nil {
		logger.Logger.Error(err)
	}
	wg.Wait()
	browser.Close()
}
