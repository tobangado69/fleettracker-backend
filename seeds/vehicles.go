package seeds

import (
	"log"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// SeedVehicles creates 10 vehicles with realistic Indonesian data
func SeedVehicles(db *gorm.DB) error {
	log.Println("🚛 Seeding vehicles...")

	vehicles := []models.Vehicle{
		// Jakarta company vehicles (5)
		{
			ID:             "770e8400-e29b-41d4-a716-446655440001",
			CompanyID:      "550e8400-e29b-41d4-a716-446655440001",
			LicensePlate:   "B 1234 ABC",
			Make:           "Toyota",
			Model:          "Dyna",
			Year:           2022,
			Color:          "Putih",
			Type:           "truck",
			FuelType:       "diesel",
			TankCapacity:   80.0,
			Status:         "active",
			PurchasePrice:  450000000, // 450 juta IDR
			STNK:           "B-1234-ABC-2022",
			STNKExpiry:     ptrDate(time.Now().AddDate(1, 0, 0)),
			BPKB:           "BPKB-JKT-001",
			OdometerReading: 45000,
			LastServiceDate: ptrDate(time.Now().AddDate(0, -1, 0)),
			NextServiceDate: ptrDate(time.Now().AddDate(0, 2, 0)),
			IsGPSEnabled:    true,
			IsActive:        true,
			CreatedAt:      time.Now().AddDate(0, -6, 0),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             "770e8400-e29b-41d4-a716-446655440002",
			CompanyID:      "550e8400-e29b-41d4-a716-446655440001",
			LicensePlate:   "B 5678 DEF",
			Make:           "Mitsubishi",
			Model:          "L300",
			Year:           2021,
			Color:          "Silver",
			Type:           "van",
			FuelType:       "diesel",
			TankCapacity:   55.0,
			Status:         "active",
			STNK:           "B-5678-DEF-2021",
			OdometerReading: 67000,
			IsGPSEnabled:    true,
			IsActive:        true,
			CreatedAt:      time.Now().AddDate(0, -5, 0),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             "770e8400-e29b-41d4-a716-446655440003",
			CompanyID:      "550e8400-e29b-41d4-a716-446655440001",
			LicensePlate:   "B 9012 GHI",
			Make:           "Isuzu",
			Model:          "Elf",
			Year:           2023,
			Color:          "Biru",
			Type:           "truck",
			FuelType:       "diesel",
			TankCapacity:   100.0,
			Status:         "active",
			STNK:           "B-9012-GHI-2023",
			OdometerReading: 12000,
			IsGPSEnabled:    true,
			IsActive:        true,
			CreatedAt:      time.Now().AddDate(0, -3, 0),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             "770e8400-e29b-41d4-a716-446655440004",
			CompanyID:      "550e8400-e29b-41d4-a716-446655440001",
			LicensePlate:   "B 3456 JKL",
			Make:           "Toyota",
			Model:          "Avanza",
			Year:           2020,
			Color:          "Hitam",
			Type:           "car",
			FuelType:       "gasoline",
			TankCapacity:   45.0,
			Status:         "active",
			STNK:           "B-3456-JKL-2020",
			OdometerReading: 89000,
			IsGPSEnabled:    true,
			IsActive:        true,
			CreatedAt:      time.Now().AddDate(0, -4, 0),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             "770e8400-e29b-41d4-a716-446655440005",
			CompanyID:      "550e8400-e29b-41d4-a716-446655440001",
			LicensePlate:   "B 7890 MNO",
			Make:           "Honda",
			Model:          "Brio",
			Year:           2021,
			Color:          "Merah",
			Type:           "car",
			FuelType:       "gasoline",
			TankCapacity:   35.0,
			Status:         "active",
			STNK:           "B-7890-MNO-2021",
			OdometerReading: 54000,
			IsGPSEnabled:    true,
			IsActive:        true,
			CreatedAt:      time.Now().AddDate(0, -2, 0),
			UpdatedAt:      time.Now(),
		},

		// Surabaya company vehicles (5)
		{
			ID:             "770e8400-e29b-41d4-a716-446655440006",
			CompanyID:      "550e8400-e29b-41d4-a716-446655440002",
			LicensePlate:   "L 1111 PQR",
			Make:           "Hino",
			Model:          "Dutro",
			Year:           2022,
			Color:          "Putih",
			Type:           "truck",
			FuelType:       "diesel",
			TankCapacity:   90.0,
			Status:         "active",
			STNK:           "L-1111-PQR-2022",
			OdometerReading: 38000,
			IsGPSEnabled:    true,
			IsActive:        true,
			CreatedAt:      time.Now().AddDate(0, -3, 0),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             "770e8400-e29b-41d4-a716-446655440007",
			CompanyID:      "550e8400-e29b-41d4-a716-446655440002",
			LicensePlate:   "L 2222 STU",
			Make:           "Daihatsu",
			Model:          "Gran Max",
			Year:           2021,
			Color:          "Silver",
			Type:           "van",
			FuelType:       "gasoline",
			TankCapacity:   43.0,
			Status:         "active",
			STNK:           "L-2222-STU-2021",
			OdometerReading: 71000,
			IsGPSEnabled:    true,
			IsActive:        true,
			CreatedAt:      time.Now().AddDate(0, -3, 0),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             "770e8400-e29b-41d4-a716-446655440008",
			CompanyID:      "550e8400-e29b-41d4-a716-446655440002",
			LicensePlate:   "L 3333 VWX",
			Make:           "Suzuki",
			Model:          "Carry",
			Year:           2020,
			Color:          "Kuning",
			Type:           "van",
			FuelType:       "gasoline",
			TankCapacity:   40.0,
			Status:         "active",
			STNK:           "L-3333-VWX-2020",
			OdometerReading: 95000,
			IsGPSEnabled:    true,
			IsActive:        true,
			CreatedAt:      time.Now().AddDate(0, -2, 0),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             "770e8400-e29b-41d4-a716-446655440009",
			CompanyID:      "550e8400-e29b-41d4-a716-446655440002",
			LicensePlate:   "L 4444 YZA",
			Make:           "Toyota",
			Model:          "Innova",
			Year:           2023,
			Color:          "Abu-abu",
			Type:           "car",
			FuelType:       "diesel",
			TankCapacity:   50.0,
			Status:         "active",
			STNK:           "L-4444-YZA-2023",
			OdometerReading: 8000,
			IsGPSEnabled:    true,
			IsActive:        true,
			CreatedAt:      time.Now().AddDate(0, -1, 0),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             "770e8400-e29b-41d4-a716-446655440010",
			CompanyID:      "550e8400-e29b-41d4-a716-446655440002",
			LicensePlate:   "L 5555 BCD",
			Make:           "Mitsubishi",
			Model:          "Colt Diesel",
			Year:           2022,
			Color:          "Hijau",
			Type:           "truck",
			FuelType:       "diesel",
			TankCapacity:   70.0,
			Status:         "active",
			STNK:           "L-5555-BCD-2022",
			OdometerReading: 42000,
			IsGPSEnabled:    true,
			IsActive:        true,
			CreatedAt:      time.Now().AddDate(0, -2, 0),
			UpdatedAt:      time.Now(),
		},
	}

	for _, vehicle := range vehicles {
		var existing models.Vehicle
		result := db.Where("id = ?", vehicle.ID).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&vehicle).Error; err != nil {
				return err
			}
			log.Printf("  ✅ Created vehicle: %s (%s %s)", vehicle.LicensePlate, vehicle.Make, vehicle.Model)
		} else {
			log.Printf("  ⏭️  Vehicle already exists: %s", vehicle.LicensePlate)
		}
	}

	return nil
}

// Helper functions for pointers
func ptrDecimal(f float64) *float64 {
	return &f
}

