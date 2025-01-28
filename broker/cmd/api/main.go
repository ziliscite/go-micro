package main

import (
	"fmt"
	"log"
	"net/http"
)

const PORT = "80"

type Application struct {
}

func main() {
	app := Application{}

	log.Printf("Starting broker service on port %s\n", PORT)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: app.routes(),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
