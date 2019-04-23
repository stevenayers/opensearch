package service_test

import (
	"clamber/service"
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"time"
)

type (
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

func (s *StoreSuite) TestCreateAndCheckPredicate() {
	for _, test := range NodeTests {
		expectedPage := service.Page{
			Url:       test.Url,
			Timestamp: time.Now().Unix(),
		}
		crawler := service.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
		crawler.Crawl(&expectedPage, 1)
		crawler.DbWaitGroup.Wait()
		ctx := context.Background()
		txn := s.store.NewTxn()
		exists, err := s.store.CheckPredicate(&ctx, txn, expectedPage.Uid, expectedPage.Links[0].Uid)
		if err != nil {
			s.T().Fatal(err)
		}
		assert.Equal(s.T(), true, exists, "Predicate should have existed.")
	}
}

func (s *StoreSuite) TestCreateAndFindNode() {
	for _, test := range NodeTests {
		expectedPage := service.Page{
			Url:       test.Url,
			Timestamp: time.Now().Unix(),
		}
		crawler := service.Crawler{DbWaitGroup: sync.WaitGroup{}, AlreadyCrawled: make(map[string]struct{})}
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
