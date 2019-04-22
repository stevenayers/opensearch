package crawl_test

import (
	"clamber/crawl"
	"clamber/database"
	"clamber/page"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
	"time"
)

type (
	StoreSuite struct {
		suite.Suite
		store database.DbStore
	}
	CrawlTest struct {
		Url   string
		Depth int
	}
)

var (
	CrawlTests = []CrawlTest{
		{"https://golang.org", 3},
	}

	PageReturnTests = []string{
		"https://golang.org",
		"http://example.com",
		"https://google.com",
	}
)

func (s *StoreSuite) SetupSuite() {
	s.store = database.DbStore{}
	database.Connect(&s.store)
	err := database.DB.DeleteAll()
	if err != nil {
		s.T().Fatal(err)
	}
	err = s.store.SetSchema()
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *StoreSuite) SetupTest() {
	err := database.DB.DeleteAll()
	if err != nil {
		s.T().Fatal(err)
	}
	err = s.store.SetSchema()
	if err != nil {
		s.T().Fatal(err)
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

func TestCrawlSuite(t *testing.T) {
	s := new(StoreSuite)
	suite.Run(t, s)
}

func (s *StoreSuite) TestAlreadyCrawled() {
	for _, test := range CrawlTests {
		crawler := crawl.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		rootPage := page.Page{Url: test.Url, Timestamp: time.Now().Unix()}
		crawler.Crawl(&rootPage, test.Depth)
		for Url := range crawler.AlreadyCrawled { // Iterate through crawled AlreadyCrawled and recursively search for each one
			var countedDepths []int
			crawledCounter := 0
			recursivelySearchPages(s.T(), &rootPage, test.Depth, Url, &crawledCounter, &countedDepths)()
		}
	}
}

func (s *StoreSuite) TestAllPagesReturned() {
	for _, testUrl := range PageReturnTests {
		crawler := crawl.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		rootPage := page.Page{Url: testUrl, Timestamp: time.Now().Unix()}
		Urls, _ := rootPage.FetchChildPages()
		crawler.Crawl(&rootPage, 1)
		assert.Equal(s.T(), len(Urls), len(rootPage.Children), "page.Children and fetch Urls length expected to match.")
	}
}

func recursivelySearchPages(t *testing.T, p *page.Page, depth int, Url string, counter *int, depths *[]int) func() {
	return func() {
		for _, v := range p.Children {
			if v.Children != nil && v.Url == Url { // Check if page has links
				childDepth := depth - 1
				*depths = append(*depths, childDepth) // Log the depth it was counted (useful when inspecting data structure)
				*counter++
				assert.NotEqualf(t, 2, *counter, "%s: Url was counted more than once.", v.Url)
				recursivelySearchPages(t, v, childDepth, Url, counter, depths) // search child page
			}
		}
	}
}
