package main_test

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/stevenayers/clamber/clamber/cmd/api"
	"github.com/stevenayers/clamber/pkg/config"
	"github.com/stevenayers/clamber/pkg/crawl"
	"github.com/stevenayers/clamber/pkg/database/relationship"
	"github.com/stevenayers/clamber/pkg/logging"
	"github.com/stevenayers/clamber/pkg/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var QueryParamsTests = []QueryParamsTest{
	{"https://golang.org", 1, false},
	{"https://golang.org", 2, false},
}

type (
	QueryParamsTest struct {
		Url                string
		Depth              int
		AllowExternalLinks bool
	}

	StoreSuite struct {
		suite.Suite
		store   relationship.Store
		crawler crawl.Crawler
	}
)

func (s *StoreSuite) SetupSuite() {
	var err error
	main.InitFlags(&main.AppFlags)
	*main.AppFlags.ConfigFile = "/Users/steven/git/clamber/configs/config.toml"
	config.InitConfig(*main.AppFlags.ConfigFile)
	if err != nil {
		s.T().Fatal(err)
	}
	logging.InitJsonLogger(log.NewSyncWriter(os.Stdout), config.AppConfig.Api.LogLevel, "test")
	s.store = relationship.Store{}
	s.store.Connect()

}

func (s *StoreSuite) SetupTest() {
	var err error
	*main.AppFlags.ConfigFile = "/Users/steven/git/clamber/configs/config.toml"
	err = config.InitConfig(*main.AppFlags.ConfigFile)
	if err != nil {
		s.T().Fatal(err)
	}
	s.store.Connect()
	err = s.store.DeleteAll()
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

func TestSuite(t *testing.T) {
	s := new(StoreSuite)
	suite.Run(t, s)
}

func (s *StoreSuite) TestSearchHandler() {
	for _, test := range QueryParamsTests {
		req, _ := http.NewRequest("GET", "/search", nil)
		q := req.URL.Query()
		q.Add("url", test.Url)
		q.Add("depth", strconv.Itoa(test.Depth))
		req.URL.RawQuery = q.Encode()
		response := httptest.NewRecorder()
		router := route.NewRouter(main.Routes)
		router.ServeHTTP(response, req)
		assert.Equal(s.T(), 200, response.Code, "StatusOK response is expected")
	}
}

func (s *StoreSuite) TestSearchHandlerConfigError() {
	*main.AppFlags.ConfigFile = "../test/incorrectpath.toml"
	for _, test := range QueryParamsTests {
		req, _ := http.NewRequest("GET", "/search", nil)
		q := req.URL.Query()
		q.Add("url", test.Url)
		q.Add("depth", strconv.Itoa(test.Depth))
		req.URL.RawQuery = q.Encode()
		response := httptest.NewRecorder()
		router := route.NewRouter(main.Routes)
		router.ServeHTTP(response, req)
		assert.Equal(s.T(), 503, response.Code, "Service Unavailable response is expected")
	}
}

func (s *StoreSuite) TestSearchHandlerBadUrl() {
	req, _ := http.NewRequest("GET", "/search", nil)
	q := req.URL.Query()
	q.Add("url", "http://[fe80::%31%25en0]/")
	q.Add("depth", strconv.Itoa(1))
	req.URL.RawQuery = q.Encode()
	response := httptest.NewRecorder()
	router := route.NewRouter(main.Routes)
	router.ServeHTTP(response, req)
	assert.Equal(s.T(), 400, response.Code, "BadRequest response is expected")
	//app.ApiCrawler.DbWaitGroup.Wait()
}

func (s *StoreSuite) TestSearchHandlerBadDepth() {
	req, _ := http.NewRequest("GET", "/search", nil)
	q := req.URL.Query()
	q.Add("url", "https://golang.org")
	q.Add("depth", "stringnotint")
	req.URL.RawQuery = q.Encode()
	response := httptest.NewRecorder()
	router := route.NewRouter(main.Routes)
	router.ServeHTTP(response, req)
	assert.Equal(s.T(), 400, response.Code, "BadRequest response is expected")
	//app.ApiCrawler.DbWaitGroup.Wait()
}

func (s *StoreSuite) TestSearchHandlerNotFound() {
	req, _ := http.NewRequest("GET", "/search", nil)
	q := req.URL.Query()
	q.Add("url", "http://blsdadadadadsa.uk")
	q.Add("depth", strconv.Itoa(2))
	req.URL.RawQuery = q.Encode()
	response := httptest.NewRecorder()
	router := route.NewRouter(main.Routes)
	router.ServeHTTP(response, req)
	assert.Equal(s.T(), 404, response.Code, "BadRequest response is expected")
	//app.ApiCrawler.DbWaitGroup.Wait()
}

func (s *StoreSuite) TestWriteHeader() {
	router := mux.NewRouter().StrictSlash(true)
	router.
		Methods("GET").
		Path("/test").
		Name("test").
		Handler(testHandlerFunc(http.HandlerFunc(testHandler)))
	req, _ := http.NewRequest("GET", "/test", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	result := response.Result()
	assert.Equal(s.T(), http.StatusCreated, result.StatusCode)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)

}

func testHandlerFunc(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := logging.NewRichResponseWriter(w)
		handler.ServeHTTP(rw, r)
	})
}
