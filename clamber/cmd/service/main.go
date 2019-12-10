package main

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/stevenayers/clamber/pkg/config"
	"github.com/stevenayers/clamber/pkg/crawl"
	"github.com/stevenayers/clamber/pkg/logging"
	"github.com/stevenayers/clamber/pkg/route"
	stdlog "log"
	"net/http"
	"os"
)

func main() {
	InitFlags(&AppFlags)
	err := config.InitConfig(*AppFlags.ConfigFile)
	if *AppFlags.Port != 0 {
		config.AppConfig.Service.Port = *AppFlags.Port
	}
	if *AppFlags.Verbose {
		config.AppConfig.Service.LogLevel = "debug"
	}
	logging.InitJsonLogger(log.NewSyncWriter(os.Stdout), config.AppConfig.Service.LogLevel, "service")
	if err != nil {
		stdlog.Fatal(err.Error())
		return
	}
	crawler := crawl.New()
	go crawler.Start()
	router := route.NewRouter(Routes)
	_ = level.Info(logging.Logger).Log(
		"port", config.AppConfig.Service.Port,
		"msg", "clamber service started",
	)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.AppConfig.Service.Port), router)
	if err != nil {
		_ = level.Error(logging.Logger).Log("msg", err.Error())
	}

}
