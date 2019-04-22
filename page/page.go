/*
Fetches page data, converts the HTML into AlreadyCrawled, and formats the URLs
*/
package page

import (
	"clamber/utils"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"strings"
	"sync"
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
	doc, err := utils.ParseHtml(resp)
	if err != nil {
		log.Printf("failed to parse HTML: %v", err)
		return
	}
	localProcessed := make(map[string]struct{}) // Ensures we don't store the same Url twice and
	// end up spawning 2 goroutines for same result
	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		href, ok := item.Attr("href")
		if ok && utils.IsRelativeUrl(href) && utils.IsRelativeHtml(href) && href != "" {
			absoluteUrl := utils.ParseRelativeUrl(page.Url, href) // Standardises URL
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
		return utils.MaxIntSlice(childDepths) + 1
	} else {
		return 0
	}
}

func ConvertToPage(parentPage *Page, jsonPage *utils.JsonPage) (currentPage *Page) {
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
		go func(childJsonPage *utils.JsonPage) {
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

func ConvertToJsonPage(currentPage *Page) (jsonPage utils.JsonPage) {
	return utils.JsonPage{
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
	jsonMap := make(map[string][]utils.JsonPage)
	err = json.Unmarshal(pb, &jsonMap)
	jsonPages := jsonMap["result"]
	if len(jsonPages) > 0 {
		currentPage = ConvertToPage(nil, &jsonPages[0])
	}

	return
}
