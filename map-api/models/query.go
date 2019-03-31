package models

type Query struct {
	Url                string `json:"url"`
	Depth              int    `json:"depth"`
	AllowExternalLinks bool   `json:"allow_external_links"`
}

type Queries []Query
