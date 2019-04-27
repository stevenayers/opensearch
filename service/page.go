package service

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
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

	// Page holds page data
	Page struct {
		Uid       string  `json:"uid,omitempty"`
		Url       string  `json:"url,omitempty"`
		Links     []*Page `json:"links,omitempty"`
		Parent    *Page   `json:"-"`
		Timestamp int64   `json:"timestamp,omitempty"`
	}

	// JsonPage is used to turn Page into a dgraph compatible struct
	JsonPage struct {
		Uid       string      `json:"uid,omitempty"`
		Url       string      `json:"url,omitempty"`
		Timestamp int64       `json:"timestamp,omitempty"`
		Children  []*JsonPage `json:"links,omitempty"`
	}

	// JsonPredicate is used to hold the Predicate result from dgraph
	JsonPredicate struct {
		Matching int `json:"matching"`
	}
)

// FetchChildPages function converts http response into child page objects
func (page *Page) FetchChildPages(resp *http.Response) (childPages []*Page, err error) {
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		APILogger.LogDebug("context", "failed to parse HTML", "url", page.Url, "msg", err.Error())
		return
	}
	defer resp.Body.Close()
	localProcessed := make(map[string]struct{})
	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		href, ok := item.Attr("href")
		if ok && page.IsRelativeUrl(href) && page.IsRelativeHtml(href) && href != "" {
			absoluteUrl, _ := page.ParseRelativeUrl(href)
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

// MaxDepth function gets the max depth of the recursive page structure
func (page *Page) MaxDepth() (countDepth int) {
	if page.Links != nil {
		var childDepths []int
		for _, childPage := range page.Links {
			childDepths = append(childDepths, childPage.MaxDepth())
		}
		countDepth = maxIntSlice(childDepths) + 1
	}
	return
}

// ParseRelativeUrl function parses a relative URL string into a URL object
func (page *Page) ParseRelativeUrl(relativeUrl string) (absoluteUrl *url.URL, err error) {
	parsedRootUrl, err := url.Parse(page.Url)
	if err != nil {
		return nil, err
	}
	absoluteUrl, err = url.Parse(parsedRootUrl.Scheme + "://" + parsedRootUrl.Host + path.Clean("/"+relativeUrl))
	if err != nil {
		return nil, err
	}
	absoluteUrl.Fragment = "" // Removes '#' identifiers from Url
	return
}

// IsRelativeUrl function checks for relative URL path
func (page *Page) IsRelativeUrl(href string) bool {
	match, _ := regexp.MatchString("^(?:[a-zA-Z]+:)?//", href)
	return !match
}

// IsRelativeHtml function checks to see if relative URL points to a HTML file
func (page *Page) IsRelativeHtml(href string) bool {
	htmlMatch, _ := regexp.MatchString(`(\.html$)`, href) // Doesn't cover all allowed file extensions
	if htmlMatch {
		return htmlMatch
	}
	match, _ := regexp.MatchString(`(\.[a-zA-Z0-9]+$)`, href)
	return !match

}

// Converts JSONPage into a Page
func convertToPage(parentPage *Page, jsonPage *JsonPage) (currentPage *Page) {
	currentPage = &Page{
		Uid:       jsonPage.Uid,
		Url:       jsonPage.Url,
		Timestamp: jsonPage.Timestamp,
	}
	if parentPage != nil {
		currentPage.Parent = parentPage
	}
	wg := sync.WaitGroup{}
	convertPagesChan := make(chan *Page)
	for _, childJsonPage := range jsonPage.Children {
		wg.Add(1)
		go func(childJsonPage *JsonPage) {
			defer wg.Done()
			childPage := convertToPage(currentPage, childJsonPage)
			convertPagesChan <- childPage
		}(childJsonPage)
	}
	go func() {
		wg.Wait()
		close(convertPagesChan)

	}()
	for childPages := range convertPagesChan {
		currentPage.Links = append(currentPage.Links, childPages)
	}
	return
}

// Converts a Page to a JSONPage
func convertToJsonPage(currentPage *Page) (jsonPage JsonPage) {
	return JsonPage{
		Url:       currentPage.Url,
		Timestamp: currentPage.Timestamp,
	}
}

// Turns a Page into a JSON string
func serializePage(currentPage *Page) (pb []byte, err error) {
	p := convertToJsonPage(currentPage)
	pb, _ = json.Marshal(p)
	return
}

// Turns JSON dgraph result into a Page
func deserializePage(pb []byte) (currentPage *Page, err error) {
	jsonMap := make(map[string][]JsonPage)
	err = json.Unmarshal(pb, &jsonMap)
	jsonPages := jsonMap["result"]
	if len(jsonPages) > 0 {
		currentPage = convertToPage(nil, &jsonPages[0])
	}
	return
}

// Checks JSON dgraph edge result to see if edge exists
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

// Return max int in slice
func maxIntSlice(v []int) int {
	sort.Ints(v)
	return v[len(v)-1]
}
