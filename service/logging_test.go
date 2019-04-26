package service_test

import (
	"bytes"
	"encoding/json"
	"github.com/stevenayers/clamber/service"
	"github.com/stretchr/testify/assert"
)

type (
	LogOutput struct {
		Level   string `json:"level,omitempty"`
		Node    string `json:"node,omitempty"`
		Service string `json:"service,omitempty"`
		Msg     string `json:"msg,omitempty"`
		Context string `json:"context,omitempty"`
	}
)

func (s *StoreSuite) TestLogDebug() {
	buf := new(bytes.Buffer)
	testLogger := service.ApiLogger{}
	var logOutput LogOutput
	testLogger.InitJsonLogger(buf, "debug")
	err := testLogger.LogDebug("msg", "test")
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
	testLogger := service.ApiLogger{}
	var logOutput LogOutput
	testLogger.InitJsonLogger(buf, "info")
	err := testLogger.LogInfo("msg", "test")
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
	testLogger := service.ApiLogger{}
	var logOutput LogOutput
	testLogger.InitJsonLogger(buf, "error")
	err := testLogger.LogError("msg", "test")
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
	testLogger := service.ApiLogger{}
	testLogger.InitJsonLogger(buf, "defaultShouldGoToInfo")
	err := testLogger.LogDebug("msg", "test")
	if err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), "", buf.String())
}
