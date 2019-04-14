package database_test

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-clamber/database"
	"go-clamber/page"
	"net/url"
	"testing"
)

type StoreSuite struct {
	suite.Suite
	store  *database.DbStore
	driver neo4j.Driver
}

func (s *StoreSuite) SetupSuite() {
	s.store = &database.DbStore{}
	var err error
	s.driver, err = neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		fmt.Print(err)
	}
	s.store.Session, err = s.driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *StoreSuite) SetupTest() {
	result, err := s.store.Run("MATCH (n) DELETE n", map[string]interface{}{})
	if err = result.Err(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *StoreSuite) TearDownSuite() {
	err := s.driver.Close()
	if err != nil {
		fmt.Print(err)
	}
}

func TestStoreSuite(t *testing.T) {
	s := new(StoreSuite)
	suite.Run(t, s)
}

func (s *StoreSuite) TestCreatePage() {
	defer s.store.Close()
	parsedUrl, _ := url.Parse("https://google.com")
	err := s.store.CreatePage(&page.Page{
		Url:  parsedUrl,
		Body: "",
	})
	result, err := s.store.Run("MATCH (n:Page) WHERE n.url = $url RETURN count(*)", map[string]interface{}{
		"url": parsedUrl.String(),
	})
	count := 0
	for result.Next() {
		count = int(result.Record().GetByIndex(0).(int64))
	}
	if err = result.Err(); err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), 1, count, "Expected to find created page.")

}

func (s *StoreSuite) TestGet() {
	testUrl := "https://google.com"
	parsedUrl, _ := url.Parse(testUrl)
	expectedPage := page.Page{
		Url:  parsedUrl,
		Body: "test",
	}
	_, err := s.store.Run("CREATE (n:Page { url: $url, body: $body }) RETURN n.url, n.body", map[string]interface{}{
		"url":  parsedUrl.String(),
		"body": "test",
	})
	if err != nil {
		s.T().Fatal(err)
	}
	pages, err := s.store.GetPage(testUrl)
	if err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), 1, len(pages), "Expected matching length")
	assert.Equal(s.T(), expectedPage, *&pages[0], "Expected matching page data")
}
