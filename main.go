package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ChromePath = "/Users/beer/java/chrome-mac/Chromium.app/Contents/MacOS/Chromium"
	Headless   = true
)

func main() {
	options := []chromedp.ExecAllocatorOption{
		//all config see allocate.go
		chromedp.ExecPath(ChromePath), //可忽略，直接用 chrome
		chromedp.Flag("headless", Headless),
		chromedp.Flag("ignore-certificate-errors", true), //ignoreHTTPSErrors
		chromedp.Flag("no-sandbox", true),
		chromedp.WindowSize(1920, 1080),
	}
	ctx_, cancel_ := chromedp.NewExecAllocator(
		context.Background(), options...)
	defer cancel_()
	ctx, cancel := chromedp.NewContext(ctx_)
	defer cancel()
	url := "https://www.beer5214.com/"

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			fmt.Println("event received:")
			fmt.Println(ev.Type)

			go func() {
				c := chromedp.FromContext(ctx)
				rbp := network.GetResponseBody(ev.RequestID)
				body, err := rbp.Do(cdp.WithExecutor(ctx, c.Target))
				if err != nil {
					fmt.Println(err)
				}

				path := strings.Replace(ev.Response.URL, url, "", 1)
				fileName := ""
				if path == "" {
					fileName = "index.html"
				} else {
					fileName = path
				}
				finalPath := filepath.Join("tmp", fileName)
				if err = os.MkdirAll(filepath.Dir(finalPath), os.ModePerm); err != nil {
					//ignore
				}
				if len(body) != 0 {
					if err = ioutil.WriteFile(finalPath, body, os.ModePerm); err != nil {
						log.Fatal(err)
					}
				}
			}()
		}
	})

	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.Sleep(time.Second*60), //todo
	)
	if err != nil {
		log.Println(err)
	}
}
