package message_broker

import (
	"fmt"
	"github.com/faramarzQ/sms-gateway-service/internals/value_objects"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeSMSDispatch = "sms.dispatch"
	ExchangeSMSAck      = "sms.ack"

	QueueStandardOrdinary = "sms.standard.ordinary"
	QueueStandardExpress  = "sms.standard.express"
	QueueBulkOrdinary     = "sms.bulk.ordinary"
	QueueBulkExpress      = "sms.bulk.express"

	QueueSMSAck = "sms.ack"
)

func Setup(ch *amqp.Channel) error {
	if err := declareSMSDispatchExchange(ch); err != nil {
		return err
	}

	if err := declareSMSAckExchange(ch); err != nil {
		return err
	}

	if err := declareDispatchQueues(ch); err != nil {
		return err
	}

	return declareAckQueue(ch)

}

func declareSMSDispatchExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		ExchangeSMSDispatch,
		"direct",
		true,  // durable
		false, // auto delete
		false, // internal
		false, // no wait
		nil,
	)
}

func declareSMSAckExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		ExchangeSMSAck,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
}

func declareDispatchQueues(ch *amqp.Channel) error {
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
			ExchangeSMSDispatch,
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func declareAckQueue(ch *amqp.Channel) error {
	_, err := ch.QueueDeclare(
		QueueSMSAck,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return ch.QueueBind(
		QueueSMSAck,
		"sms.ack",
		ExchangeSMSAck,
		false,
		nil,
	)
}
