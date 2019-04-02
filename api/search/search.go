package search

import "go-clamber/service/page"

type (
	Results []page.Page

	Queries []Query

	Query struct {
		Url                string `json:"url"`
		Depth              int    `json:"depth"`
		AllowExternalLinks bool   `json:"allow_external_links"`
	}

	Search struct {
		Query   Query   `json:"query"`
		Results Results `json:"results"`
	}
)

func (Search) Initiate() {
	// Check database for query
	// If exists, return
	// If not, send URL to producer.
}
