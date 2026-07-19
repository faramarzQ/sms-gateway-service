// @title           SMS Gateway Service API
// @version         1.0
// @description     REST API for sending SMS, managing user balance, and
package main

import (
	"github.com/faramarzQ/sms-gateway-service/internals/app"
	"github.com/faramarzQ/sms-gateway-service/internals/cache"
eway-service/docs"
	"github.com/faramarzQ/sms-gateway-servic
	"github.com/faramarzQ/sms-gateway-service/internals/config"
	"github.com/faramarzQ/sms-gateway-service/internals/database"
	"github.com/faramarzQ/sms-gateway-service/internals/logger"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db := database.ConnectPostgres(cfg.Postgres)

	rabbitmq := message_broker.ConnectRabbitMQ(cfg.RabbitMQ)

	redis := cache.ConnectRedis(cfg.Redis)

	err = logger.Init()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	application := app.NewApp(app.AppAPI, db, rabbitmq, redis, router)
	err = application.Build()
	if err != nil {
		log.Fatal("Error building application")
	}

	err = application.Run()
	if err != nil {
		log.Fatal("Error running application")
	}

}
