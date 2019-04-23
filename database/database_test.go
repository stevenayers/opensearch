package database_test

import (
	"clamber/crawl"
	"clamber/database"
	"clamber/page"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
	"time"
)

type (
	StoreSuite struct {
		suite.Suite
		store database.DbStore
	}

	NodeTest struct {
		Url   string
		Depth int
	}
)

var (
	NodeTests = []NodeTest{
		{"https://golang.org", 2},
		{"https://google.com", 1},
		{"https://youtube.com", 2},
	}
)

func (s *StoreSuite) SetupSuite() {
	s.store = database.DbStore{}
	database.Connect(&s.store)
	err := database.DB.DeleteAll()
	if err != nil {
		s.T().Fatal(err)
	}
	err = s.store.SetSchema()
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *StoreSuite) SetupTest() {
	err := database.DB.DeleteAll()
	if err != nil {
		s.T().Fatal(err)
	}
	err = s.store.SetSchema()
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *StoreSuite) TearDownSuite() {
	for _, conn := range s.store.Connection {
		err := conn.Close()
		if err != nil {
			fmt.Print(err)
		}
	}
}

func TestStoreSuite(t *testing.T) {
	s := new(StoreSuite)
	suite.Run(t, s)
}

func (s *StoreSuite) TestCreateAndCheckPredicate() {
	for _, test := range NodeTests {
		expectedPage := page.Page{
			Url:       test.Url,
			Timestamp: time.Now().Unix(),
		}
		crawler := crawl.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		crawler.Crawl(&expectedPage, 1)
		crawler.DbWaitGroup.Wait()
		ctx := context.Background()
		txn := s.store.NewTxn()
		exists, err := s.store.CheckPredicate(&ctx, txn, expectedPage.Uid, expectedPage.Children[0].Uid)
		if err != nil {
			s.T().Fatal(err)
		}
		assert.Equal(s.T(), true, exists, "Predicate should have existed.")
	}
}

func (s *StoreSuite) TestCreateAndFindNode() {
	for _, test := range NodeTests {
		expectedPage := page.Page{
			Url:       test.Url,
			Timestamp: time.Now().Unix(),
		}
		crawler := crawl.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		crawler.Crawl(&expectedPage, test.Depth)
		crawler.DbWaitGroup.Wait()
		ctx := context.Background()
		txn := s.store.NewTxn()
		resultPage, err := s.store.FindNode(&ctx, txn, test.Url, test.Depth)
		if err != nil {
			s.T().Fatal(err)
		}
		assert.Equal(s.T(), expectedPage.Uid, resultPage.Uid, "Uid should have matched.")
		assert.Equal(s.T(), expectedPage.Url, resultPage.Url, "Url should have matched.")
		assert.Equal(s.T(), expectedPage.Timestamp, resultPage.Timestamp, "Timestamp should have matched.")
	}

}
