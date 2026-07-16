package main

import (
	"github.com/faramarzQ/sms-gateway-service/internals/app"
	"github.com/faramarzQ/sms-gateway-service/internals/cache"
	"github.com/faramarzQ/sms-gateway-service/internals/database"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	db := database.ConnectPostgres()

	rabbitmq := message_broker.ConnectRabbitMQ()

	redis := cache.ConnectRedis()

	router := gin.Default()

	application := app.NewApp(db, rabbitmq, redis, router)
	err := application.Build()
	if err != nil {
		log.Fatal("Error building application")
	}

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
