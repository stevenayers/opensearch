package api

import (
	"github.com/stevenayers/clamber/service"
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
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
		Url          string        `json:"url"`
		Depth        int           `json:"depth"`
		DisplayDepth int           `json:"display_depth"`
		StatusCode   int           `json:"statusCode"`
		Results      *service.Page `json:"results"`
	}
)

// SearchHandler function handles /search endpoint. Initiates a database connection, tries to find the url in the database with the
// required depth, and if it doesn't exist, initiate a crawl.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	requestUid := r.Header.Get("Clamber-Request-ID")
	appConfig, err := service.InitConfig(*AppFlags.ConfigFile)
	logger := InitJsonLogger(log.NewSyncWriter(os.Stdout), appConfig.General.LogLevel)
	statusCode := http.StatusOK
	if err != nil {
		statusCode = http.StatusServiceUnavailable
		w.WriteHeader(statusCode)
		return
	}
	query, err := parseQuery(r)
	if err != nil {
		statusCode = http.StatusBadRequest
		w.WriteHeader(statusCode)
		return
	}
	store := service.DbStore{}
	store.Connect(appConfig.Database)
	var result *service.Page
	if query.Depth >= 0 {
		ctx := context.Background()
		txn := store.NewTxn()
		result, err = store.FindNode(&ctx, txn, query.Url, query.Depth)
		if err != nil {
			if !strings.Contains(err.Error(), "Depth does not match dgraph result.") {
				statusCode = http.StatusServiceUnavailable
				w.WriteHeader(statusCode)
				return
			}
		}
	}
	if result == nil {
		start := time.Now()
		crawler := &service.Crawler{
			DbWaitGroup:    sync.WaitGroup{},
			AlreadyCrawled: make(map[string]struct{}),
			Db:             store,
			StartUrl:       query.Url,
			Config:         appConfig,
			Logger:         logger,
			CrawlUid:       uuid.New(),
		}
		_ = level.Info(logger).Log(
			"requestUid", requestUid,
			"url", query.Url,
			"depth", query.Depth,
			"context", "Crawl",
			"crawlUid", crawler.CrawlUid,
			"backgroundDepth", crawler.BackgroundCrawlDepth,
			"msg", "initiating search",
		)
		result = &service.Page{Url: query.Url, Depth: query.DisplayDepth}
		crawler.Crawl(result)

		if appConfig.General.WaitCrawl {
			crawlFinished(crawler, query, start, requestUid)
		} else {
			go crawlFinished(crawler, query, start, requestUid)
		}
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

func parseQuery(r *http.Request) (query Query, err error) {
	var start *url.URL
	var depth int
	var displayDepth int
	start, err = url.Parse(r.URL.Query().Get("url"))
	if err != nil {
		return
	}
	depth, err = strconv.Atoi(r.URL.Query().Get("depth"))
	if err != nil {
		return
	}
	if dDepth := r.URL.Query().Get("display_depth"); dDepth != "" {
		displayDepth, err = strconv.Atoi(dDepth)
		if err != nil {
			return
		} else if displayDepth == 0 {
			displayDepth = 10
		}
	} else {
		displayDepth = 10
	}
	if depth != -1 && displayDepth > depth {
		displayDepth = depth
	}
	query = Query{Url: start.String(), Depth: depth, DisplayDepth: displayDepth}
	return
}

func crawlFinished(crawler *service.Crawler, query Query, start time.Time, requestUid string) {
	crawler.DbWaitGroup.Wait()
	_ = level.Info(crawler.Logger).Log(
		"requestUid", requestUid,
		"url", query.Url,
		"depth", query.Depth,
		"context", "Crawl",
		"duration", time.Since(start),
		"crawlUid", crawler.CrawlUid,
		"backgroundDepth", crawler.BackgroundCrawlDepth,
		"msg", "finished writing displayed result to dgraph",
	)
}
