package main

import (
	"github.com/ziliscite/go-micro-listener/event"
	"log/slog"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// Connect to rabbitmq
	conn, err := connectAMQP()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// Create consumer
	consumer, err := event.NewConsumer(conn)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// Watch queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARN", "log.ERROR"})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
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
