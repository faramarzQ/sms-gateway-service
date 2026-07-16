package http

import (
	"github.com/faramarzQ/sms-gateway-service/internals/http/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(ginRouter *gin.Engine, userHandler *handlers.UserHandler) {

	users := ginRouter.Group("/user")
	{
		users.GET("/:id", userHandler.GetUser)
		// update balance
	}

}
