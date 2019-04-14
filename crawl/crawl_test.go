package crawl_test

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-clamber/crawl"
	"go-clamber/database"
	"go-clamber/page"
	"net/url"
	"testing"
)

type StoreSuite struct {
	suite.Suite
	store *database.DbStore
}

func (s *StoreSuite) SetupSuite() {
	driver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		fmt.Print(err)
	}
	s.store = &database.DbStore{Driver: driver}
	database.InitStore(s.store)
}

func (s *StoreSuite) SetupTest() {
	session, err := s.store.Driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		s.T().Fatal(err)
	}
	_, err = session.Run("MATCH (n) DETACH DELETE n", map[string]interface{}{})
	if err != nil {
		s.T().Fatal(err)
	}
	defer session.Close()
}

func (s *StoreSuite) TearDownSuite() {
	err := s.store.Driver.Close()
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
	{"https://golang.org/", 1},
	{"https://golang.org/", 5},
	{"https://golang.org/", 30},
	{"http://example.edu", 1},
	{"http://example.edu", 5},
	{"http://example.edu", 10},
	{"https://google.com", 1},
	{"https://google.com", 5},
	{"https://google.com", 30},
}

var PageReturnTests = []string{
	"https://golang.org/",
	"http://example.com",
	"https://google.com",
}

func (s *StoreSuite) TestAlreadyCrawled() {
	for _, test := range CrawlTests {
		crawler := crawl.Crawler{AlreadyCrawled: make(map[string]struct{})}
		rootUrl, _ := url.Parse(test.Url)
		rootPage := page.Page{Url: rootUrl, Depth: test.Depth}
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
		rootUrl, _ := url.Parse(testUrl)
		rootPage := page.Page{Url: rootUrl, Depth: 1}
		Urls, _ := rootPage.FetchChildPages()
		crawler.Crawl(&rootPage)
		assert.Equal(s.T(), len(Urls), len(rootPage.Children), "page.Children and fetch Urls length expected to match.")
	}
}

// Helper Function for TestAlreadyCrawled. Cleanest way to implement recursive map checking into the test.
func recursivelySearchPages(t *testing.T, p *page.Page, Url string, counter *int, depths *[]int) func() {
	return func() {
		for _, v := range p.Children {
			if v.Children != nil && v.Url.String() == Url { // Check if page has links
				*depths = append(*depths, v.Depth) // Log the depth it was counted (useful when inspecting data structure)
				*counter++
				assert.NotEqualf(t, 2, *counter, "%s: Url was counted more than once.", v.Url.String())
				recursivelySearchPages(t, v, Url, counter, depths) // search child page
			}
		}
	}
}
