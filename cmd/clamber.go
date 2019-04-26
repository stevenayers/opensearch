package main

import (
	"fmt"
	kitlog "github.com/go-kit/kit/log"
	"github.com/stevenayers/clamber/api"
	"github.com/stevenayers/clamber/service"
	"log"
	"net/http"
	"os"
)

func main() {
	service.InitFlags()
	err := service.InitConfig()
	if err != nil {
		log.Fatal(err)
		return
	}
	service.APILogger.InitJsonLogger(kitlog.NewSyncWriter(os.Stdout), service.AppConfig.General.LogLevel)
	router := api.NewRouter()
	service.APILogger.LogInfo(
		"port", service.AppConfig.General.Port,
		"msg", "clamber api started listening",
	)
	err = http.ListenAndServe(fmt.Sprintf(":%d", service.AppConfig.General.Port), router)
	if err != nil {
		log.Fatal(err)
	}
}
