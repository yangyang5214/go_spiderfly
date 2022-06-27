package tools

import (
	"net/url"
	"path/filepath"
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

func GetFilePathByUrl(domain string, targetDir string, url string) string {
	path := strings.Replace(url, domain, "", 1)
	fileName := ""
	if path == "/" {
		fileName = "index.html"
	} else {
		fileName = path
	}
	return filepath.Join(targetDir, fileName)
}
