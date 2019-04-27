package api_test

import (
	"github.com/gorilla/mux"
	"github.com/stevenayers/clamber/api"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
)

type QueryParamsTest struct {
	Url                string
	Depth              int
	AllowExternalLinks bool
	DisplayDepth       int
}

var QueryParamsTests = []QueryParamsTest{
	{"https://golang.org", 1, false, 10},
	{"https://golang.org", 2, false, 10},
}

func (s *StoreSuite) TestSearchHandler() {
	for _, test := range QueryParamsTests {
		req, _ := http.NewRequest("GET", "/search", nil)
		q := req.URL.Query()
		q.Add("url", test.Url)
		q.Add("depth", strconv.Itoa(test.Depth))
		q.Add("display_depth", strconv.Itoa(test.DisplayDepth))
		req.URL.RawQuery = q.Encode()
		response := httptest.NewRecorder()
		router := api.NewRouter()
		router.ServeHTTP(response, req)
		assert.Equal(s.T(), 200, response.Code, "StatusOK response is expected")
		api.ApiCrawler.DbWaitGroup.Wait()
	}
}

func (s *StoreSuite) TestSearchHandlerConfigError() {
	*api.AppFlags.ConfigFile = "../test/incorrectpath.toml"
	for _, test := range QueryParamsTests {
		req, _ := http.NewRequest("GET", "/search", nil)
		q := req.URL.Query()
		q.Add("url", test.Url)
		q.Add("depth", strconv.Itoa(test.Depth))
		q.Add("display_depth", strconv.Itoa(test.DisplayDepth))
		req.URL.RawQuery = q.Encode()
		response := httptest.NewRecorder()
		router := api.NewRouter()
		router.ServeHTTP(response, req)
		assert.Equal(s.T(), 503, response.Code, "Service Unavailable response is expected")
		api.ApiCrawler.DbWaitGroup.Wait()
	}
}

func (s *StoreSuite) TestSearchHandlerBadUrl() {
	req, _ := http.NewRequest("GET", "/search", nil)
	q := req.URL.Query()
	q.Add("url", "http://[fe80::%31%25en0]/")
	q.Add("depth", strconv.Itoa(1))
	req.URL.RawQuery = q.Encode()
	response := httptest.NewRecorder()
	router := api.NewRouter()
	router.ServeHTTP(response, req)
	assert.Equal(s.T(), 400, response.Code, "BadRequest response is expected")
	api.ApiCrawler.DbWaitGroup.Wait()
}

func (s *StoreSuite) TestSearchHandlerBadDepth() {
	req, _ := http.NewRequest("GET", "/search", nil)
	q := req.URL.Query()
	q.Add("url", "https://golang.org")
	q.Add("depth", "stringnotint")
	req.URL.RawQuery = q.Encode()
	response := httptest.NewRecorder()
	router := api.NewRouter()
	router.ServeHTTP(response, req)
	assert.Equal(s.T(), 400, response.Code, "BadRequest response is expected")
	api.ApiCrawler.DbWaitGroup.Wait()
}

func (s *StoreSuite) TestSearchHandlerNotFound() {
	req, _ := http.NewRequest("GET", "/search", nil)
	q := req.URL.Query()
	q.Add("url", "http://blsdadadadadsa.uk")
	q.Add("depth", strconv.Itoa(2))
	q.Add("display_depth", strconv.Itoa(2))
	req.URL.RawQuery = q.Encode()
	response := httptest.NewRecorder()
	router := api.NewRouter()
	router.ServeHTTP(response, req)
	assert.Equal(s.T(), 404, response.Code, "BadRequest response is expected")
	api.ApiCrawler.DbWaitGroup.Wait()
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
		rw := api.NewRichResponseWriter(w)
		handler.ServeHTTP(rw, r)
	})
}
