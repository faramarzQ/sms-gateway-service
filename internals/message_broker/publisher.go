package message_broker

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const SMSExchange = "sms.dispatch"

type Publisher struct {
	channel *amqp.Channel
}

func NewPublisher(conn *amqp.Connection) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := Setup(ch); err != nil {
		log.Fatal(err)
	}

	return &Publisher{
		channel: ch,
	}, nil
}

func (p *Publisher) Publish(
	ctx context.Context,
	routingKey string,
	body []byte,
) error {

	return p.channel.PublishWithContext(
		ctx,
		SMSExchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
