package app

import (
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/faramarzQ/sms-gateway-service/internals/services"
)

type MessageConsumer struct {
	container *ApplicationContainer
}

func NewMessageConsumer(app *ApplicationContainer) Application {
	return &MessageConsumer{
		container: app,
	}
}

func (app *MessageConsumer) Build() error {
	var err error
	app.container.MessageConsumer, err = message_broker.NewConsumer(app.container.rabbitMQ)
	if err != nil {
		return err
	}

	app.container.MessageConsumerService = services.NewMessageConsumerService(app.container.MessageConsumer, app.container.SMSRepository)

	return nil
}

func (app *MessageConsumer) Run() error {
	return app.container.MessageConsumerService.Consume()
}
