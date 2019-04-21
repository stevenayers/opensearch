/*
Fetches page data, converts the HTML into AlreadyCrawled, and formats the URLs
*/
package page

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
)

type (
	Page struct {
		Uid       string  `json:"uid,omitempty"`
		Url       string  `json:"url,omitempty"`
		Children  []*Page `json:"links,omitempty"`
		Parent    *Page   `json:"-"`
		Timestamp int64   `json:"timestamp,omitempty"`
	}
)

func (page *Page) FetchChildPages() (childPages []*Page, err error) {
	resp, err := http.Get(page.Url)
	if err != nil {
		log.Printf("failed to get URL %s: %v", page.Url, err)
		return
	}
	defer resp.Body.Close()                                               // Closes response body FetchUrls function is done.
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") { // Check if HTML file
		return
	}
	doc, err := parseHtml(resp)
	if err != nil {
		log.Printf("failed to parse HTML: %v", err)
		return
	}
	localProcessed := make(map[string]struct{}) // Ensures we don't store the same Url twice and
	// end up spawning 2 goroutines for same result
	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		href, ok := item.Attr("href")
		if ok && IsRelativeUrl(href) && IsRelativeHtml(href) && href != "" {
			absoluteUrl := ParseRelativeUrl(page.Url, href) // Standardises URL
			_, isPresent := localProcessed[absoluteUrl.Path]
			if !isPresent {
				localProcessed[absoluteUrl.Path] = struct{}{}
				childPage := Page{
					Url:       strings.TrimRight(absoluteUrl.String(), "/"),
					Parent:    page,
					Timestamp: time.Now().Unix(),
				}
				childPages = append(childPages, &childPage)
			}
		}
	})
	return
}

func (page *Page) MaxDepth() (countDepth int) {
	if page.Children != nil {
		var childDepths []int
		for _, childPage := range page.Children {
			childDepths = append(childDepths, childPage.MaxDepth())
		}
		return MaxIntSlice(childDepths) + 1
	} else {
		return 0
	}
}

func MaxIntSlice(v []int) int {
	sort.Ints(v)
	return v[len(v)-1]
}

func parseHtml(resp *http.Response) (doc *goquery.Document, err error) {

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	return
}

func ParseRelativeUrl(rootUrl string, relativeUrl string) (absoluteUrl *url.URL) {
	parsedRootUrl, err := url.Parse(rootUrl)
	if err != nil {
		return nil
	}
	absoluteUrl, err = url.Parse(parsedRootUrl.Scheme + "://" + parsedRootUrl.Host + path.Clean("/"+relativeUrl))
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
