package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	// specify who is allowed to connect
	mux.Use(
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
		middleware.Heartbeat("/ping"),
	)

	mux.Route("/v1", func(v1 chi.Router) {
		v1.Post("/logs", app.writeLog)
		v1.Get("/logs", app.listLogs)
	})

	return mux
}

func (app *application) serve() error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", ApiPort),
		Handler: app.routes(),
	}

	slog.Info("Starting logger service", "port", ApiPort)
	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
