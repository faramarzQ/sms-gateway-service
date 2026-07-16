package database

import (
	"gorm.io/gorm/logger"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgres() *gorm.DB {
	dsn := "host=localhost user=postgres password=postgres dbname=sms_gateway port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	return db
}
