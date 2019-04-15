package database

import (
	"encoding/json"
	"go-clamber/page"
	"log"
)

type (
	JSONPage struct {
		Uid       string      `json:"uid,omitempty"`
		Url       string      `json:"url,omitempty"`
		Timestamp int64       `json:"timestamp,omitempty"`
		Body      string      `json:"body,omitempty"`
		Children  []*JSONPage `json:"child.Of,omitempty"`
	}
)

func ConvertPageToJson(currentPage *page.Page) (pb []byte, err error) {
	p := JSONPage{
		Url:       currentPage.Url,
		Timestamp: currentPage.Timestamp,
		Body:      currentPage.Body,
	}
	pb, err = json.Marshal(p)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func ConvertJsonToPage(pb []byte) (currentPage *page.Page, err error) {
	jsonMap := make(map[string][]JSONPage)
	err = json.Unmarshal(pb, &jsonMap)
	jsonPage := jsonMap["result"][0]
	currentPage = &page.Page{
		Url:       jsonPage.Url,
		Timestamp: jsonPage.Timestamp,
		Body:      jsonPage.Body,
	}
	return
}
