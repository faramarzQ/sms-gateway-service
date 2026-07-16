package message_broker

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ConnectRabbitMQ() *amqp.Connection {
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}

	return conn
}
