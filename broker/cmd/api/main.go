package main

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const ApiPort = "80"

type application struct {
	rabbit *amqp.Connection
}

func newApplication(conn *amqp.Connection) application {
	return application{
		rabbit: conn,
	}
}

func main() {
	// Connect to rabbitmq
	conn, err := connectAMQP()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	app := newApplication(conn)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", ApiPort),
		Handler: app.routes(),
	}

	slog.Info("Starting broker service", "port", ApiPort)
	if err = server.ListenAndServe(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	// Push things to rabbitmq queue so that listener service can get that event
}

func connectAMQP() (*amqp.Connection, error) {
	counts := 0
	backOff := 1 * time.Second
	var conn *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			slog.Error("Rabbitmq not ready", "error", err)
			counts++
		} else {
			conn = c
			break
		}

		if counts > 5 {
			slog.Error("Failed to connect to rabbitmq", "error", err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		time.Sleep(backOff)
	}

	slog.Info("Connected to rabbitmq")
	return conn, nil
}
