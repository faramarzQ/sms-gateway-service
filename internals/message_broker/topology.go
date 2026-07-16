package message_broker

import (
	"fmt"
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeSMS = "sms.dispatch"

	QueueStandardOrdinary = "sms.standard.ordinary"
	QueueStandardExpress  = "sms.standard.express"
	QueueBulkOrdinary     = "sms.bulk.ordinary"
	QueueBulkExpress      = "sms.bulk.express"
)

func Setup(ch *amqp.Channel) error {
	if err := declareExchange(ch); err != nil {
		return err
	}

	return declareQueues(ch)
}

func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		ExchangeSMS,
		"direct",
		true,  // durable
		false, // auto delete
		false, // internal
		false, // no wait
		nil,
	)
}

func declareQueues(ch *amqp.Channel) error {
	queues := []struct {
		Name       string
		RoutingKey string
	}{
		{
			Name:       QueueStandardOrdinary,
			RoutingKey: string(value_objects.TrafficClassStandard),
		},
		{
			Name: QueueStandardExpress,
			RoutingKey: fmt.Sprintf("%s.%s",
				value_objects.TrafficClassStandard,
				value_objects.SMSTypeExpress,
			),
		},
		{
			Name:       QueueBulkOrdinary,
			RoutingKey: string(value_objects.TrafficClassBulk),
		},
		{
			Name: QueueBulkExpress,
			RoutingKey: fmt.Sprintf("%s.%s",
				value_objects.TrafficClassBulk,
				value_objects.SMSTypeExpress,
			),
		},
	}

	for _, q := range queues {
		_, err := ch.QueueDeclare(
			q.Name,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}

		err = ch.QueueBind(
			q.Name,
			q.RoutingKey,
			ExchangeSMS,
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
