package main

import (
	"github.com/go-chi/chi/v5/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
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

	mux.Post("/", app.broker)
	mux.Post("/handle", app.gateway)

	mux.Post("/log/grpc", app.logGRPC)

	return middleware.Recoverer(mux)
}
