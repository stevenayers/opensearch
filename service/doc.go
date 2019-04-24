/*
Package service provides the clamber crawling package.

To initiate a crawl, create a Crawler with an empty sync.WaitGroup and struct map. DbWaitGroup is needed to ensure the
clamber process does not exit before the crawler is done writing to the database. AlreadyCrawled keeps track of the
URLs which have been crawled already in that crawl process.

	crawler = service.Crawler{
		DbWaitGroup: sync.WaitGroup{},
		AlreadyCrawled: make(map[string]struct{})
	}

Create a page object with the starting URL of your crawl.

	page := &service.Page{Url: "https://golang.org"}

Call Crawl on the Crawler object, passing in your page, and the depth of the crawl you want.

	crawler.Crawl(result, 5)

Ensure your go process does not end before the crawled data has been saved to dgraph. If you need more logic to execute
first, put the line below after this, as your application will hang on Wait() until we're done writing.

	crawler.DbWaitGroup.Wait()

*/
package service
