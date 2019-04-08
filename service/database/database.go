package database

import (
	"database/sql"
	"go-clamber/service/page"
	"net/url"
	"time"
)

type (
	Store interface {
		Create()
		FindByUrl()
		Update()
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

func (store *DbStore) Get() ([]*page.Page, error) {
	rows, err := store.Db.Query("SELECT url from pages")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []*page.Page
	for rows.Next() {
		var Url string
		if err := rows.Scan(&Url); err != nil {
			return nil, err
		}
		parsedUrl, _ := url.Parse(Url)
		thisPage := page.Page{Url: parsedUrl}
		pages = append(pages, &thisPage)
	}
	return pages, nil
}

var store Store

func InitStore(s Store) {
	store = s
}
