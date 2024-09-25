package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/config"
	"github.com/ProjectLighthouseCAU/heimdall/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// connectPostgres opens a connection to the PostgreSQL database for GORM to use
func connectPostgres() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.GetString("DB_HOST", "localhost"),
		config.GetInt("DB_PORT", 5432),
		config.GetString("DB_USER", "postgres"),
		config.GetString("DB_PASS", "postgres"),
		config.GetString("DB_NAME", "heimdall"))

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

func connectRedis(dbNumber int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.GetString("REDIS_HOST", "127.0.0.1") + ":" + config.GetString("REDIS_PORT", "6379"),
		Username: config.GetString("REDIS_USER", ""),
		Password: config.GetString("REDIS_PASSWORD", ""),
		DB:       dbNumber,
	})
	if err := rdb.Ping(context.TODO()).Err(); err != nil {
		return nil, model.InternalServerError{Message: "Failed to connect to redis", Err: err}
	}
	return rdb, nil
}
