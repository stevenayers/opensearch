package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go-clamber/map-api/models"
	"go-clamber/map-api/search"
	"net/http"
	"strconv"
)

func Search(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	depth, err := strconv.Atoi(vars["depth"])
	allowExternalLinks, err := strconv.ParseBool(vars["allow_external_links"])
	if err != nil {
		panic(err)
	}
	query := models.Query{Url: vars["url"], Depth: depth, AllowExternalLinks: allowExternalLinks}

	searcher := search.Searcher{Query: query}
	searcher.Search()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(searcher.Results)
}
