package seeds

import (
	"log"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// SeedUsers creates 5 users with different roles
func SeedUsers(db *gorm.DB) error {
	log.Println("👥 Seeding users...")

	// Note: Password will be auto-hashed by BeforeCreate hook in User model
	users := []models.User{
		{
			ID:        "660e8400-e29b-41d4-a716-446655440001",
			CompanyID: "550e8400-e29b-41d4-a716-446655440001", // Jakarta company
			Email:     "admin@logistikjkt.co.id",
			Username:  "admin.jakarta",
			Password:  "password123", // Will be hashed by model hook
			
			FirstName: "Ahmad",
			LastName:  "Santoso",
			Phone:     "+62 812-3456-7890",
			
			// Indonesian Fields
			NIK:        "3171012585001234",
			Address:    "Jl. Sudirman No. 45, RT 005/RW 002",
			City:       "Jakarta Pusat",
			Province:   "DKI Jakarta",
			PostalCode: "10210",
			
			// Role
			Role: "admin",
			Permissions: models.JSON{
				"vehicles": []string{"create", "read", "update", "delete"},
				"drivers": []string{"create", "read", "update", "delete"},
				"users": []string{"create", "read", "update", "delete"},
				"reports": []string{"read", "export"},
				"settings": []string{"read", "update"},
			},
			
			// Status
			Status:     "active",
			IsActive:   true,
			IsVerified: true,
			
		// Preferences
		Language:              "id",
		Timezone:              "Asia/Jakarta",
		
		CreatedAt: time.Now().AddDate(0, -6, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "660e8400-e29b-41d4-a716-446655440002",
			CompanyID: "550e8400-e29b-41d4-a716-446655440001", // Jakarta company
			Email:     "manager.jakarta@logistikjkt.co.id",
			Username:  "manager.jakarta",
			Password:  "password123",
			
			FirstName: "Budi",
			LastName:  "Wijaya",
			Phone:     "+62 813-4567-8901",
			
			NIK:        "3172022090002345",
			Address:    "Jl. Gatot Subroto No. 67",
			City:       "Jakarta Selatan",
			Province:   "DKI Jakarta",
			PostalCode: "12190",
			
			Role: "manager",
			Permissions: models.JSON{
				"vehicles": []string{"read", "update"},
				"drivers": []string{"read", "update"},
				"reports": []string{"read", "export"},
			},
			
		Status:     "active",
		IsActive:   true,
		IsVerified: true,
		Language:   "id",
		Timezone:   "Asia/Jakarta",
		
		CreatedAt: time.Now().AddDate(0, -5, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "660e8400-e29b-41d4-a716-446655440003",
			CompanyID: "550e8400-e29b-41d4-a716-446655440001", // Jakarta company
			Email:     "operator.jakarta@logistikjkt.co.id",
			Username:  "operator.jakarta",
			Password:  "password123",
			
			FirstName: "Dewi",
			LastName:  "Kusuma",
			Phone:     "+62 821-5678-9012",
			
			NIK:        "3173032595003456",
			Address:    "Jl. Rasuna Said No. 12",
			City:       "Jakarta Selatan",
			Province:   "DKI Jakarta",
			PostalCode: "12920",
			
			Role: "operator",
			Permissions: models.JSON{
				"vehicles": []string{"read"},
				"drivers": []string{"read"},
				"tracking": []string{"read", "monitor"},
			},
			
			Status:     "active",
		IsActive:   true,
		IsVerified: true,
		Language:   "id",
		Timezone:   "Asia/Jakarta",
		
		CreatedAt: time.Now().AddDate(0, -4, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "660e8400-e29b-41d4-a716-446655440004",
			CompanyID: "550e8400-e29b-41d4-a716-446655440002", // Surabaya company
			Email:     "manager.surabaya@transportsby.co.id",
			Username:  "manager.surabaya",
			Password:  "password123",
			
			FirstName: "Eko",
			LastName:  "Pratama",
			Phone:     "+62 822-6789-0123",
			
			NIK:        "3578012088004567",
			Address:    "Jl. Basuki Rahmat No. 45",
			City:       "Surabaya",
			Province:   "Jawa Timur",
			PostalCode: "60271",
			
			Role: "manager",
			Permissions: models.JSON{
				"vehicles": []string{"read", "update"},
				"drivers": []string{"read", "update"},
				"reports": []string{"read"},
			},
			
			Status:     "active",
		IsActive:   true,
		IsVerified: true,
		Language:   "id",
		Timezone:   "Asia/Jakarta",
		
		CreatedAt: time.Now().AddDate(0, -3, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "660e8400-e29b-41d4-a716-446655440005",
			CompanyID: "550e8400-e29b-41d4-a716-446655440002", // Surabaya company
			Email:     "operator.surabaya@transportsby.co.id",
			Username:  "operator.surabaya",
			Password:  "password123",
			
			FirstName: "Fitri",
			LastName:  "Saputra",
			Phone:     "+62 823-7890-1234",
			
			NIK:        "3579022592005678",
			Address:    "Jl. Diponegoro No. 34",
			City:       "Surabaya",
			Province:   "Jawa Timur",
			PostalCode: "60285",
			
			Role: "operator",
			Permissions: models.JSON{
				"vehicles": []string{"read"},
				"drivers": []string{"read"},
				"tracking": []string{"read"},
			},
			
		Status:     "active",
		IsActive:   true,
		IsVerified: true,
		Language:   "id",
		Timezone:   "Asia/Jakarta",
		
		CreatedAt: time.Now().AddDate(0, -2, 0),
			UpdatedAt: time.Now(),
		},
	}

	for _, user := range users {
		var existing models.User
		result := db.Where("id = ?", user.ID).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&user).Error; err != nil {
				return err
			}
			log.Printf("  ✅ Created user: %s (%s) - Role: %s", user.Email, user.FirstName+" "+user.LastName, user.Role)
		} else {
			log.Printf("  ⏭️  User already exists: %s", user.Email)
		}
	}

	log.Println("  💡 Login credentials: admin@logistikjkt.co.id / password123")

	return nil
}

