package handlers

import (
	"clamber/database"
	"clamber/search"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strconv"
)

func Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	depth, err := strconv.Atoi(vars["depth"])
	allowExternalLinks, err := strconv.ParseBool(vars["allow_external_links"])
	if err != nil {
		panic(err)
	}
	parsedUrl, err := url.Parse(vars["url"])
	query := database.Query{Url: parsedUrl, Depth: depth, AllowExternalLinks: allowExternalLinks}
	searcher := search.Search{Query: query}
	searcher.Initiate()

	if len(searcher.Results) == 0 {
		w.WriteHeader(http.StatusNotFound)
	}
	json.NewEncoder(w).Encode(searcher)

}
