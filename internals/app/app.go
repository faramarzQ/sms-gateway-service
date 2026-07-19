package app

import (
	"github.com/faramarzQ/sms-gateway-service/internals/http/handlers"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	"github.com/faramarzQ/sms-gateway-service/internals/services"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	AppAPI               string = "API"
	AppTrafficClassifier string = "Traffic Classifier"
	AppMessageConsumer   string = "Message Consumer"
)

var RegisteredApplications = map[string]ApplicationFactory{
	AppAPI:               NewAPI,
	AppTrafficClassifier: NewTrafficClassifier,
	AppMessageConsumer:   NewMessageConsumer,
}

type App struct {
	Name string

	RunnableApplication Application

	db       *gorm.DB
	rabbitMQ *amqp.Connection
	Router   *gin.Engine
	Redis    *redis.Client

	MessagePublisher *message_broker.Publisher
	MessageConsumer  *message_broker.Consumer

	UserRepository *repositories.UserRepository
	SMSRepository  *repositories.SMSRepository

	UserService              *services.UserService
	SMSService               *services.SMSService
	TrafficClassifierService *services.TrafficClassifierService
	MessageConsumerService   *services.MessageConsumerService

	UserHandler *handlers.UserHandler
	SMSHandler  *handlers.SMSHandler
}

func NewApp(name string, db *gorm.DB, rabbitMQ *amqp.Connection, redis *redis.Client, router *gin.Engine) *App {
	return &App{
		Name:     name,
		db:       db,
		rabbitMQ: rabbitMQ,
		Router:   router,
		Redis:    redis,
	}
}

func (app *App) Build() error {
	err := app.BuildDependencies()
	if err != nil {
		return err
	}

	app.RunnableApplication = RegisteredApplications[app.Name](app)

	err = app.RunnableApplication.Build()
	if err != nil {
		return err
	}

	return nil
}

func (app *App) BuildDependencies() error {
	var err error
	app.MessagePublisher, err = message_broker.NewPublisher(app.rabbitMQ)
	if err != nil {
		return err
	}

	app.UserRepository = repositories.NewUserRepository(app.db)
	app.SMSRepository = repositories.NewSMSRepository(app.db)

	app.UserService = services.NewUserService(app.UserRepository, app.Redis)
	app.SMSService = services.NewSMSService(app.SMSRepository, app.MessagePublisher, app.UserService, app.Redis)

	return nil
}

func (app *App) Run() error {
	return app.RunnableApplication.Run()
}
