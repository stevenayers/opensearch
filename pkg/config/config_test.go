package config_test

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/stevenayers/opensearch/pkg/config"
	"github.com/stevenayers/opensearch/pkg/crawl"
	"github.com/stevenayers/opensearch/pkg/database/relationship"
	"github.com/stevenayers/opensearch/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"strings"
	"testing"
)

type StoreSuite struct {
	suite.Suite
	store   relationship.Store
	crawler crawl.Crawler
}

func (s *StoreSuite) SetupSuite() {
	var err error
	configFile := "/Users/steven/git/opensearch/configs/config.toml"
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
	configFile := "/Users/steven/git/opensearch/configs/config.toml"
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

func (s *StoreSuite) TestConfigPath() {
	configFile := "../test/incorrectpath.toml"
	err := config.InitConfig(configFile)
	assert.Equal(s.T(), true, err != nil)
	if err != nil {
		assert.Equal(s.T(), true, strings.Contains(
			err.Error(), "no such file or directory"))
	}
}

func (s *StoreSuite) TestConfigParse() {
	configFile := "/Users/steven/git/opensearch/test/invaild_config.toml"
	err := config.InitConfig(configFile)
	assert.Equal(s.T(), true, err != nil)
	if err != nil {
		assert.Equal(s.T(), true, strings.Contains(
			err.Error(), "cannot load TOML value of type string into a Go integer"))
	}
}
