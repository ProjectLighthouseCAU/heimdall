package setup

import (
	"fmt"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/config"
	"github.com/ProjectLighthouseCAU/heimdall/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// connectPostgres opens a connection to the PostgreSQL database for GORM to use
func connectPostgres() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DatabaseHost,
		config.DatabasePort,
		config.DatabaseUser,
		config.DatabasePassword,
		config.DatabaseName)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{
		TranslateError: true,
		PrepareStmt:    true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, model.InternalServerError{Message: "Could not connect to postgres database", Err: err}
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, model.InternalServerError{Message: "Failed to get underlying sql.DB", Err: err}
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
