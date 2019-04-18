package handlers

import (
	"clamber/search"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	depth, err := strconv.Atoi(vars["depth"])
	//allowExternalLinks, err := strconv.ParseBool(vars["allow_external_links"])
	if err != nil {
		panic(err)
	}
	_, err = url.Parse(vars["url"])
	if err != nil {
		log.Fatal(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	query := search.Query{Url: vars["url"], Depth: depth}
	searcher := search.Search{Query: query}
	searcher.Initiate()

	if searcher.Results == nil {
		w.WriteHeader(http.StatusNotFound)
	}
	json.NewEncoder(w).Encode(searcher)

}
