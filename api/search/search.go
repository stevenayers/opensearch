package search

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"go-clamber/service/crawl"
	"go-clamber/service/database"
	"go-clamber/service/page"
)

type (
	Search struct {
		Query   database.Query   `json:"query"`
		Results database.Results `json:"results"`
	}
)

func (search Search) Initiate() {
	db, err := sql.Open("sqlite3", "testing/pages.sqlite")
	if err != nil {
		fmt.Print(err)
	}
	database.InitStore(&database.DbStore{Db: db})
	results, err := database.DB.Get(search.Query)
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
