package main

import (
	"fmt"
	"github.com/stevenayers/clamber/api"
	"github.com/stevenayers/clamber/service"
	"log"
	"net/http"
)

func main() {
	service.InitConfig()
	service.APILogger.InitJsonLogger(service.AppConfig.General.LogLevel)
	router := api.NewRouter()
	service.APILogger.LogInfo(
		"port", service.AppConfig.General.Port,
		"msg", "clamber api started listening",
	)
	err := http.ListenAndServe(fmt.Sprintf(":%d", service.AppConfig.General.Port), router)
	if err != nil {
		log.Fatal(err)
	}
}
