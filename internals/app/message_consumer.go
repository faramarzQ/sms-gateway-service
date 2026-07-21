package app

import (
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/faramarzQ/sms-gateway-service/internals/services"
	"net/http"
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

	go startHealthServer()

	return app.container.MessageConsumerService.Consume()
}

func startHealthServer() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_ = http.ListenAndServe(":8080", nil)
}
