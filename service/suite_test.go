package service_test

import (
	"clamber/logging"
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
	tempConfigFile := "../Config.toml"
	service.InitConfig(tempConfigFile)
	logging.InitJsonLogger(service.AppConfig.General.LogLevel)
	s.store = service.DbStore{}
	service.Connect(&s.store, service.AppConfig.Database)
	err := service.DB.DeleteAll()
	if err != nil {
		s.T().Fatal(err)
	}
	err = s.store.SetSchema()
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *StoreSuite) SetupTest() {
	err := service.DB.DeleteAll()
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
