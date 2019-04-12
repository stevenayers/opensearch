/*
Fetches page data, converts the HTML into AlreadyCrawled, and formats the URLs
*/
package page

import (
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

type Page struct {
	Url       *url.URL  `json:"url"`
	Children  []*Page   `json:"links"`
	Parent    *Page     `json:"parent"`
	Depth     int       `json:"depth"`
	Timestamp time.Time `json:"timestamp"`
	Body      string    `json:"body"`
}

func (page *Page) FetchChildPages() (childPages []*Page, err error) {
	resp, err := http.Get(page.Url.String())
	if err != nil {
		log.Printf("failed to get URL %s: %v", page.Url.String(), err)
		return
	}
	defer resp.Body.Close()                                               // Closes response body FetchUrls function is done.
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") { // Check if HTML file
		return
	}
	doc, body, err := parseHtml(resp)
	if err != nil {
		log.Printf("failed to parse HTML: %v", err)
		return
	}

	localProcessed := make(map[string]struct{}) // Ensures we don't store the same Url twice and
	// end up spawning 2 goroutines for same result
	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		href, ok := item.Attr("href")
		if ok && IsRelativeUrl(href) && IsRelativeHtml(href) && href != "" {
			absoluteUrl := ParseRelativeUrl(page.Url, strings.TrimRight(href, "/")) // Standardises URL
			_, isPresent := localProcessed[absoluteUrl.Path]
			if !isPresent {
				localProcessed[absoluteUrl.Path] = struct{}{}
				childPage := Page{
					Url:    absoluteUrl,
					Parent: page,
					Depth:  page.Depth - 1,
					Body:   body,
				}
				childPages = append(childPages, &childPage)
			}
		}
	})
	return
}

func parseHtml(resp *http.Response) (doc *goquery.Document, body string, err error) {
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	body = string(bodyBytes)
	return
}

func ParseRelativeUrl(rootUrl *url.URL, relativeUrl string) (absoluteUrl *url.URL) {
	absoluteUrl, err := url.Parse(rootUrl.Scheme + "://" + rootUrl.Host + path.Clean("/"+relativeUrl))
	if err != nil {
		return nil
	}
	absoluteUrl.Fragment = "" // Removes '#' identifiers from Url
	return
}

func IsRelativeUrl(href string) bool {
	match, _ := regexp.MatchString("^(?:[a-zA-Z]+:)?//", href)
	return !match
}

func IsRelativeHtml(href string) bool {
	htmlMatch, _ := regexp.MatchString(`(\.html$)`, href) // Doesn't cover all allowed file extensions
	if htmlMatch {
		return htmlMatch
	} else {
		match, _ := regexp.MatchString(`(\.[a-zA-Z0-9]+$)`, href)
		return !match
	}

}
