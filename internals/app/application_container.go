package app

import (
	"github.com/faramarzQ/sms-gateway-service/internals/http/handlers"
	"github.com/faramarzQ/sms-gateway-service/internals/message_broker"
	"github.com/faramarzQ/sms-gateway-service/internals/middlewares"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	"github.com/faramarzQ/sms-gateway-service/internals/services"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	AppAPI               string = "API"
	AppTrafficClassifier string = "Traffic_Classifier"
	AppMessageConsumer   string = "Message_Consumer"
)

var RegisteredApplications = map[string]ApplicationFactory{
	AppAPI:               NewAPI,
	AppTrafficClassifier: NewTrafficClassifier,
	AppMessageConsumer:   NewMessageConsumer,
}

type ApplicationContainer struct {
	Name string

	RunnableApplication Application

	db       *gorm.DB
	rabbitMQ *amqp.Connection
	Router   *gin.Engine
	Redis    *redis.Client

	MessagePublisher *message_broker.Publisher
	MessageConsumer  *message_broker.Consumer

	UserRepository repositories.UserRepositoryInterface
	SMSRepository  repositories.SMSRepositoryInterface

	UserService              *services.UserService
	SMSService               *services.SMSService
	TrafficClassifierService *services.TrafficClassifierService
	MessageConsumerService   *services.MessageConsumerService

	UserHandler *handlers.UserHandler
	SMSHandler  *handlers.SMSHandler

	RateLimiterMiddleware *middlewares.RateLimitMiddleware
}

func NewApplicationContainer(name string, db *gorm.DB, rabbitMQ *amqp.Connection, redis *redis.Client, router *gin.Engine) *ApplicationContainer {
	return &ApplicationContainer{
		Name:     name,
		db:       db,
		rabbitMQ: rabbitMQ,
		Router:   router,
		Redis:    redis,
	}
}

func (c *ApplicationContainer) Build() error {
	err := c.BuildDependencies()
	if err != nil {
		return err
	}

	c.RunnableApplication = RegisteredApplications[c.Name](c)

	err = c.RunnableApplication.Build()
	if err != nil {
		return err
	}

	return nil
}

func (c *ApplicationContainer) BuildDependencies() error {
	var err error
	c.MessagePublisher, err = message_broker.NewPublisher(c.rabbitMQ)
	if err != nil {
		return err
	}

	c.UserRepository = repositories.NewUserRepository(c.db)
	c.SMSRepository = repositories.NewSMSRepository(c.db)

	c.UserService = services.NewUserService(c.UserRepository, c.Redis)
	c.SMSService = services.NewSMSService(c.SMSRepository, c.MessagePublisher, c.UserService, c.UserRepository, c.Redis)

	return nil
}

func (c *ApplicationContainer) Run() error {
	return c.RunnableApplication.Run()
}
