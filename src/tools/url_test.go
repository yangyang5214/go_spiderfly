package tools

import (
	"fmt"
	"net/url"
	"testing"
)

// GetUrlPath
// rawURL => https://xxx.com/aaa/vvv?a=1
// result => /aaa/vvv
func TestGetUrlPath(t *testing.T) {
	rawURL := "http://0.0.0.0:8088/static/templates/2021_year_01/img_happy%20new%20year.png"
	urlP, _ := url.Parse(rawURL)
	fmt.Println(urlP.Path) //static/templates/2021_year_01/img_happy new year.png
}
