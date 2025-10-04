package testutil

import (
	"testing"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates a test database for testing
// Uses Postgres test database from environment or defaults to test instance
func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
	// Hardcode test database URL for now
	// TODO: Fix the environment variable issue that's adding _test suffix
	testDBURL := "postgres://fleettracker:password123@host.docker.internal:5432/fleettracker?sslmode=disable"

	// Create database connection
	db, err := gorm.Open(postgres.Open(testDBURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
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

