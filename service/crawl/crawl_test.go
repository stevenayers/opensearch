package crawl_test

import (
	"golang-webcrawler/crawl"
	"golang-webcrawler/fetch"
	"net/url"
	"testing"
)

type CrawlTest struct {
	Url   string
	Depth int
}

var CrawlTests = []CrawlTest{
	{"https://golang.org/", 1},
	{"https://golang.org/", 5},
	{"https://golang.org/", 10},
	{"http://example.com", 1},
	{"http://example.com", 5},
	{"http://example.com", 10},
	{"https://google.com", 1},
	{"https://google.com", 5},
	{"https://google.com", 10},
}

var PageReturnTests = []string{
	"https://golang.org/",
	"http://example.com",
	"https://google.com",
}

func TestAlreadyCrawled(t *testing.T) {
	for _, test := range CrawlTests {
		crawler := crawl.Crawler{AlreadyCrawled: make(map[string]struct{})}
		rootUrl, _ := url.Parse(test.Url)
		rootPage := fetch.Page{Url: rootUrl, Depth: test.Depth}
		crawler.Crawl(&rootPage)
		for Url := range crawler.AlreadyCrawled { // Iterate through crawled AlreadyCrawled and recursively search for each one
			var countedDepths []int
			crawledCounter := 0
			recursivelySearchPages(t, &rootPage, Url, &crawledCounter, &countedDepths)()
		}
	}
}

func TestAllPagesReturned(t *testing.T) {
	for _, testUrl := range PageReturnTests {
		crawler := crawl.Crawler{AlreadyCrawled: make(map[string]struct{})}
		rootUrl, _ := url.Parse(testUrl)
		rootPage := fetch.Page{Url: rootUrl, Depth: 1}
		Urls, _ := rootPage.FetchUrls()

		crawler.Crawl(&rootPage)
		if len(rootPage.Links) != len(Urls) {
			t.Fatalf("%s: page.Links length (%d) did not match FetchUrls output length (%d).",
				testUrl, len(rootPage.Links), len(Urls))
		}
	}
}

// Helper Function for TestAlreadyCrawled. Cleanest way to implement recursive map checking into the test.
func recursivelySearchPages(t *testing.T, p *fetch.Page, Url string, counter *int, depths *[]int) func() {
	return func() {
		for _, v := range p.Links {
			if v.Links != nil && v.Url.String() == Url { // Check if page has links
				*depths = append(*depths, v.Depth) // Log the depth it was counted (useful when inspecting data structure)
				*counter++
				if *counter >= 2 {
					t.Fatalf("URL: %s was found %d times", Url, *counter)
				}
				recursivelySearchPages(t, v, Url, counter, depths) // search child page
			}
		}
	}
}
