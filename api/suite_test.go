package api_test

import (
	"fmt"
	"github.com/stevenayers/clamber/service"
	"github.com/stretchr/testify/suite"
	"testing"
)

type (
	StoreSuite struct {
		suite.Suite
		store service.DbStore
	}
)

func (s *StoreSuite) SetupSuite() {
	service.InitConfig()
	service.APILogger.InitJsonLogger(service.AppConfig.General.LogLevel)
	s.store = service.DbStore{}
	service.Connect(&s.store, service.AppConfig.Database)
}

func (s *StoreSuite) SetupTest() {
	err := s.store.DeleteAll()
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