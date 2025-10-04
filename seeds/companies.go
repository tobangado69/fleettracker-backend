package seeds

import (
	"log"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// SeedCompanies creates two Indonesian fleet management companies
func SeedCompanies(db *gorm.DB) error {
	log.Println("🏢 Seeding companies...")

	companies := []models.Company{
		{
			ID:          "550e8400-e29b-41d4-a716-446655440001",
			Name:        "PT Logistik Jakarta Raya",
			Email:       "admin@logistikjkt.co.id",
			Phone:       "+62 21-5551-2345",
			Address:     "Gedung Logistik Center Lt. 5, Jl. Jend. Sudirman Kav. 25",
			City:        "Jakarta Pusat",
			Province:    "DKI Jakarta",
			PostalCode:  "10210",
			Country:     "Indonesia",
			
			// Indonesian Compliance
			NPWP:        "01.234.567.8-901.000",
			SIUP:        "SIUP/01234/DKI/2023",
			SKT:         "SKT/01234/DKI/2023",
			PKP:         true,
			CompanyType: "PT",
			
			// Business Information
			Industry:         "Logistics & Transportation",
			FleetSize:        25,
			MaxVehicles:      50,
			SubscriptionTier: "professional",
			
			// Status
			Status:   "active",
			IsActive: true,
			Settings: models.JSON{
				"notification_email": "ops@logistikjkt.co.id",
				"working_hours": "07:00-19:00",
				"timezone": "Asia/Jakarta",
				"currency": "IDR",
				"language": "id",
				"features": map[string]interface{}{
					"gps_tracking": true,
					"fuel_monitoring": true,
					"driver_performance": true,
					"maintenance_alerts": true,
				},
			},
			
			CreatedAt: time.Now().AddDate(0, -6, 0), // 6 months ago
			UpdatedAt: time.Now(),
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440002",
			Name:        "CV Transport Surabaya Jaya",
			Email:       "admin@transportsby.co.id",
			Phone:       "+62 31-5551-2345",
			Address:     "Kompleks Ruko Surya Mas Blok A No. 15",
			City:        "Surabaya",
			Province:    "Jawa Timur",
			PostalCode:  "60271",
			Country:     "Indonesia",
			
			// Indonesian Compliance
			NPWP:        "02.345.678.9-902.000",
			SIUP:        "SIUP/02345/JTM/2023",
			SKT:         "SKT/02345/JTM/2023",
			PKP:         false,
			CompanyType: "CV",
			
			// Business Information
			Industry:         "Transportation Services",
			FleetSize:        15,
			MaxVehicles:      25,
			SubscriptionTier: "basic",
			
			// Status
			Status:   "active",
			IsActive: true,
			Settings: models.JSON{
				"notification_email": "ops@transportsby.co.id",
				"working_hours": "06:00-18:00",
				"timezone": "Asia/Jakarta",
				"currency": "IDR",
				"language": "id",
				"features": map[string]interface{}{
					"gps_tracking": true,
					"fuel_monitoring": false,
					"driver_performance": true,
					"maintenance_alerts": false,
				},
			},
			
			CreatedAt: time.Now().AddDate(0, -3, 0), // 3 months ago
			UpdatedAt: time.Now(),
		},
	}

	for _, company := range companies {
		var existing models.Company
		result := db.Where("id = ?", company.ID).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&company).Error; err != nil {
				return err
			}
			log.Printf("  ✅ Created company: %s", company.Name)
		} else {
			log.Printf("  ⏭️  Company already exists: %s", company.Name)
		}
	}

	return nil
}

