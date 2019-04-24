package main

import (
	"fmt"
	"github.com/stevenayers/clamber/api"
	"github.com/stevenayers/clamber/logging"
	"github.com/stevenayers/clamber/service"
	"log"
	"net/http"
)

func main() {
	tempConfigFile := "./Config.toml"
	service.InitConfig(tempConfigFile)
	logging.InitJsonLogger(service.AppConfig.General.LogLevel)
	router := api.NewRouter()
	logging.LogInfo(
		"port", service.AppConfig.General.Port,
		"msg", "clamber api started listening",
		"config", tempConfigFile,
	)
	err := http.ListenAndServe(fmt.Sprintf(":%d", service.AppConfig.General.Port), router)
	if err != nil {
		log.Fatal(err)
	}
}
