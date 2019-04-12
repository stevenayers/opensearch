/*
Controls the crawling through a website's structure, also manages the crawl state.
*/
package crawl

import (
	"go-clamber/database"
	"go-clamber/page"
	"net/url"
	"strings"
	"sync"
)

type Crawler struct { // Struct to manage Crawl state in one place.
	AlreadyCrawled map[string]struct{}
	sync.Mutex
}

func (crawler *Crawler) Crawl(currentPage *page.Page, DB database.Store) {
	if currentPage.Depth <= 0 {
		return
	}
	if crawler.hasAlreadyCrawled(currentPage.Url) {
		return
	}
	wg := sync.WaitGroup{}
	childPagesChan := make(chan *page.Page)
	childPages, _ := currentPage.FetchChildPages()
	for _, childPage := range childPages { // Iterate through links found on currentPage
		wg.Add(1)
		go func(childPage *page.Page) { // create goroutines for each link found and crawl the child currentPage
			defer wg.Done()

			crawler.Crawl(childPage, DB)
			childPagesChan <- childPage
		}(childPage)
	}
	go func() { // Close channel when direct child pages have returned
		wg.Wait()
		close(childPagesChan)

	}()
	for childPages := range childPagesChan { // Feed channel values into slice, possibly performance inefficient.
		currentPage.Children = append(currentPage.Children, childPages)
	}
	_ = DB.Create(currentPage)
}

func (crawler *Crawler) hasAlreadyCrawled(Url *url.URL) (isPresent bool) {
	/*
		Locks crawl, then returns true/false dependent on Url being in map.
		If false, we store the Url.
	*/
	defer crawler.Unlock()
	cleanUrl := strings.TrimRight(Url.String(), "/")
	crawler.Lock()
	_, isPresent = crawler.AlreadyCrawled[cleanUrl]
	if !isPresent {
		crawler.AlreadyCrawled[cleanUrl] = struct{}{}
	}
	return
}
