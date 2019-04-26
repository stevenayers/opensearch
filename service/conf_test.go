package service_test

import (
	"github.com/stevenayers/clamber/service"
	"github.com/stretchr/testify/assert"
	"strings"
)

func (s *StoreSuite) TestConfigPath() {
	*service.AppFlags.ConfigFile = "../test/incorrectpath.toml"
	err := service.InitConfig()
	assert.Equal(s.T(), true, err != nil)
	if err != nil {
		assert.Equal(s.T(), true, strings.Contains(
			err.Error(), "no such file or directory"))
	}
}

func (s *StoreSuite) TestConfigParse() {
	*service.AppFlags.ConfigFile = "../test/badconfig.toml"
	err := service.InitConfig()
	assert.Equal(s.T(), true, err != nil)
	if err != nil {
		assert.Equal(s.T(), true, strings.Contains(
			err.Error(), "cannot load TOML value of type string into a Go integer"))
	}
}
