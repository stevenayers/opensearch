package service_test

import (
	"context"
	"github.com/stevenayers/clamber/service"
	"github.com/stretchr/testify/assert"
	"log"
	"strings"
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
		{"https://golang.org", 1},
		{"https://google.com", 1},
		{"https://youtube.com", 1},
	}
)

func (s *StoreSuite) TestBadConnectionsSetSchema() {
	db := service.DbStore{}
	dbConfig := service.DatabaseConfig{
		Connections: []*service.Connection{
			{Host: "fakehost.local",
				Port: 999999},
		},
	}
	db.Connect(dbConfig)
	err := db.SetSchema()
	assert.Equal(s.T(), true, err != nil)
	if err != nil {
		assert.Equal(s.T(), true, strings.Contains(
			err.Error(), "transport: Error while dialing dial tcp"))
	}
}

func (s *StoreSuite) TestFindNodeBadTransaction() {
	txn := s.store.NewTxn()
	ctx := context.Background()
	txn.Discard(ctx)
	_, err := s.store.FindNode(&ctx, txn, "https://golang.org", 0)
	assert.Equal(s.T(), true, err != nil)
}

func (s *StoreSuite) TestFindNodeBadDepth() {
	ctx := context.Background()
	p := service.Page{Url: "https://golang.org", Timestamp: time.Now().Unix()}
	txn := s.store.NewTxn()
	_, err := s.store.FindOrCreateNode(&ctx, txn, &p)
	if err != nil {
		s.T().Fatal(err)
		return
	}
	txn = s.store.NewTxn()
	_, err = s.store.FindNode(&ctx, txn, "https://golang.org", 9)
	assert.Equal(s.T(), true, strings.Contains(err.Error(), "Depth does not match dgraph result."))
}

func (s *StoreSuite) TestCheckPredicateBadTransaction() {
	txn := s.store.NewTxn()
	ctx := context.Background()
	txn.Discard(ctx)
	_, err := s.store.CheckPredicate(&ctx, txn, "fakeuid1", "fakeuid2")
	assert.Equal(s.T(), true, err != nil)
}

func (s *StoreSuite) TestCheckOrCreatePredicateBadTransaction() {
	txn := s.store.NewTxn()
	ctx := context.Background()
	txn.Discard(ctx)
	_, err := s.store.CheckOrCreatePredicate(&ctx, txn, "fakeuid1", "fakeuid2")
	assert.Equal(s.T(), true, err != nil)
}

func (s *StoreSuite) TestCreateAndCheckPredicate() {
	for _, test := range NodeTests {
		expectedPage := service.Page{
			Url:       test.Url,
			Timestamp: time.Now().Unix(),
			Depth:     1,
		}
		crawler := service.Crawler{
			DbWaitGroup:    sync.WaitGroup{},
			AlreadyCrawled: make(map[string]struct{}),
			Logger:         s.logger,
			Config:         s.config,
			Db:             s.store,
		}
		crawler.Crawl(&expectedPage)
		crawler.DbWaitGroup.Wait()
		ctx := context.Background()
		txn := s.store.NewTxn()
		exists, err := s.store.CheckPredicate(&ctx, txn, expectedPage.Uid, expectedPage.Links[2].Uid)
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
			Depth:     test.Depth,
		}
		crawler := service.Crawler{
			DbWaitGroup:    sync.WaitGroup{},
			AlreadyCrawled: make(map[string]struct{}),
			Logger:         s.logger,
			Config:         s.config,
			Db:             s.store,
		}
		crawler.Crawl(&expectedPage)
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

func (s *StoreSuite) TestCreateError() {
	p := service.Page{Url: "https://golang.org"}
	err := s.store.DeleteAll()
	if err != nil {
		log.Fatalln(err)
	}
	s.crawler.Db = s.store
	s.crawler.Logger = s.logger
	err = s.crawler.Create(&p)
	assert.Equal(s.T(), true, err != nil)
	c := service.Page{Parent: &p}
	err = s.crawler.Create(&c)
	assert.Equal(s.T(), true, err != nil)
}

func (s *StoreSuite) TestDeserializePredicateDoesntExist() {
	pb := []byte(`{"edges":[{"matching":0}]}`)
	exists, err := service.DeserializePredicate(pb)
	if err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), false, exists)
}

func (s *StoreSuite) TestDeserializePredicateError() {
	pb := []byte(`{"edges":[{"matching":"hello"}]}`)
	_, err := service.DeserializePredicate(pb)
	assert.Equal(s.T(), true, err != nil)
}
