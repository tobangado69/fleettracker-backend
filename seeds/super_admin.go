package seeds

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// SeedSuperAdmin creates the initial super-admin user if it doesn't exist
// This is the entry point for the entire system - the first user who can create companies and owners
func SeedSuperAdmin(db *gorm.DB) error {
	log.Println("🔐 Checking for super-admin...")

	// Check if super-admin already exists
	var count int64
	if err := db.Model(&models.User{}).Where("role = ?", "super-admin").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Println("✅ Super-admin already exists, skipping creation")
		return nil
	}

	// Generate secure temporary password
	tempPassword := "ChangeMe123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create super-admin user
	now := time.Now()
	superAdmin := models.User{
		Email:              "admin@fleettracker.id",
		Username:           "superadmin",
		FirstName:          "Super",
		LastName:           "Administrator",
		Role:               "super-admin",
		Password:           string(hashedPassword),
		IsActive:           true,
		MustChangePassword: true, // Force password change on first login
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := db.Create(&superAdmin).Error; err != nil {
		return err
	}

	log.Println("")
	log.Println("✅ ═══════════════════════════════════════════════════════════")
	log.Println("✅ Super-admin created successfully!")
	log.Println("✅ ═══════════════════════════════════════════════════════════")
	log.Println("📧 Email: admin@fleettracker.id")
	log.Println("🔑 Temporary Password: ChangeMe123!")
	log.Println("⚠️  IMPORTANT: Change this password immediately after first login!")
	log.Println("✅ ═══════════════════════════════════════════════════════════")
	log.Println("")

	return nil
}

