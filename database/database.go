package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"lighthouse.uni-kiel.de/lighthouse-api/config"
)

// Connect opens a connection to the PostgreSQL database for GORM to use
func Connect() *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.GetString("DB_HOST", "localhost"),
		config.GetInt("DB_PORT", 5432),
		config.GetString("DB_USER", "postgres"),
		config.GetString("DB_PASS", "postgres"),
		config.GetString("DB_NAME", "LighthouseAPI"))

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{})

	if err != nil {
		log.Fatalln("Failed to connect to database")
	}
	return db
}
