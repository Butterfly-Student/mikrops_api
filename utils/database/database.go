package database

import (
	"context"
	"os"
	"time"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mikrops/utils"
	"mikrops/utils/log"
)

func InitGormDatabase(ctx context.Context, outboundDatabaseDriver string) *gorm.DB {
	dsn := utils.GetDatabaseString()

	logLevel := logger.Warn
	if os.Getenv("APP_MODE") != "release" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.WithContext(ctx).Fatalf("failed to open gorm database: %+v", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.WithContext(ctx).Fatalf("failed to get sql.DB from gorm: %+v", err)
		os.Exit(1)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	if err := sqlDB.Ping(); err != nil {
		log.WithContext(ctx).Fatalf("failed to connect database: %+v", err)
		os.Exit(1)
	}

	if err := goose.Up(sqlDB, utils.GetMigrationDir()); err != nil {
		log.WithContext(ctx).Fatalf("failed to running migration: %+v", err)
		os.Exit(1)
	}

	return db
}
