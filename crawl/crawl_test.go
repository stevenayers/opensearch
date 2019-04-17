package crawl_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-clamber/crawl"
	"go-clamber/database"
	"go-clamber/page"
	"testing"
	"time"
)

type StoreSuite struct {
	suite.Suite
	store database.DbStore
}

func (s *StoreSuite) SetupSuite() {
	s.store = database.DbStore{}
	database.InitStore(&s.store)
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
	err := s.store.Connection.Close()
	if err != nil {
		fmt.Print(err)
	}
}

func TestCrawlSuite(t *testing.T) {
	s := new(StoreSuite)
	suite.Run(t, s)
}

type CrawlTest struct {
	Url   string
	Depth int
}

var CrawlTests = []CrawlTest{
	{"https://golang.org", 1},
	{"https://golang.org", 3},
	{"http://example.edu", 1},
	{"http://example.edu", 3},
	{"https://google.com", 1},
	{"https://google.com", 3},
}

var PageReturnTests = []string{
	"https://golang.org",
	"http://example.com",
	"https://google.com",
}

func (s *StoreSuite) TestAlreadyCrawled() {
	for _, test := range CrawlTests {
		fmt.Printf("- %s - %d \n", test.Url, test.Depth)
		crawler := crawl.Crawler{AlreadyCrawled: make(map[string]struct{})}
		rootPage := page.Page{Url: test.Url, Depth: test.Depth, Timestamp: time.Now().Unix()}
		crawler.Crawl(&rootPage)
		for Url := range crawler.AlreadyCrawled { // Iterate through crawled AlreadyCrawled and recursively search for each one
			var countedDepths []int
			crawledCounter := 0
			recursivelySearchPages(s.T(), &rootPage, Url, &crawledCounter, &countedDepths)()
		}
	}
}

func (s *StoreSuite) TestAllPagesReturned() {
	for _, testUrl := range PageReturnTests {
		crawler := crawl.Crawler{AlreadyCrawled: make(map[string]struct{})}
		rootPage := page.Page{Url: testUrl, Depth: 1, Timestamp: time.Now().Unix()}
		Urls, _ := rootPage.FetchChildPages()
		crawler.Crawl(&rootPage)
		assert.Equal(s.T(), len(Urls), len(rootPage.Children), "page.Children and fetch Urls length expected to match.")
	}
}

// Helper Function for TestAlreadyCrawled. Cleanest way to implement recursive map checking into the test.
func recursivelySearchPages(t *testing.T, p *page.Page, Url string, counter *int, depths *[]int) func() {
	return func() {
		for _, v := range p.Children {
			if v.Children != nil && v.Url == Url { // Check if page has links
				*depths = append(*depths, v.Depth) // Log the depth it was counted (useful when inspecting data structure)
				*counter++
				assert.NotEqualf(t, 2, *counter, "%s: Url was counted more than once.", v.Url)
				recursivelySearchPages(t, v, Url, counter, depths) // search child page
			}
		}
	}
}
