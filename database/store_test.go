package database_test

//
// TODO need to overhaul this for the new database methods
//
//import (
//	"context"
//	"fmt"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/suite"
//	"go-clamber/database"
//	"go-clamber/page"
//	"log"
//	"net/url"
//	"testing"
//	"time"
//)
//
//type StoreSuite struct {
//	suite.Suite
//	store database.DbStore
//}
//
//func (s *StoreSuite) SetupSuite() {
//	s.store = database.DbStore{}
//	database.InitStore(&s.store)
//	err := database.DB.DeleteAll()
//	if err != nil {
//		s.T().Fatal(err)
//	}
//	err = s.store.SetSchema()
//	if err != nil {
//		s.T().Fatal(err)
//	}
//}
//
//func (s *StoreSuite) SetupTest() {
//	err := database.DB.DeleteAll()
//	if err != nil {
//		s.T().Fatal(err)
//	}
//}
//
//func (s *StoreSuite) TearDownSuite() {
//	err := s.store.Connection.Close()
//	if err != nil {
//		fmt.Print(err)
//	}
//}
//
//func TestStoreSuite(t *testing.T) {
//	s := new(StoreSuite)
//	suite.Run(t, s)
//}
//
//func (s *StoreSuite) TestCreateChildPage() {
//	ctx := context.Background()
//
//	// Create Parent page and send it to the database
//	parentUrl, _ := url.Parse("https://google.com/")
//	p := page.Page{
//		Url:  parentUrl.String(),
//		Timestamp: time.Now().Unix(),
//		Body: "test",
//	}
//	parentUid, err := s.store.CreatePage(ctx, &p)
//	if err != nil {
//		s.T().Fatal(err)
//	}
//	// Create child page and send to database
//	childUrl, _ := url.Parse("https://google.com/")
//	_ = page.Page{
//		Url:  childUrl.String(),
//		Timestamp: time.Now().Unix(),
//		Body: "test",
//	}
//	currentPage, err := s.store.GetPageByUid(ctx, parentUid)
//	if err != nil {
//		log.Fatal(err)
//		return
//	}
//	assert.Equal(s.T(), p, *currentPage, "Expected same input as output")
//}
//
//func (s *StoreSuite) TestCreatePage() {
//	ctx := context.Background()
//	parsedUrl, _ := url.Parse("https://google.com/")
//	p := page.Page{
//		Url:  parsedUrl.String(),
//		Timestamp: time.Now().Unix(),
//		Body: "test",
//	}
//	uid, err := s.store.CreatePage(ctx, &p)
//	if err != nil {
//		s.T().Fatal(err)
//	}
//	currentPage, err := s.store.GetPageByUid(ctx, uid)
//	if err != nil {
//		log.Fatal(err)
//		return
//	}
//	assert.Equal(s.T(), p, *currentPage, "Expected same input as output")
//}
//
//func (s *StoreSuite) TestGet() {
//
//	//testUrl := "https://google.com"
//	//parsedUrl, _ := url.Parse(testUrl)
//	//expectedPage := page.Page{
//	//	Url:  parsedUrl,
//	//	Body: "test",
//	//}
//	//_, err := s.store.Run("CREATE (n:Page { url: $url, body: $body }) RETURN n.url, n.body", map[string]interface{}{
//	//	"url":  parsedUrl.String(),
//	//	"body": "test",
//	//})
//	//if err != nil {
//	//	s.T().Fatal(err)
//	//}
//	//pages, err := s.store.GetPage(testUrl)
//	//if err != nil {
//	//	s.T().Fatal(err)
//	//}
//	//assert.Equal(s.T(), 1, len(pages), "Expected matching length")
//	//assert.Equal(s.T(), expectedPage, *&pages[0], "Expected matching page data")
//}
