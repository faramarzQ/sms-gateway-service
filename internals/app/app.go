package app

import (
	"github.com/faramarzQ/sms-gateway-service/internals/http"
	"github.com/faramarzQ/sms-gateway-service/internals/http/handlers"
	"github.com/faramarzQ/sms-gateway-service/internals/repositories"
	"github.com/faramarzQ/sms-gateway-service/internals/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	db     *gorm.DB
	router *gin.Engine
}

func NewApp(db *gorm.DB, router *gin.Engine) *App {
	return &App{
		db:     db,
		router: router,
	}
}

func (app *App) Build() {

	userRepository := repositories.NewUserRepository(app.db)
	userService := services.NewUserService(userRepository)
	userHandler := handlers.NewUserHandler(userService)

	http.RegisterRoutes(app.router, userHandler)

}
