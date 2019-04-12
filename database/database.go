package database

import (
	"database/sql"
	"fmt"
	"go-clamber/page"
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

	var err error
	if page.Parent == nil {
		_, err = store.Db.Exec(
			"INSERT INTO pages(url, domain, parent, timestamp, body) VALUES ($1,$2,$3,$4,$5)",
			page.Url.String(), page.Url.Host, nil, time.Now(), page.Body,
		)
		if err != nil {
			fmt.Print(err)
			fmt.Println("LA")
		}
	} else {
		_, err = store.Db.Exec(
			"INSERT INTO pages(url, domain, parent, timestamp, body) VALUES ($1,$2,$3,$4,$5)",
			page.Url.String(), page.Url.Host, page.Parent.Url.String(), time.Now(), page.Body,
		)
		if err != nil {
			fmt.Print(err)
			fmt.Println("LAA")
		}
	}

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
