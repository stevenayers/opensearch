package handlers_test

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go-clamber/api/handlers"
	"go-clamber/api/logger"
	"go-clamber/api/routes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type QueryParamsTest struct {
	Url                string
	Depth              int
	AllowExternalLinks bool
}

var QueryParamsTests = []QueryParamsTest{
	{"https://golang.org/", 1, true},
	{"https://golang.org/", 5, false},
	{"https://golang.org/", 10, false},
	{"http://example.com", 1, false},
	{"http://example.com", 5, false},
	{"http://example.com", 10, false},
	{"https://google.com", 1, false},
	{"https://google.com", 5, false},
	{"https://google.com", 10, false},
}

func TestSearchHandler(t *testing.T) {
	for _, test := range QueryParamsTests {
		req, _ := http.NewRequest("GET", "/search", nil)
		q := req.URL.Query()
		q.Add("url", test.Url)
		q.Add("depth", strconv.Itoa(test.Depth))
		q.Add("allow_external_links", strconv.FormatBool(test.AllowExternalLinks))
		req.URL.RawQuery = q.Encode()
		response := httptest.NewRecorder()
		Router().ServeHTTP(response, req)
		assert.Equal(t, 404, response.Code, "OK response is expected")
	}
}

func Router() *mux.Router {
	queries := []string{
		"url", "{url}",
		"depth", "{depth}",
		"allow_external_links", "{allow_external_links}",
	}
	route := routes.Route{
		Name:        "Initiate",
		Method:      "GET",
		Pattern:     "/search",
		HandlerFunc: handlers.Search,
		Queries:     queries,
	}
	router := mux.NewRouter()
	handler := logger.Logger(route.HandlerFunc, route.Name)
	router.
		Methods(route.Method).
		Path(route.Pattern).
		Name(route.Name).
		Handler(handler).Queries(route.Queries...)
	return router
}
