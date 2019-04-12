package database

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"go-clamber/page"
	"net/url"
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
		neo4j.Session
	}
)

func (store *DbStore) Create(page *page.Page) error {
	_, err := store.Run("CREATE (n:Page { url: $url, body: $body }) RETURN n.url, n.body", map[string]interface{}{
		"url":  page.Url.String(),
		"body": page.Body,
	})
	return err
}

func (store *DbStore) Get(queryUrl string) (pages Results, err error) {
	result, err := store.Run("MATCH (n:Page) WHERE n.url = $url RETURN n.url, n.body", map[string]interface{}{
		"url": queryUrl,
	})
	if err != nil {
		return
	}
	for result.Next() {
		parsedUrl, _ := url.Parse(result.Record().GetByIndex(0).(string))
		thisPage := page.Page{
			Url:  parsedUrl,
			Body: result.Record().GetByIndex(1).(string),
		}
		pages = append(pages, thisPage)
	}
	return
}

var DB Store

func InitStore(s Store) {
	DB = s
}
