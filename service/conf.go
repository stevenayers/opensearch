package service

import (
	"flag"
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
		MaxGoroutines int `toml:"max_goroutines"`
		Port          int
		LogLevel      string `toml:"log_level"`
	}

	// DatabaseConfig holds database section of toml config
	DatabaseConfig struct {
		Connections []*Connections
	}

	// Connections holds the database connection data
	Connections struct {
		Host string
		Port int
	}

	// Flags holds the app Flags
	Flags struct {
		ConfigFile *string
		Port       *int
		Verbose    *bool
	}
)

var (
	// AppFlags makes a global Flag struct
	AppFlags Flags
	// AppConfig makes a global Config struct
	AppConfig Config
)

// InitConfig loads config in from specified TOML file.
func InitConfig() {
	AppFlags.ConfigFile = flag.String("config", "../cmd/Config.toml", "Config file path")
	AppFlags.Port = flag.Int("port", 8002, "Port to listen on")
	AppFlags.Verbose = flag.Bool("verbose", false, "Verbosity")
	flag.Parse()
	tomlData, err := ioutil.ReadFile(*AppFlags.ConfigFile)
	if err != nil {
		log.Fatalf("Could not read config file: %s - %s", *AppFlags.ConfigFile, err.Error())
	}
	if _, err := toml.Decode(string(tomlData), &AppConfig); err != nil {
		log.Fatalf("Could not parse TOML config: %s - %s", *AppFlags.ConfigFile, err.Error())
	}
}
