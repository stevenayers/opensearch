package route

import (
	"github.com/gorilla/mux"
	"github.com/stevenayers/opensearch/pkg/logging"
	"net/http"
)

type (
	// Route contains all route data
	Route struct {
		Name        string
		Method      string
		Pattern     string
		HandlerFunc http.HandlerFunc
		Params      []string
	}
)

// NewRouter function initiates a mux router object with custom HTTP Response logger.
func NewRouter(definedRoutes []Route) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range definedRoutes {
		handler := logging.HttpResponseLogger(route.HandlerFunc)
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler).Queries(route.Params...)
	}
	//router.Handle("/metrics", promhttp.Handler())

	return router
}
