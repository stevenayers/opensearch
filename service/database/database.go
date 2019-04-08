package database

import (
	"database/sql"
	"go-clamber/service/page"
	"net/url"
	"time"
)

type (
	Results []page.Page

	Queries []Query

	Query struct {
		Url                *url.URL `json:"url"`
		Depth              int      `json:"depth"`
		AllowExternalLinks bool     `json:"allow_external_links"`
	}

	Store interface {
		Get(query Query) (Results, error)
		Create(page *page.Page) error
	}

	DbStore struct {
		Db *sql.DB
	}
)

func (store *DbStore) Create(page *page.Page) error {
	_, err := store.Db.Exec(
		"INSERT INTO pages(url, parent, timestamp, body) VALUES ($1,$2,$3,$4)",
		page.Url.String(), nil, time.Now(), nil,
	)
	return err
}

func (store *DbStore) Get(query Query) (Results, error) {
	rows, err := store.Db.Query("SELECT url from pages")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages Results
	for rows.Next() {
		var Url string
		if err := rows.Scan(&Url); err != nil {
			return nil, err
		}
		parsedUrl, _ := url.Parse(Url)
		thisPage := page.Page{Url: parsedUrl}
		pages = append(pages, thisPage)
	}
	return pages, nil
}

var DB Store

func InitStore(s Store) {
	DB = s
}
