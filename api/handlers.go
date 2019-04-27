package api

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/stevenayers/clamber/service"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	// Query contains queried URL, depth and the resulting page data
	Query struct {
		Url     string        `json:"url"`
		Depth   int           `json:"depth"`
		Results *service.Page `json:"results"`
	}
)

var ApiCrawler service.Crawler

// SearchHandler function handles /search endpoint. Initiates a database connection, tries to find the url in the database with the
// required depth, and if it doesn't exist, initiate a crawl.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	depth, err := strconv.Atoi(vars["depth"])
	//allowExternalLinks, err := strconv.ParseBool(vars["allow_external_links"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = url.Parse(vars["url"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	query := Query{Url: vars["url"], Depth: depth}
	store := service.DbStore{}
	service.Connect(&store, service.AppConfig.Database)
	ctx := context.Background()
	txn := store.NewTxn()
	result, err := service.DB.FindNode(&ctx, txn, query.Url, query.Depth)
	if err != nil {
		if !strings.Contains(err.Error(), "Depth does not match dgraph result.") {
			return
		}
	}
	if result == nil {
		start := time.Now()
		_ = service.APILogger.LogDebug(
			"uid", r.Header.Get("Clamber-Request-ID"),
			"url", query.Url,
			"depth", query.Depth,
			"msg", "initiating search",
		)
		ApiCrawler = service.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		result = &service.Page{Url: query.Url}
		ApiCrawler.Crawl(result, query.Depth)
		go func() {
			ApiCrawler.DbWaitGroup.Wait()
			_ = service.APILogger.LogDebug(
				"uid", r.Header.Get("Clamber-Request-ID"),
				"url", query.Url,
				"depth", query.Depth,
				"duration", time.Since(start),
				"msg", "finished writing result to dgraph",
			)
		}()
	}
	query.Results = result
	if query.Results.Links == nil {
		w.WriteHeader(http.StatusNotFound)
	}
	json.NewEncoder(w).Encode(query)
}
