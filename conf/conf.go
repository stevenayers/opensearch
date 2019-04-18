package conf

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
	config = flag.String("config", "Config.toml", "Config file path")
)

func GetConfig() (conf Config) {
	flag.Parse()
	tomlData, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatalf("Could not read config file: %s - %s", *config, err.Error())
	}
	if _, err := toml.Decode(string(tomlData), &conf); err != nil {
		log.Fatalf("Could not parse TOML config: %s - %s", *config, err.Error())
	}
	return
}
