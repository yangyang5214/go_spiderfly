package main

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	mapset "github.com/deckarep/golang-set"
	"github.com/sirupsen/logrus"
	"net/url"
	"path/filepath"
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
	//EntryPoint = "http://10.0.83.172:5004/"
	EntryPoint = "https://10.0.83.35/owa/"
	Cookie = ""
)

var nodeMap = mapset.NewSet()

func TaskActions(task model.Task) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// create cookie expiration
			expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
			// add cookies to chrome
			cookies := strings.Split(task.Cookie, ";")
			for i := 0; i < len(cookies); i++ {
				cookieArr := strings.Split(cookies[i], "=")
				if len(cookieArr) != 2 {
					continue
				}
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

		//for owa
		chromedp.SetValue("document.querySelector('#userName')", "MING/Administrator", chromedp.ByJSPath),
		chromedp.SetValue("document.querySelector('#password')", "TCC@202206", chromedp.ByJSPath),
		chromedp.Click("#lgnDiv > div.signInEnter > div", chromedp.ByQuery),
		//
		//for owa
		//chromedp.SetValue("//*[@id='name']", "admin", chromedp.BySearch),
		//chromedp.SetValue("//*[@id='password']", "admin123", chromedp.BySearch),
		//chromedp.Click("//*[@class='login_btn']", chromedp.BySearch),

		chromedp.Sleep(10 * time.Second),

		chromedp.ActionFunc(func(ctx context.Context) error {
			var res []byte
			chromedp.FullScreenshot(&res, 100).Do(ctx)
			tools.WriteFile(filepath.Join(task.TargetDir, task.EntryPointHost, "main.jpg"), res)
			return nil
		}),
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
			if ev.RedirectResponse != nil {
				urlMap[ev.DocumentURL] = ev.RedirectResponse.URL
			}
		case *network.EventLoadingFinished:
			logger.Logger.WithFields(logrus.Fields{
				"EventLoadingFinished-RequestID": ev.RequestID.String(),
			}).Debug()

			localUrl := urlMap[ev.RequestID.String()]
			delete(urlMap, ev.RequestID.String())

			if localUrl == "" {
				return
			}

			wg.Add(1)
			go func() {
				c := chromedp.FromContext(browser.Ctx)
				body, err := network.GetResponseBody(ev.RequestID).Do(cdp.WithExecutor(browser.Ctx, c.Target))
				if err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"Url":             localUrl,
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

	_ = chromedp.Run(browser.Ctx, chromedp.Nodes("//*", &nodes, chromedp.BySearch))

	ClickNodes(nodes, browser.Ctx)

	wg.Wait()

	browser.Close()
}

func ClickNodes(nodes []*cdp.Node, ctx context.Context) {
	for _, itemNode := range nodes {
		hash := tools.Md5(tools.StringToByte(append(itemNode.Attributes, itemNode.LocalName)))
		if nodeMap.Contains(hash) {
			continue
		}
		if !config.AllowedClickNode.Contains(itemNode.LocalName) {
			continue
		}
		if tools.Contains(itemNode.Attributes, "logout") { //todo 待优化
			continue
		}
		logger.Logger.WithFields(logrus.Fields{
			"Attributes": itemNode.Attributes,
			"LocalName":  itemNode.LocalName,
		}).Info("try to click node ...")

		_ = chromedp.Run(ctx,
			chromedp.MouseClickNode(itemNode),
			chromedp.Sleep(5*time.Second), //todo 有没有更优的方式 (一些页面有 lading... 等待 5s)
		)
		nodeMap.Add(hash)
		ClickNodes(itemNode.Children, ctx)
	}
}
