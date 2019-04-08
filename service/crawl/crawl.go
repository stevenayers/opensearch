/*
Controls the crawling through a website's structure, also manages the crawl state.
*/
package crawl

import (
	"go-clamber/service/page"
	"net/url"
	"strings"
	"sync"
)

type Crawler struct { // Struct to manage Crawl state in one place.
	AlreadyCrawled map[string]struct{}
	sync.Mutex
}

func (crawler *Crawler) Crawl(currentPage *page.Page) {
	if currentPage.Depth <= 0 {
		return
	}
	if crawler.hasAlreadyCrawled(currentPage.Url) {
		return
	}
	wg := sync.WaitGroup{}
	childPagesChan := make(chan *page.Page)
	childUrls, _ := currentPage.FetchUrls()
	for _, childUrl := range childUrls { // Iterate through links found on currentPage
		wg.Add(1)
		go func(childUrl *url.URL) { // create goroutines for each link found and crawl the child currentPage
			defer wg.Done()
			childPage := page.Page{Url: childUrl, Depth: currentPage.Depth - 1}
			crawler.Crawl(&childPage)
			childPagesChan <- &childPage
		}(childUrl)
	}
	go func() { // Close channel when direct child pages have returned
		wg.Wait()
		close(childPagesChan)
	}()
	for childPages := range childPagesChan { // Feed channel values into slice, possibly performance inefficient.
		currentPage.Links = append(currentPage.Links, childPages)
	}
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
