package main

import (
	"clamber/api"
	"clamber/service"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config := service.GetConfig()
	router := api.NewRouter()
	log.Printf("Listening on port %d...", config.General.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.General.Port), router)
	if err != nil {
		log.Fatal(err)
	}
}
