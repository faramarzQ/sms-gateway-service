package main

import (
	"github.com/faramarzQ/sms-gateway-service/internals/app"
	"github.com/faramarzQ/sms-gateway-service/internals/database"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	db := database.Connect()

	router := gin.Default()

	application := app.NewApp(db, router)
	application.Build()

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
