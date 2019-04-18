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

	fmt.Print(config)
	router := routes.NewRouter()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.General.Port), router))
}
