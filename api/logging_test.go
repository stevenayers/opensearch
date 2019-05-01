package api_test

import (
	"bytes"
	"github.com/stevenayers/clamber/api"
	"encoding/json"
	"github.com/go-kit/kit/log/level"
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
	var logOutput LogOutput
	logger := api.InitJsonLogger(buf, "debug")
	err := level.Debug(logger).Log("msg", "test")
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
	logger := api.InitJsonLogger(buf, "info")
	err := level.Info(logger).Log("msg", "test")
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
	logger := api.InitJsonLogger(buf, "error")
	err := level.Error(logger).Log("msg", "test")
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
	logger := api.InitJsonLogger(buf, "defaultShouldGoToInfo")
	err := level.Debug(logger).Log("msg", "test")
	if err != nil {
		s.T().Fatal(err)
	}
	assert.Equal(s.T(), "", buf.String())
}
