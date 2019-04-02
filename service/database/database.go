package database

type (
	db interface {
		Create()
		FindByUrl()
		Update()
	}
)
