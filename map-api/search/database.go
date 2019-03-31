package search

import (
	"database/sql"
	"go-clamber/map-api/models"
)

func GetResults(query *models.Query) (results *models.Results) {
	db, err := sql.Open("mysql", "root@/blog")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	return
}
