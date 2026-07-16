package app

import (
	"github.com/faramarzQ/sms-gateway-service/internals/http"
	"github.com/faramarzQ/sms-gateway-service/internals/http/handlers"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	"github.com/faramarzQ/sms-gateway-service/internals/services"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	db       *gorm.DB
	rabbitMQ *amqp.Connection
	router   *gin.Engine
	redis    *redis.Client
}

func NewApp(db *gorm.DB, rabbitMQ *amqp.Connection, redis *redis.Client, router *gin.Engine) *App {
	return &App{
		db:       db,
		rabbitMQ: rabbitMQ,
		router:   router,
		redis:    redis,
	}
}

func (app *App) Build() error {

	userRepository := repositories.NewUserRepository(app.db)
	userService := services.NewUserService(userRepository, app.redis)
	userHandler := handlers.NewUserHandler(userService)

	smsRepository := repositories.NewSMSRepository(app.db)

	messagePublisher, err := message_broker.NewPublisher(app.rabbitMQ)
	if err != nil {
		return err
	}

	smsService := services.NewSMSService(smsRepository, messagePublisher, userService)
	smsHandler := handlers.NewSMSHandler(smsService)

	http.RegisterRoutes(app.router, userHandler, smsHandler)

	return nil
}
