package routes

import (
	"github.com/gorilla/mux"
	"go-clamber/api/handlers"
	"go-clamber/api/logger"
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
		handlers.Search,
		[]string{
			"url", "{url}",
			"depth", "{depth}",
			"allow_external_links", "{allow_external_links}",
		},
	},
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range DefinedRoutes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = logger.Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler).Queries(route.Queries...)
	}

	return router
}
