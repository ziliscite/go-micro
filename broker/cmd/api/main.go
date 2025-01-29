package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
)

const PORT = "80"

type application struct {
}

func main() {
	app := application{}

	slog.Info("Starting broker service", "port", PORT)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: app.routes(),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
