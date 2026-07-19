package main

import (
	"github.com/faramarzQ/sms-gateway-service/internals/app"
	"github.com/faramarzQ/sms-gateway-service/internals/config"
	"github.com/faramarzQ/sms-gateway-service/internals/database"
	"github.com/faramarzQ/sms-gateway-service/internals/logger"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db := database.ConnectPostgres(cfg.Postgres)

	rabbitmq := message_broker.ConnectRabbitMQ(cfg.RabbitMQ)

	err = logger.Init()
	if err != nil {
		log.Fatal(err)
	}

	application := app.NewApplicationContainer(app.AppMessageConsumer, db, rabbitmq, nil, nil)
	err = application.Build()
	if err != nil {
		log.Fatal("Error building application")
	}

	err = application.Run()
	if err != nil {
		log.Fatal("Error running application")
	}
}
