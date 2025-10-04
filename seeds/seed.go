package seeds

import (
	"log"

	"gorm.io/gorm"
)

// RunAll executes all seed functions in the correct order
func RunAll(db *gorm.DB) error {
	log.Println("🌱 Starting database seeding...")

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
	log.Println("  - Companies: 2")
	log.Println("  - Users: 5 (1 admin, 2 managers, 2 operators)")
	log.Println("  - Vehicles: 10")
	log.Println("  - Drivers: 5")
	log.Println("  - GPS Tracks: 100+")
	log.Println("  - Trips: 20")
	log.Println("")
	log.Println("🔐 Test Login Credentials:")
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

