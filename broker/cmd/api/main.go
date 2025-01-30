package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

const ApiPort = "80"

type application struct {
}

func main() {
	app := application{}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", ApiPort),
		Handler: app.routes(),
	}

	slog.Info("Starting broker service", "port", ApiPort)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
