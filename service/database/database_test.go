package database_test

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-clamber/service/database"
	"go-clamber/service/page"
	"net/url"
	"testing"
)

type StoreSuite struct {
	suite.Suite
	store *database.DbStore
	db    *sql.DB
}

func (s *StoreSuite) SetupSuite() {
	db, err := sql.Open("sqlite3", "testing/pages.sqlite")
	if err != nil {
		s.T().Fatal(err)
	}
	s.db = db
	s.store = &database.DbStore{Db: db}
}

func (s *StoreSuite) SetupTest() {
	_, err := s.db.Exec("DELETE FROM pages")
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *StoreSuite) TearDownSuite() {
	err := s.db.Close()
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func TestStoreSuite(t *testing.T) {
	s := new(StoreSuite)
	suite.Run(t, s)
}

func (s *StoreSuite) TestCreatePage() {
	parsedUrl, _ := url.Parse("https://google.com")
	err := s.store.Create(&page.Page{
		Url: parsedUrl,
	})
	if err != nil {
		s.T().Fatal(err)
	}
	res, err := s.db.Query(`SELECT COUNT(*) FROM pages WHERE url=$1`, parsedUrl.String())
	if err != nil {
		s.T().Fatal(err)
	}
	var count int
	for res.Next() {
		err := res.Scan(&count)
		if err != nil {
			s.T().Error(err)
		}
	}
	assert.Equal(s.T(), 1, count, "Expected to find created page.")
}

func (s *StoreSuite) TestGet() {
	testUrl := "https://google.com"
	parsedUrl, _ := url.Parse(testUrl)
	expectedPage := page.Page{Url: parsedUrl}
	_, err := s.db.Exec(`INSERT INTO pages (url) VALUES($1)`, testUrl)
	if err != nil {
		s.T().Fatal(err)
	}
	pages, err := s.store.Get()
	if err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), 1, len(pages), "Expected matching length")
	assert.Equal(s.T(), expectedPage, *pages[0], "Expected matching page data")
}
