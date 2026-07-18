package app

import "github.com/faramarzQ/sms-gateway-service/internals/services"

type TrafficClassifier struct {
	app *App
}

func NewTrafficClassifier(app *App) Application {
	return &TrafficClassifier{
		app: app,
	}
}

func (a *TrafficClassifier) Build() error {
	a.app.TrafficClassifierService = services.NewTrafficClassifierService(
		a.app.Redis,
		a.app.UserRepository,
	)

	return nil
}

func (a *TrafficClassifier) Run() error {
	return a.app.TrafficClassifierService.Execute()
}
