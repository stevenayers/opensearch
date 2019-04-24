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
	config := service.GetConfig()
	logging.InitJsonLogger(config.General.LogLevel)
	router := api.NewRouter()
	logging.LogInfo(
		"port", config.General.Port,
		"msg", "clamber api started listening",
		"config", *service.ConfigFile,
	)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.General.Port), router)
	if err != nil {
		log.Fatal(err)
	}
}
