package main

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/stevenayers/opensearch/pkg/config"
	"github.com/stevenayers/opensearch/pkg/logging"
	"github.com/stevenayers/opensearch/pkg/route"
	stdlog "log"
	"net/http"
	"os"
)

func main() {
	InitFlags(&AppFlags)
	err := config.InitConfig(*AppFlags.ConfigFile)
	if *AppFlags.Port != 0 {
		config.AppConfig.Api.Port = *AppFlags.Port
	}
	if *AppFlags.Verbose {
		config.AppConfig.Api.LogLevel = "debug"
	}
	logging.InitJsonLogger(log.NewSyncWriter(os.Stdout), config.AppConfig.Api.LogLevel, "api")
	if err != nil {
		stdlog.Fatal(err.Error())
		return
	}
	router := route.NewRouter(Routes)
	_ = level.Info(logging.Logger).Log(
		"port", config.AppConfig.Api.Port,
		"msg", "opensearch api started",
	)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.AppConfig.Api.Port), router)
	if err != nil {
		_ = level.Error(logging.Logger).Log("msg", err.Error())
	}
}
