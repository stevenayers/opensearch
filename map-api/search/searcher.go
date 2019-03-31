package search

import (
	"go-clamber/map-api/models"
)

type Searcher struct {
	Query   models.Query
	Results models.Results
}

func (Searcher) Search() {

}
