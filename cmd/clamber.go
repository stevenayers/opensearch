package main

import (
	"github.com/stevenayers/clamber/api"
	"github.com/stevenayers/clamber/service"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	stdlog "log"
	"net/http"
	"os"
)

func main() {
	api.InitFlags(&api.AppFlags)
	appConfig, err := service.InitConfig(*api.AppFlags.ConfigFile)
	if *api.AppFlags.Port != 0 {
		appConfig.General.Port = *api.AppFlags.Port
	}
	if *api.AppFlags.Verbose {
		appConfig.General.LogLevel = "debug"
	}
	logger := api.InitJsonLogger(log.NewSyncWriter(os.Stdout), appConfig.General.LogLevel)
	if err != nil {
		stdlog.Fatal(err.Error())
		return
	}
	router := api.NewRouter()
	_ = level.Info(logger).Log(
		"port", appConfig.General.Port,
		"msg", "clamber api started listening",
	)
	err = http.ListenAndServe(fmt.Sprintf(":%d", appConfig.General.Port), router)
	if err != nil {
		_ = level.Error(logger).Log("msg", err.Error())
	}
}
