package api_test

import (
	"github.com/stevenayers/clamber/api"
	"github.com/stevenayers/clamber/service"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
)

type QueryParamsTest struct {
	Url                string
	Depth              int
	AllowExternalLinks bool
}

var QueryParamsTests = []QueryParamsTest{
	{"https://golang.org", 1, false},
	{"https://golang.org", 2, false},
	{"https://golang.org", 3, false},
}

func (s *StoreSuite) TestSearchHandler() {
	service.InitConfig()
	service.APILogger.InitJsonLogger(service.AppConfig.General.LogLevel)
	for _, test := range QueryParamsTests {
		req, _ := http.NewRequest("GET", "/search", nil)
		q := req.URL.Query()
		q.Add("url", test.Url)
		q.Add("depth", strconv.Itoa(test.Depth))
		req.URL.RawQuery = q.Encode()
		response := httptest.NewRecorder()
		router := api.NewRouter()
		router.ServeHTTP(response, req)
		assert.Equal(s.T(), 200, response.Code, "NotFound response is expected")
	}
}
