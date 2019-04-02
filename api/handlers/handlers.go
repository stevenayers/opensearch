package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go-clamber/api/search"
	"net/http"
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

	query := search.Query{Url: vars["url"], Depth: depth, AllowExternalLinks: allowExternalLinks}
	searcher := search.Search{Query: query}
	searcher.Initiate()

	switch {
	case len(searcher.Results) == 0:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(searcher)

}
