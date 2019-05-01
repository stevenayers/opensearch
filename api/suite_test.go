package api_test

import (
	"github.com/stevenayers/clamber/api"
	"github.com/stevenayers/clamber/service"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type (
	StoreSuite struct {
		suite.Suite
		store   service.DbStore
		crawler service.Crawler
		logger  log.Logger
		config  service.Config
	}
)

func (s *StoreSuite) SetupSuite() {
	var err error
	api.InitFlags(&api.AppFlags)
	*api.AppFlags.ConfigFile = "../test/Config.toml"
	s.config, err = service.InitConfig(*api.AppFlags.ConfigFile)
	if err != nil {
		s.T().Fatal(err)
	}
	s.logger = api.InitJsonLogger(log.NewSyncWriter(os.Stdout), s.config.General.LogLevel)
	s.store = service.DbStore{}
	s.store.Connect(s.config.Database)

}

func (s *StoreSuite) SetupTest() {
	var err error
	*api.AppFlags.ConfigFile = "../test/Config.toml"
	s.config, err = service.InitConfig(*api.AppFlags.ConfigFile)
	s.store.Connect(s.config.Database)
	err = s.store.DeleteAll()
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

func TestSuite(t *testing.T) {
	s := new(StoreSuite)
	suite.Run(t, s)
}
