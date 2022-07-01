package main

import (
	"context"
	"fmt"
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
	EntryPoint = "https://10.0.83.35/"
	//Cookie     = "USER_NAME_COOKIE=admin; OA_USER_ID=admin; PHPSESSID=ugko5e6pf6bc47lps4okodqpq4; SID_1=8d93b584"
	Cookie = ""
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

		chromedp.Sleep(10 * time.Second),
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

			wg.Add(1)
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

	_ = chromedp.Run(browser.Ctx, chromedp.Nodes("//*", &nodes, chromedp.BySearch))
	for _, itemNode := range nodes {
		if !config.AllowedClickNode.Contains(itemNode.LocalName) {
			continue
		}
		if tools.Contains(itemNode.Attributes, "logout") { //todo 待优化
			continue
		}
		fmt.Println("try click ... ")
		_ = chromedp.Run(browser.Ctx,
			chromedp.MouseClickNode(itemNode),
			chromedp.Sleep(3*time.Second), //todo 有没有更优的方式
		)
	}

	wg.Wait()

	browser.Close()
}
