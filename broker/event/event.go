package event

import amqp "github.com/rabbitmq/amqp091-go"

// Declare exchange for pub/sub
func declareExchange(channel *amqp.Channel) error {
	// Declare an exchange
	return channel.ExchangeDeclare(
		"logs_topic", // name of the exchange
		"topic",      // type
		true,         // durable? yes, it supposed to last
		false,        // auto-deleted
		false,        // internal, no, its gonna be used between microservice
		false,        // no-wait
		nil,          // arguments
	)
}

// Declare a random queue
func declareQueue(channel *amqp.Channel) (amqp.Queue, error) {
	return channel.QueueDeclare(
		"",    // name of the queue
		false, // durable
		false, // delete when we're done, not auto
		true,  // exclusive channel for current operation, don't share it
		false, // no-wait
		nil,   // arguments
	)
}

func bindQueueToExchange(channel *amqp.Channel, queueName, key string) error {
	return channel.QueueBind(
		queueName,    // name of the queue
		key,          // routing key
		"logs_topic", // exchange
		false,        // no-wait
		nil,          // arguments
	)
}
