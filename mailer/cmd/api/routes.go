package main

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}),
		middleware.Heartbeat("/ping"),
	)

	mux.Post("/v1/send", app.send)

	return middleware.Recoverer(mux)
}

func (app *application) serve() error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", ApiPort),
		Handler: app.routes(),
	}

	slog.Info("Starting mailer service", "port", ApiPort)
	return server.ListenAndServe()
}
