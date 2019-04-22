package main

import (
	"clamber/server"
	"clamber/utils"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config := utils.GetConfig()
	if config.General.MaxGoroutines > 0 {
		utils.InitBuffer(config.General.MaxGoroutines)
	}

	router := server.NewRouter()
	log.Printf("Listening on port %d...", config.General.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.General.Port), router)
	if err != nil {
		log.Fatal(err)
	}
}
