package tools

import (
	"net/url"
	"path/filepath"
	"pvp_spiderfly/src/model"
	"strings"
)

// GetUrlPath
// rawURL => https://xxx.com/aaa/vvv?a=1
// result => /aaa/vvv
func GetUrlPath(rawURL string) string {
	urlP, _ := url.Parse(rawURL)
	return urlP.Path
}

func GetDomain(rawURL string) string {
	urlP, _ := url.Parse(rawURL)
	return urlP.Scheme + "://" + urlP.Host
}

func GetFilePathByUrl(task *model.Task, url string) string {
	path := strings.Replace(url, task.EntryPointDomain, "", 1)
	fileName := ""
	if path == "/" {
		fileName = "index.html"
	} else {
		fileName = path
	}
	return filepath.Join(task.TargetDir, task.EntryPointHost, fileName)
}
