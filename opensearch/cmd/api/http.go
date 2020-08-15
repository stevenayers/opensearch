package main

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log/level"
	"github.com/stevenayers/opensearch/pkg/config"
	"github.com/stevenayers/opensearch/pkg/database/relationship"
	"github.com/stevenayers/opensearch/pkg/logging"
	"github.com/stevenayers/opensearch/pkg/page"
	"github.com/stevenayers/opensearch/pkg/query"
	"github.com/stevenayers/opensearch/pkg/queue"
	"github.com/stevenayers/opensearch/pkg/route"
	"net/http"
	"strings"
)

// Routes contains defined routes data
var Routes = []route.Route{
	{
		Name:        "Initiate",
		Method:      "GET",
		Pattern:     "/search",
		HandlerFunc: SearchHandler,
		Params: []string{
			"url", "{url}",
			"depth", "{depth}",
		},
	},
}

// SearchHandler function handles /search endpoint. Initiates a database connection, tries to find the url in the database with the
// required depth, and if it doesn't exist, initiate a crawl.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "cloudformation/json; charset=UTF-8")
	requestUid := r.Header.Get("OpenSearch-Request-ID")
	statusCode := http.StatusOK
	q, err := query.New(r)
	if err != nil {
		statusCode = http.StatusBadRequest
		w.WriteHeader(statusCode)
		_ = level.Error(logging.Logger).Log("context", "requestUid", requestUid, "msg", err.Error())
		return
	}
	store := relationship.Store{}
	store.Connect()
	var result *page.Page
	if q.Depth >= 0 {
		ctx := context.Background()
		result, err = store.FindNode(&ctx, q.Url, q.Depth)
		if err != nil {
			if !strings.Contains(err.Error(), "Depth does not match dgraph result.") {
				statusCode = http.StatusServiceUnavailable
				w.WriteHeader(statusCode)
				_ = level.Error(logging.Logger).Log("context", "requestUid", requestUid, "msg", err.Error())
				return
			}
		}
	}
	if result == nil {
		qu := queue.NewQueue()
		startPage := &page.Page{
			Url:      q.Url,
			Depth:    q.DisplayDepth,
			StartUrl: q.Url,
		}
		qu.Publish(startPage)
		if config.AppConfig.Api.WaitCrawl {
			result, err = q.PollForFinishedCrawl(store)
			if err != nil {
				statusCode = http.StatusServiceUnavailable
				w.WriteHeader(statusCode)
				_ = level.Error(logging.Logger).Log("context", "polling for finished crawl", "requestUid", requestUid, "msg", err.Error())
				return
			}
		} else {
			go func() {
				result, err = q.PollForFinishedCrawl(store)
				if err != nil {
					_ = level.Error(logging.Logger).Log("context", "polling for finished crawl", "requestUid", requestUid, "msg", err.Error())
					return
				}
			}()
		}
	}
	q.Results = result
	if q.Results.Links == nil {
		q.Results = nil
		statusCode = http.StatusNotFound
		w.WriteHeader(statusCode)
	}
	q.StatusCode = statusCode
	json.NewEncoder(w).Encode(q)
}
