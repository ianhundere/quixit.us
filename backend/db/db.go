package db

import (
	"log"

	"sample-exchange/backend/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func SetupDB() *gorm.DB {
	var err error
	db, err = gorm.Open(sqlite.Open("sample_exchange.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	return db
}

func GetDB() *gorm.DB {
	return db
}
