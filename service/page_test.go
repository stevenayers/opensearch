package service_test

import (
	"github.com/stevenayers/clamber/service"
	"github.com/stretchr/testify/assert"
	"net/http"
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
)

var FetchUrlTests = []FetchUrlTest{
	{"http://example.edu", false},
	{"HTTP://EXAMPLE.EDU", false},
	{"https://www.exmaple.com", true},
	{"ftp://example.edu/file.txt", true},
	{"//cdn.example.edu/lib.js", true},
	{"/myfolder/test.txt", true},
	{"test", true},
}

var RelativeUrlTests = []RelativeUrlTest{
	{"http://example.edu", false},
	{"HTTP://EXAMPLE.EDU", false},
	{"https://www.exmaple.com", false},
	{"ftp://example.edu/file.txt", false},
	{"//cdn.example.edu/lib.js", false},
	{"/myfolder/test.txt", true},
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
		thisPage := service.Page{Url: test.Url}
		_, err := http.Get(thisPage.Url)
		if err != nil {
			service.APILogger.LogDebug("context", "failed to get URL", "url", thisPage.Url, "msg", err.Error())
		}
		assert.Equal(s.T(), test.httpError, err != nil)
	}
}

// simulate a page with a given number of links, and check that the number of links
// on the page reflect the number of links returned.
// Another test case is checking correct errors from parseHtml

func (s *StoreSuite) TestParseHtml() {
	testUrl := "https://golang.org"
	p := service.Page{Url: testUrl}
	_, err := p.FetchChildPages(nil)
	assert.Equal(s.T(), true, err != nil, "nil")
	p = service.Page{Url: testUrl}
	resp, err := http.Get(testUrl)
	if err != nil {
		s.T().Fatal(err)
	}
	_, err = p.FetchChildPages(resp)
	assert.Equal(s.T(), false, err != nil, testUrl)

}

func (s *StoreSuite) TestIsRelativeUrl() {
	for _, test := range RelativeUrlTests {
		p := &service.Page{}
		assert.Equal(s.T(), test.IsRelative, p.IsRelativeUrl(test.Url))
	}
}

func (s *StoreSuite) TestParseRelativeUrl() {
	rootUrl := "http://example.edu"
	for _, test := range ParseUrlTests {
		p := &service.Page{Url: rootUrl}
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
		p := &service.Page{Url: rootUrl}
		_, err := p.ParseRelativeUrl(test.Url)
		assert.Equal(s.T(), true, err != nil)
	}
}

func (s *StoreSuite) TestParseRelativeError() {
	rootUrl := "http://example.edu"
	p := &service.Page{Url: rootUrl}
	_, err := p.ParseRelativeUrl("@£$%@@%")
	assert.Equal(s.T(), true, err != nil)

}
