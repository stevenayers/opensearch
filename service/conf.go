package service

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
)

type (
	Config struct {
		General  GeneralConfig
		Database DatabaseConfig
	}

	GeneralConfig struct {
		MaxGoroutines int `toml:"max_goroutines"`
		Port          int
		LogLevel      string `toml:"log_level"`
	}
	DatabaseConfig struct {
		Connections []*Connections
	}
	Connections struct {
		Host string
		Port int
	}
)

var (
	// Bug with flag redefined in service package tests.
	//ConfigFile = flag.String("config", "../Config.toml", "Config file path")
	//Port       = flag.Int("port", 8002, "Port to listen on")
	//Verbose    = flag.Bool("verbose", false, "Verbosity")
	AppConfig Config
)

// Load config in from specified TOML file.
func InitConfig(configFile string) {
	tomlData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Could not read config file: %s - %s", configFile, err.Error())
	}
	if _, err := toml.Decode(string(tomlData), &AppConfig); err != nil {
		log.Fatalf("Could not parse TOML config: %s - %s", configFile, err.Error())
	}
}
