package main

import (
	"clamber/conf"
	"clamber/routes"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config := conf.GetConfig()

	router := routes.NewRouter()
	log.Printf("Listening on port %d...", config.General.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.General.Port), router)
	if err != nil {
		log.Fatal(err)
	}
}
