package app

import (
	"github.com/faramarzQ/sms-gateway-service/internals/http"
	"github.com/faramarzQ/sms-gateway-service/internals/http/handlers"
)

type API struct {
	app *App
}

func NewAPI(app *App) Application {
	return &API{
		app: app,
	}
}

func (a *API) Build() error {
	a.app.UserHandler = handlers.NewUserHandler(a.app.UserService)
	a.app.SMSHandler = handlers.NewSMSHandler(a.app.SMSService)

	http.RegisterRoutes(
		a.app.Router,
		a.app.UserHandler,
		a.app.SMSHandler,
	)

	return nil
}

func (a *API) Run() error {
	return a.app.Router.Run(":8080")

}
