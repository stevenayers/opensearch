/*
Controls the crawling through a website's structure, also manages the crawl state.
*/
package crawl

import (
	"golang-webcrawler/fetch"
	"net/url"
	"strings"
	"sync"
)

type Crawler struct { // Struct to manage Crawl state in one place.
	AlreadyCrawled map[string]struct{}
	sync.Mutex
}

func (crawler *Crawler) Crawl(page *fetch.Page) {
	if page.Depth <= 0 {
		return
	}
	if crawler.hasAlreadyCrawled(page.Url) {
		return
	}
	wg := sync.WaitGroup{}
	childPagesChan := make(chan *fetch.Page)
	childUrls, _ := page.FetchUrls()
	for _, childUrl := range childUrls { // Iterate through links found on page
		wg.Add(1)
		go func(childUrl *url.URL) { // create goroutines for each link found and crawl the child page
			defer wg.Done()
			childPage := fetch.Page{Url: childUrl, Depth: page.Depth - 1}
			crawler.Crawl(&childPage)
			childPagesChan <- &childPage
		}(childUrl)
	}
	go func() { // Close channel when direct child pages have returned
		wg.Wait()
		close(childPagesChan)
	}()
	for childPages := range childPagesChan { // Feed channel values into slice, possibly performance inefficient.
		page.Links = append(page.Links, childPages)
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
