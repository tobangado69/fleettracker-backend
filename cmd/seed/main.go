package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/tobangado69/fleettracker-pro/backend/internal/common/config"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/database"
	"github.com/tobangado69/fleettracker-pro/backend/seeds"
)

func main() {
	// Command line flags
	companiesOnly := flag.Bool("companies", false, "Seed companies only")
	usersOnly := flag.Bool("users", false, "Seed users only")
	clear := flag.Bool("clear", false, "Clear all seed data before seeding")
	help := flag.Bool("help", false, "Show help message")
	
	flag.Parse()

	// Show help
	if *help {
		showHelp()
		return
	}

	// ASCII art banner
	printBanner()

	// Load configuration
	log.Println("📋 Loading configuration...")
	cfg := config.Load()

	// Connect to database
	log.Println("🔌 Connecting to database...")
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database successfully")
	log.Println("")

	// Clear data if requested
	if *clear {
		if err := seeds.ClearAll(db); err != nil {
			log.Fatalf("❌ Failed to clear data: %v", err)
		}
		log.Println("")
	}

	// Execute seeding based on flags
	if *companiesOnly {
		log.Println("🏢 Seeding companies only...")
		if err := seeds.SeedCompanies(db); err != nil {
			log.Fatalf("❌ Failed to seed companies: %v", err)
		}
	} else if *usersOnly {
		log.Println("👥 Seeding users only...")
		if err := seeds.SeedUsers(db); err != nil {
			log.Fatalf("❌ Failed to seed users: %v", err)
		}
	} else {
		// Run all seeds
		if err := seeds.RunAll(db); err != nil {
			log.Fatalf("❌ Seeding failed: %v", err)
		}
	}

	log.Println("")
	log.Println("🎉 Seeding completed successfully!")
	log.Println("")
	showQuickStart()
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║           🚛  FleetTracker Pro - Database Seeder         ║
║        Indonesian Fleet Management SaaS Platform         ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func showHelp() {
	help := `
FleetTracker Pro Database Seeder

Usage:
  go run cmd/seed/main.go [flags]

Flags:
  --companies      Seed companies only
  --users          Seed users only
  --clear          Clear all existing seed data before seeding
  --help           Show this help message

Examples:
  # Seed all data
  go run cmd/seed/main.go

  # Seed companies only
  go run cmd/seed/main.go --companies

  # Clear and reseed all data
  go run cmd/seed/main.go --clear

Using Make commands:
  make seed              # Seed all data
  make seed-companies    # Seed companies only
  make seed-users        # Seed users only
  make db-reset          # Drop, migrate, and seed (full reset)
  make db-status         # Check database status

Seed Data Includes:
  ✅ 2 Indonesian companies (Jakarta & Surabaya)
  ✅ 5 users with different roles (admin, manager, operator)
  ✅ 10 vehicles with Indonesian license plates
  ✅ 5 drivers with valid SIM (driver's licenses)
  ✅ 100+ GPS tracking points (Jakarta & Surabaya routes)
  ✅ 20 completed trips with fuel consumption data

Test Login Credentials:
  Email:    admin@logistikjkt.co.id
  Password: password123
  Role:     Admin (full access)

More users:
  manager.jakarta@logistikjkt.co.id     / password123 (Manager)
  operator.jakarta@logistikjkt.co.id    / password123 (Operator)
  manager.surabaya@transportsby.co.id   / password123 (Manager)
  operator.surabaya@transportsby.co.id  / password123 (Operator)

Database Requirements:
  - PostgreSQL with UUID and PostGIS extensions
  - Run migrations first: make migrate-up
  - Connection string in .env or DATABASE_URL environment variable

For more information, see:
  - migrations/README.md
  - seeds/README.md
  - DOCKER_SETUP.md
`
	fmt.Println(help)
}

func showQuickStart() {
	quickStart := `
╔═══════════════════════════════════════════════════════════╗
║                    🚀 QUICK START                         ║
╚═══════════════════════════════════════════════════════════╝

1️⃣  Start the backend server:
   make run
   # or
   go run cmd/server/main.go

2️⃣  Open Swagger API documentation:
   http://localhost:8080/swagger/index.html

3️⃣  Test login endpoint:
   POST /api/v1/auth/login
   {
     "email": "admin@logistikjkt.co.id",
     "password": "password123"
   }

4️⃣  Explore the API:
   GET /api/v1/vehicles           # View all 10 vehicles
   GET /api/v1/drivers            # View all 5 drivers
   GET /api/v1/tracking/vehicles  # View GPS tracking data

📊 Database Status:
   make db-status

🔄 Reset Database:
   make db-reset

📚 Documentation:
   - API Docs: http://localhost:8080/swagger/index.html
   - Migrations: migrations/README.md
   - Seeds: seeds/README.md

Happy Testing! 🎉
`
	fmt.Println(quickStart)
}

