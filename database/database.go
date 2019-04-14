package database

import (
	"fmt"
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
		GetPage(pageUrl string) (Results, error)
		GetRelationship(parentPage *page.Page, childPage *page.Page) (relationshipCount int, err error)
		Create(currentPage *page.Page) (err error)
		CreatePage(currentPage *page.Page) (err error)
		CreateRelationship(parentPage *page.Page, childPage *page.Page) (err error)
		CreateChildPage(parentPage *page.Page, childPage *page.Page) (err error)
	}

	DbStore struct {
		Driver neo4j.Driver
	}
)

func (store *DbStore) CreatePage(currentPage *page.Page) (err error) {
	session, err := store.Driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return
	}
	defer session.Close()
	_, err = session.Run(`CREATE (n:Page { url: $url, body: $body }) RETURN n.url, n.body`,
		map[string]interface{}{
			"url":  currentPage.Url.String(),
			"body": currentPage.Body,
		})
	if err != nil {
		fmt.Print(err)
	}
	return err
}

func (store *DbStore) CreateRelationship(parentPage *page.Page, childPage *page.Page) (err error) {
	session, err := store.Driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return
	}
	defer session.Close()
	_, err = session.Run(
		`MATCH (p:Page), (c:Page)
				WHERE p.url = $parentUrl AND c.url = $childUrl
				CREATE (c)-[r:CHILD_OF]->(p)`,
		map[string]interface{}{
			"parentUrl": parentPage.Url.String(),
			"childUrl":  childPage.Url.String(),
		})
	if err != nil {
		fmt.Print(err)
	}
	return err
}

func (store *DbStore) CreateChildPage(parentPage *page.Page, childPage *page.Page) (err error) {
	session, err := store.Driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return
	}
	defer session.Close()
	_, err = session.Run(
		`MATCH (p:Page)
				WHERE p.url = $parentUrl
				CREATE (c:Page { url: $childUrl, body: $childBody })-[r:CHILD_OF]->(p)`,
		map[string]interface{}{
			"parentUrl": parentPage.Url.String(),
			"childUrl":  childPage.Url.String(),
			"childBody": childPage.Body,
		})
	if err != nil {
		fmt.Print(err)
	}
	return err
}

func (store *DbStore) GetPage(pageUrl string) (pages Results, err error) {
	session, err := store.Driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return
	}
	defer session.Close()
	result, err := session.Run(`MATCH (n:Page) WHERE n.url = $url RETURN n.url, n.body`, map[string]interface{}{
		"url": pageUrl,
	})
	if err != nil {
		fmt.Print(err)
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

func (store *DbStore) GetRelationship(parentPage *page.Page, childPage *page.Page) (relationshipCount int, err error) {
	session, err := store.Driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return
	}
	defer session.Close()
	result, err := session.Run(
		`MATCH (p)<-[r:CHILD_OF]-(c) 
				WHERE p.url = $parentUrl AND c.url = $childUrl
				RETURN count(r)`,
		map[string]interface{}{
			"parentUrl": parentPage.Url.String(),
			"childUrl":  childPage.Url.String(),
		})
	if err != nil {
		fmt.Print(err)
		return
	}
	for result.Next() {
		relationshipCount = int(result.Record().GetByIndex(0).(int64))
	}
	return
}

func (store *DbStore) Create(currentPage *page.Page) (err error) {

	currentResults, err := DB.GetPage(currentPage.Url.String())
	currentPageNotFound := len(currentResults) == 0

	if currentPage.Parent == nil && currentPageNotFound {
		err = DB.CreatePage(currentPage)
		return
	}
	if currentPage.Parent == nil && !currentPageNotFound {
		return
	}
	if currentPage.Parent != nil && currentPageNotFound {
		err = DB.CreateChildPage(currentPage.Parent, currentPage)
		return
	}
	if currentPage.Parent != nil && !currentPageNotFound {
		relationshipCount, err := DB.GetRelationship(currentPage.Parent, currentPage)
		if err != nil {
			return err
		}
		relationshipNotFound := relationshipCount == 0
		if relationshipNotFound {
			err = DB.CreateRelationship(currentPage.Parent, currentPage)
			return err
		}
	}

	return
}

var DB Store

func InitStore(s Store) {
	DB = s
}
