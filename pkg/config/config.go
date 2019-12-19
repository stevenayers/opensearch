package config

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
)

type (

	// Config holds Service and Database config from TOML file
	Config struct {
		Api      ApiConfig
		Service  ServiceConfig
		Database DatabaseConfig
		Queue    QueueConfig
	}

	// GeneralConfig holds general section of toml config
	ApiConfig struct {
		MaxGoroutines int `toml:"max_goroutines"`
		Port          int
		LogLevel      string `toml:"log_level"`
		WaitCrawl     bool   `toml:"wait_crawl"`
	}

	// GeneralConfig holds general section of toml config
	ServiceConfig struct {
		MaxGoroutines       int `toml:"max_goroutines"`
		Port                int
		LogLevel            string `toml:"log_level"`
		HttpRetryAttempts   int    `toml:"http_retry_attempts"`
		HttpBackOffDuration int    `toml:"http_back_off_duration"`
		NumConsumers        int    `toml:"sqs_consumers_per_node"`
	}

	// DatabaseConfig holds database section of toml config
	DatabaseConfig struct {
		Connections []*Connection
	}

	QueueConfig struct {
		QueueURL                      string `toml:"queue_url"`
		QueueName                     string `toml:"queue_name"`
		AwsRegion                     string `toml:"aws_region"`
		MaxConcurrentReceivedMessages int64  `toml:"max_concurrent_received_messages"`
		SQSWaitTimeSeconds            int64  `toml:"sqs_wait_time_seconds"`
	}

	// Connection holds the database connection data
	Connection struct {
		Host string
		Port int
	}
)

var AppConfig Config

// InitConfig loads config in from specified TOML file.
func InitConfig(path string) (err error) {
	tomlData, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Could not read config file: %s - %s", path, err.Error())
		return
	}
	_, err = toml.Decode(string(tomlData), &AppConfig)
	if err != nil {
		log.Printf("Could not parse TOML config: %s - %s", path, err.Error())
	}
	return
}
