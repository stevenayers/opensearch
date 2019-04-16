package database

import (
	"encoding/json"
	"fmt"
	"go-clamber/page"
)

type (
	JSONPage struct {
		Uid       string      `json:"uid,omitempty"`
		Url       string      `json:"url,omitempty"`
		Timestamp int64       `json:"timestamp,omitempty"`
		Children  []*JSONPage `json:"child.Of,omitempty"`
	}
	PredicateCount struct {
		Matching int `json:"matching"`
	}
)

func ConvertPageToJson(currentPage *page.Page) (pb []byte, err error) {
	p := JSONPage{
		Url:       currentPage.Url,
		Timestamp: currentPage.Timestamp,
	}
	pb, err = json.Marshal(p)
	if err != nil {
		fmt.Print(err)
	}
	return
}

func ConvertJsonToPage(pb []byte) (currentPage *page.Page, err error) {
	jsonMap := make(map[string][]JSONPage)
	err = json.Unmarshal(pb, &jsonMap)
	jsonPages := jsonMap["result"]
	if len(jsonPages) > 0 {
		currentPage = &page.Page{
			Url:       jsonPages[0].Url,
			Timestamp: jsonPages[0].Timestamp,
		}
	}

	return
}
