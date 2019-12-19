package crawl_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/stevenayers/clamber/pkg/config"
	"github.com/stevenayers/clamber/pkg/crawl"
	"github.com/stevenayers/clamber/pkg/database/relationship"
	"github.com/stevenayers/clamber/pkg/logging"
	"github.com/stevenayers/clamber/pkg/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type (
	CrawlTest struct {
		Url   string
		Depth int
	}
	LogOutput struct {
		Level   string `json:"level,omitempty"`
		Node    string `json:"node,omitempty"`
		Service string `json:"app,omitempty"`
		Msg     string `json:"msg,omitempty"`
		Context string `json:"context,omitempty"`
	}
)

var (
	CrawlTests = []CrawlTest{
		//{"https://golang.org", 2},
		//{"https://youtube.com", 1},
		//{"https://google.com", 1},
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

	NodeTests = []NodeTest{
		{"https://golang.org", 1},
		{"https://google.com", 1},
		{"https://youtube.com", 1},
	}

	NodeDepthTests = []NodeTest{
		{"https://golang.org", 1},
		{"https://golang.org", 2},
		{"https://google.com", 2},
	}
)

type (
	NodeTest struct {
		Url   string
		Depth int
	}

	StoreSuite struct {
		suite.Suite
		store   relationship.Store
		crawler crawl.Crawler
	}
)

func (s *StoreSuite) SetupSuite() {
	var err error
	configFile := "/Users/steven/git/clamber/configs/config.toml"
	err = config.InitConfig(configFile)
	if err != nil {
		s.T().Fatal(err)
	}

	logging.InitJsonLogger(log.NewSyncWriter(os.Stdout), config.AppConfig.Service.LogLevel, "test")
	s.store = relationship.Store{}
	s.store.Connect()
}

func (s *StoreSuite) SetupTest() {
	var err error
	configFile := "/Users/steven/git/clamber/configs/config.toml"
	err = config.InitConfig(configFile)
	if err != nil {
		s.T().Fatal(err)
	}
	s.store.Connect()
	if !strings.Contains(s.T().Name(), "TestLog") && !strings.Contains(s.T().Name(), "TestConnect") {
		err := s.store.DeleteAll()
		if err != nil {
			s.T().Fatal(err)
		}
		err = s.store.SetSchema()
		if err != nil {
			s.T().Fatal(err)
		}
	}
}

func (s *StoreSuite) TearDownSuite() {
	for _, conn := range s.store.Connection {
		err := conn.Close()
		if err != nil {
			fmt.Print(err)
		}
	}
}

func TestSuite(t *testing.T) {
	s := new(StoreSuite)
	suite.Run(t, s)
}

func (s *StoreSuite) TestAlreadyCrawled() {
	for _, test := range CrawlTests {
		crawler := crawl.Crawler{
			DbWaitGroup:    sync.WaitGroup{},
			AlreadyCrawled: make(map[string]struct{}),
			Store:          &s.store,
		}
		rootPage := page.Page{Url: test.Url, Timestamp: time.Now().Unix(), Depth: test.Depth}
		crawler.Crawl(&rootPage)
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
		var logOutput LogOutput
		logging.InitJsonLogger(buf, "debug", "test")
		crawler := crawl.Crawler{
			DbWaitGroup:    sync.WaitGroup{},
			AlreadyCrawled: make(map[string]struct{}),
			Store:          &s.store,
		}
		rootPage := page.Page{Url: testUrl, Timestamp: time.Now().Unix(), Depth: 0}
		crawler.Crawl(&rootPage)
		err := json.Unmarshal(buf.Bytes(), &logOutput)
		if err != nil {
			s.T().Fatal(err.Error())
		}
		assert.Equal(s.T(), "error", logOutput.Level)
		assert.Equal(s.T(), "HTTP failure", logOutput.Context)
	}
}

func (s *StoreSuite) TestCrawlBadCreate() {
	for _, testUrl := range CrawlBadCreateTests {
		buf := new(bytes.Buffer)
		var logOutput LogOutput
		logging.InitJsonLogger(buf, "debug", "test")
		crawler := crawl.Crawler{
			DbWaitGroup:    sync.WaitGroup{},
			AlreadyCrawled: make(map[string]struct{}),
			Store:          &s.store,
		}
		rootPage := page.Page{Url: testUrl, Timestamp: time.Now().Unix(), Depth: 0}
		_ = s.store.DeleteAll()
		crawler.Crawl(&rootPage)
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
		crawler := crawl.Crawler{
			DbWaitGroup:    sync.WaitGroup{},
			AlreadyCrawled: make(map[string]struct{}),
			Store:          &s.store,
		}
		rootPage := page.Page{Url: testUrl, Timestamp: time.Now().Unix(), Depth: 1}
		resp, err := http.Get(rootPage.Url)
		var Urls []*page.Page
		if err != nil {
			_ = level.Error(logging.Logger).Log("context", "failed to get URL", "url", rootPage.Url, "msg", err.Error())
			Urls, _ = rootPage.FetchChildPages(resp)
			crawler.Crawl(&rootPage)
			crawler.DbWaitGroup.Wait()
			assert.Equal(s.T(), len(Urls), len(rootPage.Links), "page.Links and fetch Urls length expected to match.")
		}
	}
}

func recursivelySearchPages(t *testing.T, p *page.Page, depth int, Url string, counter *int, depths *[]int) func() {
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

func (s *StoreSuite) TestCreateAndFindNode() {
	for _, test := range NodeDepthTests {
		expectedPage := page.Page{
			Url:       test.Url,
			Timestamp: time.Now().Unix(),
			Depth:     test.Depth,
		}
		crawler := crawl.Crawler{
			DbWaitGroup:    sync.WaitGroup{},
			AlreadyCrawled: make(map[string]struct{}),
			Store:          &s.store,
		}
		crawler.Crawl(&expectedPage)
		crawler.DbWaitGroup.Wait()
		ctx := context.Background()
		resultPage, err := s.store.FindNode(&ctx, test.Url, test.Depth)
		if err != nil {
			s.T().Fatal(err)
		}
		assert.Equal(s.T(), expectedPage.Uid, resultPage.Uid, "Uid should have matched.")
		assert.Equal(s.T(), expectedPage.Url, resultPage.Url, "Url should have matched.")
	}

}

func (s *StoreSuite) TestCreateError() {
	p := page.Page{Url: "https://golang.org"}
	err := s.store.DeleteAll()
	if err != nil {
		s.T().Fatal(err)
	}
	s.crawler.Store = &s.store
	err = s.crawler.Create(&p)
	assert.Equal(s.T(), true, err != nil)
	c := page.Page{Parent: &p}
	err = s.crawler.Create(&c)
	assert.Equal(s.T(), true, err != nil)
}
