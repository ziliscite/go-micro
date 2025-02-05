package event

import (
	"bytes"
	"encoding/json"
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"net/http"
)

// Consumer receives events
type Consumer struct {
	conn *amqp.Connection
	qn   string // queue name
}

func NewConsumer(conn *amqp.Connection) (*Consumer, error) {
	c := &Consumer{
		conn: conn,
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer channel.Close()

	err = declareExchange(channel)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type payload struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (c *Consumer) Listen(topics []string) error {
	// get a channel
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// get a queue
	q, err := declareQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		// bind channels to each of these topics
		// channel bind s to a queue
		if err = bindQueueToExchange(ch, q.Name, s); err != nil {
			return err
		}
	}

	// look for messages
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	// consume til application exits
	forever := make(chan bool)
	go func() {
		for m := range msgs {
			var p payload
			// encode payload
			_ = json.Unmarshal(m.Body, &p)

			go handlePayload(p)
		}
	}()

	slog.Info("Listening for events [Exchange, Queue]", "logs_topic", q.Name)
	<-forever

	return nil
}

func handlePayload(p payload) {
	switch p.Title {
	//case "log", "event":
	// when queue is a log or event
	case "auth":
		// some auth logic when something happens
	default:
		err := logEvent(p)
		if err != nil {
			slog.Error("Failed to log event", "error", err)
		}
	}
}

// stub handler -- log event when receive some from rabbitmq
func logEvent(entry payload) error {
	// Create the payload
	p, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Now http, later rpc
	req, err := http.NewRequest(http.MethodPost, "http://logger/v1/logs", bytes.NewBuffer(p))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var message string
	switch resp.StatusCode {
	case http.StatusConflict:
		message = "a conflict occurred"
	case http.StatusBadRequest:
		message = "invalid log data"
	case http.StatusGatewayTimeout:
		message = "gateway timeout"
	case http.StatusNotFound:
		message = "resource not found"
	default:
		message = "log service could not process your request"
	}

	// Check the status code
	if resp.StatusCode != http.StatusAccepted {
		return errors.New(message)
	}

	slog.Info("Event logged", "content", resp.Body)
	return nil
}
