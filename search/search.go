package search

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"go-clamber/crawl"
	"go-clamber/database"
	"go-clamber/page"
	"net/url"
)

type (
	Search struct {
		Query   Query        `json:"query"`
		Results []*page.Page `json:"results"`
	}

	Queries []Query

	Query struct {
		Url                *url.URL `json:"url"`
		Depth              int      `json:"depth"`
		AllowExternalLinks bool     `json:"allow_external_links"`
	}
)

func (search Search) Initiate() {
	_, err := sql.Open("sqlite3", "../database/testing/pages.sqlite")
	if err != nil {
		fmt.Print(err)
	}
	database.InitStore(&database.DbStore{})
	results, err := database.DB.GetPage(search.Query.Url.String())
	if err != nil {
		fmt.Print(err)
	}
	if len(results) == 0 {
		crawler := crawl.Crawler{AlreadyCrawled: make(map[string]struct{})}
		crawler.Crawl(&page.Page{Url: search.Query.Url})
	} else {
		search.Results = results
	}
}
