package api

import (
	"clamber/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type (
	Query struct {
		Url     string        `json:"url"`
		Depth   int           `json:"depth"`
		Results *service.Page `json:"results"`
	}
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
		w.WriteHeader(http.StatusNotFound)
		log.Fatal(err.Error())
		return
	}
	query := Query{Url: vars["url"], Depth: depth}
	store := service.DbStore{}
	service.Connect(&store)
	ctx := context.Background()
	txn := store.NewTxn()
	result, err := service.DB.FindNode(&ctx, txn, query.Url, query.Depth)
	if err != nil {
		fmt.Print(err)
	}
	var crawler service.Crawler
	if result == nil {
		start := time.Now()
		log.Printf("%s: initiating search (url: %s depth: %d)",
			r.Header.Get("Clamber-Request-ID"),
			query.Url,
			query.Depth,
		)
		crawler = service.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		result = &service.Page{Url: query.Url}
		crawler.Crawl(result, query.Depth)
		go func() {
			crawler.DbWaitGroup.Wait()
			log.Printf("%s: finished writing result to dgraph (duration: %s)",
				r.Header.Get("Clamber-Request-ID"),
				time.Since(start),
			)
		}()
	}
	query.Results = result
	if query.Results == nil {
		w.WriteHeader(http.StatusNotFound)
	}
	json.NewEncoder(w).Encode(query)
}
