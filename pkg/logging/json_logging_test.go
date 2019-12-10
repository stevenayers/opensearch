package logging_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/stevenayers/clamber/pkg/config"
	"github.com/stevenayers/clamber/pkg/crawl"
	"github.com/stevenayers/clamber/pkg/database/relationship"
	"github.com/stevenayers/clamber/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type (
	LogOutput struct {
		Level   string `json:"level,omitempty"`
		Node    string `json:"node,omitempty"`
		Service string `json:"app,omitempty"`
		Msg     string `json:"msg,omitempty"`
		Context string `json:"context,omitempty"`
	}

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

func (s *StoreSuite) TestLogDebug() {
	buf := new(bytes.Buffer)
	var logOutput LogOutput
	logging.InitJsonLogger(buf, "debug", "test")
	err := level.Debug(logging.Logger).Log("msg", "test")
	if err != nil {
		s.T().Fatal(err)
	}
	err = json.Unmarshal(buf.Bytes(), &logOutput)
	if err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), "debug", logOutput.Level)
	assert.Equal(s.T(), "test", logOutput.Msg)
}

func (s *StoreSuite) TestLogInfo() {
	buf := new(bytes.Buffer)
	var logOutput LogOutput
	logging.InitJsonLogger(buf, "info", "test")
	err := level.Info(logging.Logger).Log("msg", "test")
	if err != nil {
		s.T().Fatal(err)
	}
	err = json.Unmarshal(buf.Bytes(), &logOutput)
	if err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), "info", logOutput.Level)
	assert.Equal(s.T(), "test", logOutput.Msg)
}

func (s *StoreSuite) TestLogError() {
	buf := new(bytes.Buffer)
	var logOutput LogOutput
	logging.InitJsonLogger(buf, "error", "test")
	err := level.Error(logging.Logger).Log("msg", "test")
	if err != nil {
		s.T().Fatal(err)
	}
	err = json.Unmarshal(buf.Bytes(), &logOutput)
	if err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), "error", logOutput.Level)
	assert.Equal(s.T(), "test", logOutput.Msg)
}

func (s *StoreSuite) TestLogFilter() {
	buf := new(bytes.Buffer)
	logging.InitJsonLogger(buf, "defaultShouldGoToInfo", "test")
	err := level.Debug(logging.Logger).Log("msg", "test")
	if err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), "", buf.String())
}
