package app

import "github.com/faramarzQ/sms-gateway-service/internals/services"

type TrafficClassifier struct {
	container *ApplicationContainer
}

func NewTrafficClassifier(app *ApplicationContainer) Application {
	return &TrafficClassifier{
		container: app,
	}
}

func (app *TrafficClassifier) Build() error {
	app.container.TrafficClassifierService = services.NewTrafficClassifierService(
		app.container.Redis,
		app.container.UserRepository,
	)

	return nil
}

func (app *TrafficClassifier) Run() error {
	return app.container.TrafficClassifierService.Execute()
}
