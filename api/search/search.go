package search

import "go-clamber/service/page"

type Results []page.Page

type Search struct {
	Query   Query   `json:"query"`
	Results Results `json:"results"`
}

func (Search) Initiate() {
	// Check database for query
	// If exists, return
	// If not, send URL to producer.
}
