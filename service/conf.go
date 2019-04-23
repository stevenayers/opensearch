package service

import (
	"flag"
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
	ConfigFile = flag.String("config", "../Config.toml", "Config file path")
	Port       = flag.Int("port", 8000, "Port to listen on")
	Verbose    = flag.Bool("verbose", false, "Verbosity")
)

func GetConfig() (conf Config) {
	flag.Parse()
	tomlData, err := ioutil.ReadFile(*ConfigFile)
	if err != nil {
		log.Fatalf("Could not read config file: %s - %s", *ConfigFile, err.Error())
	}
	if _, err := toml.Decode(string(tomlData), &conf); err != nil {
		log.Fatalf("Could not parse TOML config: %s - %s", *ConfigFile, err.Error())
	}
	return
}
