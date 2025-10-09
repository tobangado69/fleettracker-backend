package seeds

import (
	"log"

	"gorm.io/gorm"
)

// RunAll executes all seed functions in the correct order
func RunAll(db *gorm.DB) error {
	log.Println("🌱 Starting database seeding...")

	// Seed super-admin FIRST (entry point for the entire system)
	if err := SeedSuperAdmin(db); err != nil {
		return err
	}

	// Seed in dependency order
	if err := SeedCompanies(db); err != nil {
		return err
	}

	if err := SeedUsers(db); err != nil {
		return err
	}

	if err := SeedVehicles(db); err != nil {
		return err
	}

	if err := SeedDrivers(db); err != nil {
		return err
	}

	if err := SeedGPSTracks(db); err != nil {
		return err
	}

	if err := SeedTrips(db); err != nil {
		return err
	}

	log.Println("✅ Database seeding completed successfully!")
	log.Println("")
	log.Println("📊 Seed Data Summary:")
	log.Println("  - Super-Admin: 1 (platform administrator)")
	log.Println("  - Companies: 2")
	log.Println("  - Users: 5 (1 admin, 2 managers, 2 operators)")
	log.Println("  - Vehicles: 10")
	log.Println("  - Drivers: 5")
	log.Println("  - GPS Tracks: 100+")
	log.Println("  - Trips: 20")
	log.Println("")
	log.Println("🔐 Super-Admin Login (CHANGE PASSWORD IMMEDIATELY):")
	log.Println("  Email: admin@fleettracker.id")
	log.Println("  Password: ChangeMe123!")
	log.Println("")
	log.Println("🔐 Test Company Login:")
	log.Println("  Email: admin@logistikjkt.co.id")
	log.Println("  Password: password123")
	log.Println("")

	return nil
}

// ClearAll deletes all seed data (useful for testing)
func ClearAll(db *gorm.DB) error {
	log.Println("🧹 Clearing all seed data...")

	// Delete in reverse dependency order
	tables := []string{
		"gps_tracks",
		"trips",
		"performance_logs",
		"driver_events",
		"drivers",
		"fuel_logs",
		"maintenance_logs",
		"vehicle_history",
		"vehicles",
		"password_reset_tokens",
		"audit_logs",
		"sessions",
		"users",
		"subscriptions",
		"payments",
		"invoices",
		"geofences",
		"companies",
	}

	for _, table := range tables {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			log.Printf("Warning: Failed to clear %s: %v", table, err)
		}
	}

	log.Println("✅ All seed data cleared")
	return nil
}

