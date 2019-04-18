package search

import (
	"clamber/crawl"
	"clamber/database"
	"clamber/page"
	"context"
	"fmt"
	"log"
)

type (
	Search struct {
		Query   Query  `json:"query"`
		Results []byte `json:"results"`
	}

	Queries []Query

	Query struct {
		Url                string `json:"url"`
		Depth              int    `json:"depth"`
		AllowExternalLinks bool   `json:"allow_external_links"`
	}
)

func (search Search) Initiate() {
	store := database.DbStore{}
	database.InitStore(&database.DbStore{})
	ctx := context.Background()
	txn := store.NewTxn()
	result, err := database.DB.FindNode(&ctx, txn, search.Query.Url, search.Query.Depth)
	if err != nil {
		fmt.Print(err)
	}
	if result == nil {
		log.Print("Could not find node with required depth, initiating search...")
		crawler := crawl.Crawler{AlreadyCrawled: make(map[string]struct{})}
		result = &page.Page{Url: search.Query.Url}
		crawler.Crawl(result, search.Query.Depth)
	}
	json, err := database.SerializePage(result)
	search.Results = json
}
