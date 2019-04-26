package api

import (
	"github.com/gorilla/mux"
	"net/http"
)

type (
	// Routes is slice of Route
	Routes []Route

	// Route contains all route data
	Route struct {
		Name        string
		Method      string
		Pattern     string
		HandlerFunc http.HandlerFunc
		Queries     []string
	}
)

// DefinedRoutes contains defined routes data
var DefinedRoutes = Routes{
	Route{
		"Initiate",
		"GET",
		"/search",
		SearchHandler,
		[]string{
			"url", "{url}",
			"depth", "{depth}",
		},
	},
}

// NewRouter function initiates a mux router object with custom HTTP Response logger.
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range DefinedRoutes {
		handler := HttpResponseLogger(route.HandlerFunc)
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler).Queries(route.Queries...)
	}

	return router
}
