package relationship_test

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/stevenayers/clamber/pkg/config"
	"github.com/stevenayers/clamber/pkg/crawl"
	"github.com/stevenayers/clamber/pkg/database/relationship"
	"github.com/stevenayers/clamber/pkg/logging"
	"github.com/stevenayers/clamber/pkg/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
	"time"
)

type (
	StoreSuite struct {
		suite.Suite
		store   relationship.Store
		crawler crawl.Crawler
	}
)

func (s *StoreSuite) SetupSuite() {
	var err error
	configFile := "/Users/steven/git/clamber/configs/config.toml"
	err = config.InitConfig(configFile)
	if err != nil {
		s.T().Fatal(err)
	}

	logging.InitJsonLogger(log.NewSyncWriter(os.Stdout), config.AppConfig.Service.LogLevel, "test")
	s.store = relationship.Store{}
	s.store.Connect()
}

func (s *StoreSuite) SetupTest() {
	var err error
	configFile := "/Users/steven/git/clamber/configs/config.toml"
	err = config.InitConfig(configFile)
	if err != nil {
		s.T().Fatal(err)
	}
	s.store.Connect()
	if !strings.Contains(s.T().Name(), "TestLog") && !strings.Contains(s.T().Name(), "TestConnect") {
		err := s.store.DeleteAll()
		if err != nil {
			s.T().Fatal(err)
		}
		err = s.store.SetSchema()
		if err != nil {
			s.T().Fatal(err)
		}
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

func TestSuite(t *testing.T) {
	s := new(StoreSuite)
	suite.Run(t, s)
}

func (s *StoreSuite) TestBadConnectionsSetSchema() {
	db := relationship.Store{}
	dbConfig := config.DatabaseConfig{
		Connections: []*config.Connection{
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
	txn := s.store.DB.NewTxn()
	ctx := context.Background()
	txn.Discard(ctx)
	_, err := s.store.FindNode(&ctx, txn, "https://golang.org", 0)
	assert.Equal(s.T(), true, err != nil)
}

func (s *StoreSuite) TestFindNodeBadDepth() {
	ctx := context.Background()
	p := page.Page{Url: "https://golang.org", Timestamp: time.Now().Unix()}
	txn := s.store.DB.NewTxn()
	_, err := s.store.FindOrCreateNode(&ctx, txn, &p)
	if err != nil {
		s.T().Fatal(err)
		return
	}
	txn = s.store.DB.NewTxn()
	_, err = s.store.FindNode(&ctx, txn, "https://golang.org", 9)
	assert.Equal(s.T(), true, strings.Contains(err.Error(), "Depth does not match dgraph result."))
}

func (s *StoreSuite) TestCheckPredicateBadTransaction() {
	txn := s.store.DB.NewTxn()
	ctx := context.Background()
	txn.Discard(ctx)
	_, err := s.store.CheckPredicate(&ctx, txn, "fakeuid1", "fakeuid2")
	assert.Equal(s.T(), true, err != nil)
}

func (s *StoreSuite) TestCheckOrCreatePredicateBadTransaction() {
	txn := s.store.DB.NewTxn()
	ctx := context.Background()
	txn.Discard(ctx)
	_, err := s.store.CheckOrCreatePredicate(&ctx, txn, "fakeuid1", "fakeuid2")
	assert.Equal(s.T(), true, err != nil)
}

func (s *StoreSuite) TestDeserializePredicateDoesntExist() {
	pb := []byte(`{"edges":[{"matching":0}]}`)
	exists, err := page.DeserializePredicate(pb)
	if err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), false, exists)
}

func (s *StoreSuite) TestDeserializePredicateError() {
	pb := []byte(`{"edges":[{"matching":"hello"}]}`)
	_, err := page.DeserializePredicate(pb)
	assert.Equal(s.T(), true, err != nil)
}
