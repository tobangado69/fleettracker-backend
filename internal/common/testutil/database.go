package testutil

import (
	"os"
	"testing"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates a test database for testing
// Uses Postgres test database from environment or defaults to test instance
func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
	// Get database URL from environment
	var testDBURL string
	
	// Priority 1: TEST_DATABASE_URL (specifically for tests)
	if os.Getenv("TEST_DATABASE_URL") != "" {
		testDBURL = os.Getenv("TEST_DATABASE_URL")
		t.Logf("Using TEST_DATABASE_URL from environment")
	} else if os.Getenv("DATABASE_URL") != "" {
		// Priority 2: DATABASE_URL
		testDBURL = os.Getenv("DATABASE_URL")
		t.Logf("Using DATABASE_URL from environment")
	} else {
		// Priority 3: Default to Docker Compose setup
		testDBURL = "postgres://fleettracker:password123@localhost:5432/fleettracker?sslmode=disable"
		t.Logf("Using default Docker Compose configuration")
	}

	// Create database connection with fallback configurations
	var db *gorm.DB
	var err error
	
	// Try different database configurations
	configs := []string{
		testDBURL, // First try the configured URL
		"postgres://fleettracker@localhost:5432/fleettracker?sslmode=disable", // Docker Compose (trust auth)
		"postgres://fleettracker:password123@localhost:5432/fleettracker?sslmode=disable", // Docker Compose (with password)
		"postgres://postgres@localhost:5432/postgres?sslmode=disable", // Default (trust auth)
		"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
	}
	
	for i, config := range configs {
		if config == "" {
			continue
		}
		
		db, err = gorm.Open(postgres.Open(config), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
		})
		if err == nil {
			t.Logf("Connected to database using config %d", i+1)
			break
		}
		t.Logf("Failed to connect with config %d: %v", i+1, err)
	}
	
	if err != nil {
		t.Fatalf("Failed to create test database with any configuration. Please ensure PostgreSQL is running locally. Last error: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Session{},
		&models.AuditLog{},
		&models.PasswordResetToken{},
		&models.Vehicle{},
		&models.MaintenanceLog{},
		&models.FuelLog{},
		&models.Driver{},
		&models.DriverEvent{},
		&models.PerformanceLog{},
		&models.GPSTrack{},
		&models.Trip{},
		&models.Geofence{},
		&models.VehicleHistory{},
		&models.Subscription{},
		&models.Payment{},
		&models.Invoice{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Cleanup function - clear all tables
	cleanup := func() {
		// Clear all data after test
		if err := ClearDatabase(db); err != nil {
			t.Logf("Warning: Failed to clear database: %v", err)
		}
		
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	// Clear any existing data before test
	if err := ClearDatabase(db); err != nil {
		t.Fatalf("Failed to clear database before test: %v", err)
	}

	return db, cleanup
}

// ClearDatabase removes all data from test database
func ClearDatabase(db *gorm.DB) error {
	// Delete in reverse order of dependencies
	tables := []interface{}{
		&models.Invoice{},
		&models.Payment{},
		&models.Subscription{},
		&models.VehicleHistory{},
		&models.Geofence{},
		&models.Trip{},
		&models.GPSTrack{},
		&models.PerformanceLog{},
		&models.DriverEvent{},
		&models.Driver{},
		&models.FuelLog{},
		&models.MaintenanceLog{},
		&models.Vehicle{},
		&models.PasswordResetToken{},
		&models.AuditLog{},
		&models.Session{},
		&models.User{},
		&models.Company{},
	}

	for _, table := range tables {
		if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
			return err
		}
	}

	return nil
}

