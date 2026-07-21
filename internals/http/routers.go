package http

import (
	"github.com/faramarzQ/sms-gateway-service/internals/http/handlers"
	"github.com/faramarzQ/sms-gateway-service/internals/middlewares"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(ginRouter *gin.Engine,
	userHandler *handlers.UserHandler,
	smsHandler *handlers.SMSHandler,
	rateLimiterMiddleware *middlewares.RateLimitMiddleware,
) {
	ginRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	ginRouter.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	api := ginRouter.Group("/api")

	api.Use(rateLimiterMiddleware.Handler())

	users := api.Group("/user")
	{
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id/balance", userHandler.IncreaseBalance)
	}

	sms := api.Group("/sms")
	{
		sms.POST("/", smsHandler.SendSMS)
		sms.POST("/batch", smsHandler.SendSMSBatch)
		sms.GET("/report", smsHandler.GetReport)
	}
}
