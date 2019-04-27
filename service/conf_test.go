package service_test

import (
	"github.com/stevenayers/clamber/service"
	"github.com/stretchr/testify/assert"
	"strings"
)

func (s *StoreSuite) TestConfigPath() {
	configFile := "../test/incorrectpath.toml"
	_, err := service.InitConfig(configFile)
	assert.Equal(s.T(), true, err != nil)
	if err != nil {
		assert.Equal(s.T(), true, strings.Contains(
			err.Error(), "no such file or directory"))
	}
}

func (s *StoreSuite) TestConfigParse() {
	configFile := "../test/badconfig.toml"
	_, err := service.InitConfig(configFile)
	assert.Equal(s.T(), true, err != nil)
	if err != nil {
		assert.Equal(s.T(), true, strings.Contains(
			err.Error(), "cannot load TOML value of type string into a Go integer"))
	}
}
