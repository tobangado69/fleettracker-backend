package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes connection to PostgreSQL database
func Connect(databaseURL string) (*gorm.DB, error) {
	// Configure GORM logger
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Successfully connected to PostgreSQL database")
	return db, nil
}

// ConnectRedis establishes connection to Redis
func ConnectRedis(redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Println("✅ Successfully connected to Redis")
	return client, nil
}

// ConnectTimescaleDB establishes connection to TimescaleDB for time-series data
func ConnectTimescaleDB(timeseriesURL string) (*gorm.DB, error) {
	// Configure GORM for time-series data
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // Reduce logging for time-series
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Connect to TimescaleDB
	db, err := gorm.Open(postgres.Open(timeseriesURL), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TimescaleDB: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool for time-series data
	sqlDB.SetMaxOpenConns(50)  // Fewer connections for time-series
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(2 * time.Hour) // Longer lifetime for time-series

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping TimescaleDB: %w", err)
	}

	// Enable TimescaleDB extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb").Error; err != nil {
		log.Printf("Warning: Failed to enable TimescaleDB extension: %v", err)
	}

	log.Println("✅ Successfully connected to TimescaleDB")
	return db, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	// Import models here when they are created
	// err := db.AutoMigrate(
	//     &models.User{},
	//     &models.Company{},
	//     &models.Vehicle{},
	//     &models.Driver{},
	//     &models.GPSTrack{},
	//     &models.DriverEvent{},
	//     &models.Trip{},
	//     &models.Payment{},
	//     &models.Subscription{},
	// )

	// For now, just log that migrations would run here
	log.Println("📝 Database migrations would run here")
	log.Println("✅ Database migrations completed")

	return nil
}

// Close closes database connections
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	log.Println("✅ Database connection closed")
	return nil
}

// HealthCheck performs database health check
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetStats returns database connection statistics
func GetStats(db *gorm.DB) (map[string]interface{}, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                   stats.InUse,
		"idle":                     stats.Idle,
		"wait_count":               stats.WaitCount,
		"wait_duration":            stats.WaitDuration.String(),
		"max_idle_closed":          stats.MaxIdleClosed,
		"max_idle_time_closed":     stats.MaxIdleTimeClosed,
		"max_lifetime_closed":      stats.MaxLifetimeClosed,
	}, nil
}
