package main

import (
	"database/sql"
	"fmt"
	"github.com/ziliscite/go-micro-authentication/internal/repository"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type config struct {
	port string
	dsn  string
}

type application struct {
	cfg  config
	repo repository.Repository
}

func main() {
	cfg := config{
		port: "80",
		dsn:  os.Getenv("DB_DSN"),
	}

	db, err := openDB(cfg.dsn)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	repo := repository.New(db)

	app := application{
		cfg:  cfg,
		repo: repo,
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.port),
		Handler: app.routes(),
	}

	slog.Info("Starting authentication service", "port", cfg.port)
	if err = server.ListenAndServe(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(2 * time.Hour)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	slog.Info("Connected to authentication database")
	return db, nil
}
