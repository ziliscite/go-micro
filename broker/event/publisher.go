package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
}

func NewPublisher(conn *amqp.Connection) (*Publisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer channel.Close()

	err = declareExchange(channel)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn: conn,
	}, nil
}

func (p *Publisher) Push(event, severity string) error {
	channel, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	
	return channel.Publish(
		"logs_topic", // the same exchange in the consumer
		severity,     // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	)
}
