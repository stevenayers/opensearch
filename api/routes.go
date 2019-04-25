package api

import (
	"github.com/gorilla/mux"
	"github.com/stevenayers/clamber/service"
	"net/http"
)

type (
	Routes []Route

	Route struct {
		Name        string
		Method      string
		Pattern     string
		HandlerFunc http.HandlerFunc
		Queries     []string
	}
)

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

// Initiates a mux router object with custom HTTP Response logger.
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range DefinedRoutes {
		handler := service.HttpResponseLogger(route.HandlerFunc)
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler).Queries(route.Queries...)
	}

	return router
}
