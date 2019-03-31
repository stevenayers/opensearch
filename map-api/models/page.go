package models

import "time"

type Page struct {
	Url   string    `json:"url"`
	Links []string  `json:"links"`
	Due   time.Time `json:"due"`
}

type Results []Page
