package service_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stevenayers/clamber/service"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync"
	"testing"
	"time"
)

type (
	CrawlTest struct {
		Url   string
		Depth int
	}
)

var (
	CrawlTests = []CrawlTest{
		{"https://golang.org", 2},
		{"https://youtube.com", 1},
		{"https://google.com", 1},
	}

	PageReturnTests = []string{
		"https://golang.org",
		"http://example.edu",
		"https://google.com",
	}

	CrawlBadCreateTests = []string{
		"https://golang.org",
		"http://example.edu",
		"https://google.com",
	}

	CrawlBadUrlTests = []string{
		"https://fake.link.local",
		"http://another.fake.link.local",
	}
)

func (s *StoreSuite) TestAlreadyCrawled() {
	for _, test := range CrawlTests {
		crawler := service.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		rootPage := service.Page{Url: test.Url, Timestamp: time.Now().Unix()}
		crawler.Crawl(&rootPage, test.Depth)
		crawler.DbWaitGroup.Wait()
		for Url := range crawler.AlreadyCrawled { // Iterate through crawled AlreadyCrawled and recursively search for each one
			var countedDepths []int
			crawledCounter := 0
			recursivelySearchPages(s.T(), &rootPage, test.Depth, Url, &crawledCounter, &countedDepths)()
		}
	}
}

func (s *StoreSuite) TestCrawlBadUrl() {
	for _, testUrl := range CrawlBadUrlTests {
		buf := new(bytes.Buffer)
		service.APILogger = service.ApiLogger{}
		var logOutput LogOutput
		service.APILogger.InitJsonLogger(buf, "debug")
		crawler := service.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		rootPage := service.Page{Url: testUrl, Timestamp: time.Now().Unix()}
		crawler.Crawl(&rootPage, 0)
		err := json.Unmarshal(buf.Bytes(), &logOutput)
		if err != nil {
			s.T().Fatal(err.Error())
		}
		assert.Equal(s.T(), "error", logOutput.Level)
		assert.Equal(s.T(), "failed to get URL", logOutput.Context)
	}
}

func (s *StoreSuite) TestCrawlBadCreate() {
	for _, testUrl := range CrawlBadCreateTests {
		buf := new(bytes.Buffer)
		service.APILogger = service.ApiLogger{}
		var logOutput LogOutput
		service.APILogger.InitJsonLogger(buf, "debug")
		crawler := service.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		rootPage := service.Page{Url: testUrl, Timestamp: time.Now().Unix()}
		_ = s.store.DeleteAll()
		crawler.Crawl(&rootPage, 0)
		crawler.DbWaitGroup.Wait()
		err := json.Unmarshal(buf.Bytes(), &logOutput)
		if err != nil {
			s.T().Fatal(err)
		}
		fmt.Print(logOutput.Context)
		assert.Equal(s.T(), "error", logOutput.Level)

	}
}

func (s *StoreSuite) TestAllPagesReturned() {
	for _, testUrl := range PageReturnTests {
		crawler := service.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		rootPage := service.Page{Url: testUrl, Timestamp: time.Now().Unix()}
		resp, err := http.Get(rootPage.Url)
		var Urls []*service.Page
		if err != nil {
			_ = service.APILogger.LogDebug("context", "failed to get URL", "url", rootPage.Url, "msg", err.Error())
			Urls, _ = rootPage.FetchChildPages(resp)
			crawler.Crawl(&rootPage, 1)
			crawler.DbWaitGroup.Wait()
			assert.Equal(s.T(), len(Urls), len(rootPage.Links), "page.Links and fetch Urls length expected to match.")
		}
	}
}

func recursivelySearchPages(t *testing.T, p *service.Page, depth int, Url string, counter *int, depths *[]int) func() {
	return func() {
		for _, v := range p.Links {
			if v.Links != nil && v.Url == Url { // Check if page has links
				childDepth := depth - 1
				*depths = append(*depths, childDepth) // Log the depth it was counted (useful when inspecting data structure)
				*counter++
				assert.NotEqualf(t, 2, *counter, "%s: Url was counted more than once.", v.Url)
				recursivelySearchPages(t, v, childDepth, Url, counter, depths) // search child page
			}
		}
	}
}
