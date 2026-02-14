package helpers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go-template/internal/model"
	"go-template/internal/seeds"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresContainer struct {
	Container *postgres.PostgresContainer
	DB        *gorm.DB
	URI       string
}

func SetupPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	// 1. Check if external DB DSN is provided
	if dsn := os.Getenv("TEST_DB_DSN"); dsn != "" {
		db, err := gorm.Open(postgresDriver.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to external database: %w", err)
		}

		// Auto-migrate models
		if err := autoMigrate(db); err != nil {
			return nil, fmt.Errorf("failed to auto-migrate: %w", err)
		}

		// Seed data is NOT loaded when using external DB to avoid conflicts
		// Tests using TEST_DB_DSN should manage their own seed data

		return &PostgresContainer{
			Container: nil, // No container to manage
			DB:        db,
			URI:       dsn,
		}, nil
	}

	// 2. Fallback to Testcontainers
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForExposedPort().
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container (ensure Docker is running): %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %w", err)
	}

	// Auto-load test seed data
	if err := seeds.RunForTesting(db); err != nil {
		return nil, fmt.Errorf("failed to seed test data: %w", err)
	}

	return &PostgresContainer{
		Container: pgContainer,
		DB:        db,
		URI:       connStr,
	}, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Client{},
		&model.User{},
		&model.MikrotikRouter{},
		&model.CasbinRule{},
	)
}

func (c *PostgresContainer) Terminate(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	// If using external DB, we don't terminate it
	return nil
}
