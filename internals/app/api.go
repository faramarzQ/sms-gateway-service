package app

import (
	"github.com/faramarzQ/sms-gateway-service/internals/http"
	"github.com/faramarzQ/sms-gateway-service/internals/http/handlers"
	"github.com/faramarzQ/sms-gateway-service/internals/middlewares"
)

type API struct {
	container *ApplicationContainer
}

func NewAPI(app *ApplicationContainer) Application {
	return &API{
		container: app,
	}
}

func (app *API) Build() error {
	app.container.UserHandler = handlers.NewUserHandler(app.container.UserService)
	app.container.SMSHandler = handlers.NewSMSHandler(app.container.SMSService)

	app.registerMiddlewares()

	http.RegisterRoutes(
		app.container.Router,
		app.container.UserHandler,
		app.container.SMSHandler,
		app.container.RateLimiterMiddleware,
	)

	return nil
}

func (app *API) Run() error {
	return app.container.Router.Run(":8080")

}

func (app *API) registerMiddlewares() {
	app.container.RateLimiterMiddleware = middlewares.NewRateLimitMiddleware(app.container.Redis)
}
