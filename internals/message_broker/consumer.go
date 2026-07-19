package message_broker

import amqp "github.com/rabbitmq/amqp091-go"

type Consumer struct {
	channel *amqp.Channel
}

func NewConsumer(conn *amqp.Connection) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Consumer{
		channel: ch,
	}, nil
}

func (c *Consumer) Consume(queue string) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		queue,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
}
