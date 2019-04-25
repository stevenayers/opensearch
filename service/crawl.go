package service

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// Crawler holds objects related to the crawler
type Crawler struct {
	AlreadyCrawled map[string]struct{}
	sync.Mutex
	DbWaitGroup sync.WaitGroup
}

// Crawl function adds page to db (in a goroutine so it doesn't stop initiating other crawls), gets the child pages then
// initiates crawls for each one.
func (crawler *Crawler) Crawl(currentPage *Page, depth int) {
	resp, err := http.Get(currentPage.Url)
	if err != nil {
		APILogger.LogDebug("context", "failed to get URL", "url", currentPage.Url, "msg", err.Error())
		return
	}
	crawler.DbWaitGroup.Add(1)
	go func(currentPage *Page) {
		defer crawler.DbWaitGroup.Done()
		err := DB.Create(currentPage)
		if err != nil {
			fmt.Print(currentPage.Url)
			panic(err)
		}
	}(currentPage)
	if crawler.hasAlreadyCrawled(currentPage.Url) || depth <= 0 ||
		!strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		return
	}
	pageWaitGroup := sync.WaitGroup{}
	childPagesChan := make(chan *Page)
	childPages, _ := currentPage.FetchChildPages(resp)
	for _, childPage := range childPages {
		pageWaitGroup.Add(1)
		go func(childPage *Page) {
			defer pageWaitGroup.Done()
			crawler.Crawl(childPage, depth-1)
			childPagesChan <- childPage
		}(childPage)
	}
	go func() {
		pageWaitGroup.Wait()
		close(childPagesChan)

	}()
	for childPages := range childPagesChan {
		currentPage.Links = append(currentPage.Links, childPages)
	}
}

// Locks crawl, then returns true/false dependent on Url being in map. If false, we store the Url.
func (crawler *Crawler) hasAlreadyCrawled(Url string) (isPresent bool) {
	cleanUrl := strings.TrimRight(Url, "/")
	defer crawler.Unlock()
	crawler.Lock()
	_, isPresent = crawler.AlreadyCrawled[cleanUrl]
	if !isPresent {
		crawler.AlreadyCrawled[cleanUrl] = struct{}{}
	}
	return
}
