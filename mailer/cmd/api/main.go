package main

import (
	"log/slog"
	"os"
)

const ApiPort = "80"

type application struct {
	mailer *Mailer
}

func main() {
	mailer, err := NewMailer()
	if err != nil {
		slog.Error("Failed to get mailer server", "error", err)
		os.Exit(1)
	}

	app := application{
		mailer: mailer,
	}

	if err = app.serve(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
