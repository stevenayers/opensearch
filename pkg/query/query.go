package query

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log/level"
	"github.com/stevenayers/opensearch/pkg/database/relationship"
	"github.com/stevenayers/opensearch/pkg/logging"
	"github.com/stevenayers/opensearch/pkg/page"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type (
	// Query contains queried URL, depth and the resulting page data
	Query struct {
		Url          string     `json:"url"`
		Depth        int        `json:"depth"`
		DisplayDepth int        `json:"display_depth"`
		StatusCode   int        `json:"statusCode"`
		Results      *page.Page `json:"results"`
	}
)

func New(r *http.Request) (query Query, err error) {
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

func (query *Query) PollForFinishedCrawl(store relationship.Store) (result *page.Page, err error) {
	ctx := context.Background()
	var prevResult *page.Page
	for {
		var r []byte
		var pr []byte
		_ = level.Info(logging.Logger).Log("msg", "Polling for crawl...")
		result, err = store.FindNode(&ctx, query.Url, query.Depth)
		if err == nil {
			r, err = json.Marshal(result)
			pr, err = json.Marshal(prevResult)
		}
		switch {
		case err != nil && !strings.Contains(err.Error(), "Depth does not match dgraph result."):
			return
		case prevResult == nil || result == nil:
			prevResult = result
			time.Sleep(time.Millisecond * 100)
			continue
		case prevResult != nil && len(pr) != len(r):
			prevResult = result
			time.Sleep(time.Millisecond * 100)
			continue
		default:
			return
		}
	}
}
