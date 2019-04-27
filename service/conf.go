package service

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
)

type (

	// Config holds General and Database config from TOML file
	Config struct {
		General  GeneralConfig
		Database DatabaseConfig
	}

	// GeneralConfig holds general section of toml config
	GeneralConfig struct {
		MaxGoroutines       int `toml:"max_goroutines"`
		Port                int
		LogLevel            string `toml:"log_level"`
		HttpRetryAttempts   int    `toml:"http_retry_attempts"`
		HttpBackOffDuration int    `toml:"http_back_off_duration"`
	}

	// DatabaseConfig holds database section of toml config
	DatabaseConfig struct {
		Connections []*Connection
	}

	// Connection holds the database connection data
	Connection struct {
		Host string
		Port int
	}
)

// InitConfig loads config in from specified TOML file.
func InitConfig(path string) (appConfig Config, err error) {
	tomlData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Could not read config file: %s - %s", path, err.Error())
		return
	}
	_, err = toml.Decode(string(tomlData), &appConfig)
	if err != nil {
		log.Printf("Could not parse TOML config: %s - %s", path, err.Error())
	}
	return
}
