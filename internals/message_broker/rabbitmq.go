package message_broker

import (
	"github.com/faramarzQ/sms-gateway-service/internals/config"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ConnectRabbitMQ(cfg config.RabbitMQConfig) *amqp.Connection {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}
