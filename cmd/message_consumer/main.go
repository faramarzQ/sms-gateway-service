package main

import (
	"github.com/faramarzQ/sms-gateway-service/internals/app"
	"github.com/faramarzQ/sms-gateway-service/internals/database"
	"github.com/faramarzQ/sms-gateway-service/internals/logger"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"log"
)

func main() {
	db := database.ConnectPostgres()

	rabbitmq := message_broker.ConnectRabbitMQ()

	err := logger.Init()
	if err != nil {
		log.Fatal(err)
	}

	application := app.NewApp(app.AppMessageConsumer, db, rabbitmq, nil, nil)
	err = application.Build()
	if err != nil {
		log.Fatal("Error building application")
	}

	err = application.Run()
	if err != nil {
		log.Fatal("Error running application")
	}
}
