/*
Controls the crawling through a website's structure, also manages the crawl state.
*/
package crawl

import (
	"fmt"
	"go-clamber/database"
	"go-clamber/page"
	"strings"
	"sync"
)

type Crawler struct { // Struct to manage Crawl state in one place.
	AlreadyCrawled map[string]struct{}
	sync.Mutex
}

func (crawler *Crawler) Crawl(currentPage *page.Page, depth int) {
	err := database.DB.Create(currentPage)
	if err != nil {
		fmt.Print(currentPage.Url)
		panic(err)
	}
	if depth <= 0 {
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
			//fmt.Printf("---%s\n", childPage.Url)
			crawler.Crawl(childPage, depth-1)
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
}

func (crawler *Crawler) hasAlreadyCrawled(Url string) (isPresent bool) {
	/*
		Locks crawl, then returns true/false dependent on Url being in map.
		If false, we store the Url.
	*/
	cleanUrl := strings.TrimRight(Url, "/")
	defer crawler.Unlock()
	crawler.Lock()
	_, isPresent = crawler.AlreadyCrawled[cleanUrl]
	if !isPresent {
		crawler.AlreadyCrawled[cleanUrl] = struct{}{}
	}
	return
}
