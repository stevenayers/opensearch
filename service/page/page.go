/*
Fetches page data, converts the HTML into AlreadyCrawled, and formats the URLs
*/
package page

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
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
	Links     []*Page   `json:"links"`
	Depth     int       `json:"depth"`
	Timestamp time.Time `json:"timestamp"`
	Body      string    `json:"body"`
}

func (page *Page) FetchUrls() (urls []*url.URL, err error) {
	resp, err := http.Get(page.Url.String())
	if err != nil {
		log.Printf("failed to get URL %s: %v", page.Url.String(), err)
		return
	}
	doc, err := parseDoc(resp)
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
				urls = append(urls, absoluteUrl)
			}
		}
	})
	return
}

func parseDoc(resp *http.Response) (doc *goquery.Document, err error) {
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") { // Check if HTML file
		err = errors.New("Content-Type header not 'text/hml'")
		return
	}
	defer resp.Body.Close() // Closes response body FetchUrls function is done.
	doc, err = goquery.NewDocumentFromReader(resp.Body)
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
