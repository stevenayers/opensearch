/*
 Package page fetches page data, converts the HTML into AlreadyCrawled, and formats the URLs
*/
package page

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type (
	Page struct {
		Uid       string  `json:"uid,omitempty"`
		Url       string  `json:"url,omitempty"`
		Children  []*Page `json:"links,omitempty"`
		Parent    []*Page `json:"parents,omitempty"`
		Timestamp int64   `json:"timestamp,omitempty"`
	}

	JsonPage struct {
		Uid       string      `json:"uid,omitempty"`
		Url       string      `json:"url,omitempty"`
		Timestamp int64       `json:"timestamp,omitempty"`
		Children  []*JsonPage `json:"links,omitempty"`
	}

	JsonPredicate struct {
		Matching int `json:"matching"`
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
	doc, err := ParseHtml(resp)
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
					Parent:    []*Page{page},
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
		countDepth = maxIntSlice(childDepths) + 1
	}
	return
}

func ConvertToPage(parentPage *Page, jsonPage *JsonPage) (currentPage *Page) {
	currentPage = &Page{
		Uid:       jsonPage.Uid,
		Url:       jsonPage.Url,
		Timestamp: jsonPage.Timestamp,
	}
	if parentPage != nil {
		currentPage.Parent = []*Page{parentPage}
	}
	wg := sync.WaitGroup{}
	convertPagesChan := make(chan *Page)
	for _, childJsonPage := range jsonPage.Children {
		wg.Add(1)
		go func(childJsonPage *JsonPage) {
			defer wg.Done()
			childPage := ConvertToPage(currentPage, childJsonPage)
			convertPagesChan <- childPage
		}(childJsonPage)
	}
	go func() {
		wg.Wait()
		close(convertPagesChan)

	}()
	for childPages := range convertPagesChan {
		currentPage.Children = append(currentPage.Children, childPages)
	}
	return
}

func ConvertToJsonPage(currentPage *Page) (jsonPage JsonPage) {
	return JsonPage{
		Url:       currentPage.Url,
		Timestamp: currentPage.Timestamp,
	}
}

func SerializePage(currentPage *Page) (pb []byte, err error) {
	p := ConvertToJsonPage(currentPage)
	pb, err = json.Marshal(p)
	if err != nil {
		fmt.Print(err)
	}
	return
}

func DeserializePage(pb []byte) (currentPage *Page, err error) {
	jsonMap := make(map[string][]JsonPage)
	err = json.Unmarshal(pb, &jsonMap)
	jsonPages := jsonMap["result"]
	if len(jsonPages) > 0 {
		currentPage = ConvertToPage(nil, &jsonPages[0])
	}
	return
}

func ParseHtml(resp *http.Response) (doc *goquery.Document, err error) {
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

func DeserializePredicate(pb []byte) (exists bool, err error) {
	jsonMap := make(map[string][]JsonPredicate)
	err = json.Unmarshal(pb, &jsonMap)
	if err != nil {
		return
	}
	edges := jsonMap["edges"]
	if len(edges) > 0 {
		exists = edges[0].Matching > 0
	} else {
		exists = false
	}
	return
}

func maxIntSlice(v []int) int {
	sort.Ints(v)
	return v[len(v)-1]
}
