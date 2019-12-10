package page_test

import (
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
	"testing"
)

type (
	FetchUrlTest struct {
		Url       string
		httpError bool
	}

	RelativeUrlTest struct {
		Url        string
		IsRelative bool
	}

	ParseUrlTest struct {
		Url         string
		ExpectedUrl string
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

var FetchUrlTests = []FetchUrlTest{
	{"http://example.edu", false},
	{"HTTP://EXAMPLE.EDU", false},
	{"https://www.exmaple.com", true},
	{"ftp://example.edu/file.txt", true},
	{"//cdn.example.edu/lib.js", true},
	{"/myfolder/txt", true},
	{"test", true},
}

var RelativeUrlTests = []RelativeUrlTest{
	{"http://example.edu", false},
	{"HTTP://EXAMPLE.EDU", false},
	{"https://www.exmaple.com", false},
	{"ftp://example.edu/file.txt", false},
	{"//cdn.example.edu/lib.js", false},
	{"/myfolder/txt", true},
	{"test", true},
}

var ParseUrlTests = []ParseUrlTest{
	{"/myfolder/test", "http://example.edu/myfolder/test"},
	{"test", "http://example.edu/test"},
	{"test/", "http://example.edu/test"},
	{"test#jg380gj39v", "http://example.edu/test"},
}

func (s *StoreSuite) TestFetchUrlsHttpError() {
	for _, test := range FetchUrlTests {
		thisPage := page.Page{Url: test.Url}
		_, err := http.Get(thisPage.Url)
		if err != nil {
			_ = level.Error(logging.Logger).Log("context", "failed to get URL", "url", thisPage.Url, "msg", err.Error())
		}
		assert.Equal(s.T(), test.httpError, err != nil)
	}
}

// simulate a page with a given number of links, and check that the number of links
// on the page reflect the number of links returned.
// Another test case is checking correct errors from parseHtml

func (s *StoreSuite) TestParseHtml() {
	testUrl := "https://golang.org"
	p := page.Page{Url: testUrl}
	_, err := p.FetchChildPages(nil)
	assert.Equal(s.T(), true, err != nil, "nil")
	p = page.Page{Url: testUrl}
	resp, err := http.Get(testUrl)
	if err != nil {
		s.T().Fatal(err)
	}
	_, err = p.FetchChildPages(resp)
	assert.Equal(s.T(), false, err != nil, testUrl)

}

func (s *StoreSuite) TestIsRelativeUrl() {
	for _, test := range RelativeUrlTests {
		p := &page.Page{}
		assert.Equal(s.T(), test.IsRelative, p.IsRelativeUrl(test.Url))
	}
}

func (s *StoreSuite) TestParseRelativeUrl() {
	rootUrl := "http://example.edu"
	for _, test := range ParseUrlTests {
		p := &page.Page{Url: rootUrl}
		absoluteUrl, err := p.ParseRelativeUrl(test.Url)
		if err != nil {
			s.T().Fatal(err)
		}
		assert.Equal(s.T(), test.ExpectedUrl, absoluteUrl.String())
	}
}

func (s *StoreSuite) TestParseRelativeRootError() {
	rootUrl := "£$@£%"
	for _, test := range ParseUrlTests {
		p := &page.Page{Url: rootUrl}
		_, err := p.ParseRelativeUrl(test.Url)
		assert.Equal(s.T(), true, err != nil)
	}
}

func (s *StoreSuite) TestParseRelativeError() {
	rootUrl := "http://example.edu"
	p := &page.Page{Url: rootUrl}
	_, err := p.ParseRelativeUrl("@£$%@@%")
	assert.Equal(s.T(), true, err != nil)

}
