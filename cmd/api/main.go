package main

import (
	_ "github.com/faramarzQ/sms-gateway-service/docs"
	"github.com/faramarzQ/sms-gateway-service/internals/app"
	"github.com/faramarzQ/sms-gateway-service/internals/cache"
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

	err = database.RunMigrations(cfg.Postgres)
	if err != nil {
		log.Fatal(err)
	}

	rabbitmq := message_broker.ConnectRabbitMQ(cfg.RabbitMQ)

	redis := cache.ConnectRedis(cfg.Redis)

	err = logger.Init()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	application := app.NewApplicationContainer(app.AppAPI, db, rabbitmq, redis, router)
	err = application.Build()
	if err != nil {
		log.Fatal("Error building application")
	}

	err = application.Run()
	if err != nil {
		log.Fatal("Error running application")
	}
}
