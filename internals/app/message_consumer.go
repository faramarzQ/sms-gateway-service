package app

import (
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/faramarzQ/sms-gateway-service/internals/services"
)

type MessageConsumer struct {
	app *App
}

func NewMessageConsumer(app *App) Application {
	return &MessageConsumer{
		app: app,
	}
}

func (a *MessageConsumer) Build() error {
	var err error
	a.app.MessageConsumer, err = message_broker.NewConsumer(a.app.rabbitMQ)
	if err != nil {
		return err
	}

	a.app.MessageConsumerService = services.NewMessageConsumerService(a.app.MessageConsumer, a.app.SMSRepository)

	return nil
}

func (a *MessageConsumer) Run() error {
	return a.app.MessageConsumerService.Consume()
}
