package http

import (
	"github.com/faramarzQ/sms-gateway-service/internals/http/handlers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(ginRouter *gin.Engine,
	userHandler *handlers.UserHandler,
	smsHandler *handlers.SMSHandler,
) {
	ginRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	users := ginRouter.Group("/user")
	{
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id/balance", userHandler.IncreaseBalance)
	}

	sms := ginRouter.Group("/sms")
	{
		sms.POST("/", smsHandler.SendSMS)
		sms.POST("/batch", smsHandler.SendSMSBatch)
		sms.GET("/report", smsHandler.GetReport)
	}
}
