package service_test

import (
	"fmt"
	kitlog "github.com/go-kit/kit/log"
	"github.com/stevenayers/clamber/service"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
)

type (
	StoreSuite struct {
		suite.Suite
		store service.DbStore
	}
)

func (s *StoreSuite) SetupSuite() {
	service.InitFlags()
	err := service.InitConfig()
	if err != nil {
		s.T().Fatal(err)
	}
	service.APILogger.InitJsonLogger(kitlog.NewSyncWriter(os.Stdout), service.AppConfig.General.LogLevel)
	s.store = service.DbStore{}
	service.Connect(&s.store, service.AppConfig.Database)
}

func (s *StoreSuite) SetupTest() {
	*service.AppFlags.ConfigFile = "../cmd/Config.toml"
	err := service.InitConfig()
	if err != nil {
		s.T().Fatal(err)
	}
	service.Connect(&s.store, service.AppConfig.Database)
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
