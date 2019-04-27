package api

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/stevenayers/clamber/service"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	// Query contains queried URL, depth and the resulting page data
	Query struct {
		Url        string        `json:"url"`
		Depth      int           `json:"depth"`
		StatusCode int           `json:"statusCode"`
		Results    *service.Page `json:"results"`
	}
)

var ApiCrawler service.Crawler

// SearchHandler function handles /search endpoint. Initiates a database connection, tries to find the url in the database with the
// required depth, and if it doesn't exist, initiate a crawl.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	appConfig, err := service.InitConfig(*AppFlags.ConfigFile)
	logger := InitJsonLogger(log.NewSyncWriter(os.Stdout), appConfig.General.LogLevel)
	statusCode := http.StatusOK
	if err != nil {
		statusCode = http.StatusServiceUnavailable
		w.WriteHeader(statusCode)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	depth, err := strconv.Atoi(vars["depth"])
	if err != nil {
		statusCode = http.StatusBadRequest
		w.WriteHeader(statusCode)
		return
	}
	_, err = url.Parse(vars["url"])
	if err != nil {
		statusCode = http.StatusBadRequest
		w.WriteHeader(statusCode)
		return
	}
	query := Query{Url: vars["url"], Depth: depth}
	store := service.DbStore{}
	store.Connect(appConfig.Database)
	ctx := context.Background()
	txn := store.NewTxn()
	result, err := store.FindNode(&ctx, txn, query.Url, query.Depth)
	if err != nil {
		if !strings.Contains(err.Error(), "Depth does not match dgraph result.") {
			statusCode = http.StatusServiceUnavailable
			w.WriteHeader(statusCode)
			return
		}
	}
	if result == nil {
		start := time.Now()
		ApiCrawler = service.Crawler{
			DbWaitGroup:    sync.WaitGroup{},
			AlreadyCrawled: make(map[string]struct{}),
			Db:             store,
			Config:         appConfig,
			Logger:         logger,
		}
		_ = level.Info(logger).Log(
			"uid", r.Header.Get("Clamber-Request-ID"),
			"url", query.Url,
			"depth", query.Depth,
			"msg", "initiating search",
		)
		result = &service.Page{Url: query.Url}
		ApiCrawler.Crawl(result, query.Depth)

		go func() {
			ApiCrawler.DbWaitGroup.Wait()
			_ = level.Info(logger).Log(
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
		query.Results = nil
		statusCode = http.StatusNotFound
		w.WriteHeader(statusCode)
	}
	query.StatusCode = statusCode
	json.NewEncoder(w).Encode(query)
}
