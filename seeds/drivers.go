package seeds

import (
	"log"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// SeedDrivers creates 5 drivers with valid Indonesian driver's licenses
func SeedDrivers(db *gorm.DB) error {
	log.Println("🚗 Seeding drivers...")

	drivers := []models.Driver{
		// Jakarta company drivers (3)
		{
			ID:        "880e8400-e29b-41d4-a716-446655440001",
			CompanyID: "550e8400-e29b-41d4-a716-446655440001",
			FirstName: "Joko",
			LastName:  "Susanto",
			Email:     "joko.susanto@logistikjkt.co.id",
			Phone:     "+62 856-1234-5678",
			DateOfBirth: ptrDate(time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)),
			NIK:       "3171051590001111",
			Address:   "Jl. Kebon Jeruk Raya No. 123",
			City:      "Jakarta Barat",
			Province:  "DKI Jakarta",
			PostalCode: "11530",
			
			// Driver License
			SIMNumber:         "3171-0515-9000-1111",
			SIMType:           "B2",
			SIMExpiry:         ptrTime(time.Now().AddDate(2, 0, 0)),
			
			// Employment
			HireDate:       ptrDate(time.Now().AddDate(-2, 0, 0)),
			Position:       "permanent",
			
			// Health & Safety
			// BloodType field doesn't exist in model
			// EmergencyBloodType:           "O",
			EmergencyContact1Name:        "Siti Susanto",
			EmergencyContact1Phone:       "+62 812-9876-5432",
			EmergencyContact1Relation:    "Istri",
			// MedicalCheckupDate field doesn't exist
			// MedicalCheckupDate:           ptrDate(time.Now().AddDate(0, -3, 0)),
			// MedicalCheckupExpiry field doesn't exist
			// MedicalCheckupExpiry:         ptrDate(time.Now().AddDate(0, 9, 0)),
			
			// Performance (fields don't exist in model)
			// TotalTrips:     145,
			// TotalDistance:  12450.5,
			// Rating:         4.7,
			// ViolationsCount: 2,
			
			Status:      "active",
			// IsAvailable field doesn't exist
			// IsAvailable: true,
			VehicleID:        ptrString("770e8400-e29b-41d4-a716-446655440001"),
			
			CreatedAt: time.Now().AddDate(0, -6, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "880e8400-e29b-41d4-a716-446655440002",
			CompanyID: "550e8400-e29b-41d4-a716-446655440001",
			FirstName: "Bambang",
			LastName:  "Hartono",
			Email:     "bambang.hartono@logistikjkt.co.id",
			Phone:     "+62 857-2345-6789",
			DateOfBirth: ptrDate(time.Date(1988, 3, 20, 0, 0, 0, 0, time.UTC)),
			NIK:       "3172032088002222",
			Address:   "Jl. Fatmawati No. 45",
			City:      "Jakarta Selatan",
			Province:  "DKI Jakarta",
			PostalCode: "12410",
			
			SIMNumber:         "3172-0320-8800-2222",
			SIMType:           "B2",
			SIMExpiry:         ptrTime(time.Now().AddDate(1, 6, 0)),
			
			HireDate:       ptrDate(time.Now().AddDate(-3, 0, 0)),
			Position:       "permanent",
			
			// BloodType field doesn't exist in model
			// EmergencyBloodType:           "A",
			EmergencyContact1Name:        "Dewi Hartono",
			EmergencyContact1Phone:       "+62 813-8765-4321",
			EmergencyContact1Relation:    "Istri",
			
			// TotalTrips:     198,
			// TotalDistance:  18750.3,
			// Rating:         4.8,
			// ViolationsCount: 1,
			
			Status:      "active",
			// IsAvailable field doesn't exist
			// IsAvailable: true,
			VehicleID:        ptrString("770e8400-e29b-41d4-a716-446655440002"),
			
			CreatedAt: time.Now().AddDate(0, -5, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "880e8400-e29b-41d4-a716-446655440003",
			CompanyID: "550e8400-e29b-41d4-a716-446655440001",
			FirstName: "Andi",
			LastName:  "Firmansyah",
			Email:     "andi.firmansyah@logistikjkt.co.id",
			Phone:     "+62 858-3456-7890",
			DateOfBirth: ptrDate(time.Date(1995, 8, 10, 0, 0, 0, 0, time.UTC)),
			NIK:       "3173081095003333",
			Address:   "Jl. Kuningan Timur No. 67",
			City:      "Jakarta Selatan",
			Province:  "DKI Jakarta",
			PostalCode: "12950",
			
		SIMNumber:     "3173-0810-9500-3333",
		SIMType:           "B1",
		SIMExpiry:         ptrTime(time.Now().AddDate(3, 0, 0)),
		
		HireDate:       ptrDate(time.Now().AddDate(-1, 0, 0)),
		Position:       "permanent",
			
			// BloodType field doesn't exist in model
			// EmergencyBloodType:           "B",
			EmergencyContact1Name:        "Rina Firmansyah",
			EmergencyContact1Phone:       "+62 814-7654-3210",
			EmergencyContact1Relation:    "Ibu",
			
			// TotalTrips:     67,
			// TotalDistance:  5240.8,
			// Rating:         4.5,
			// ViolationsCount: 0,
			
			Status:      "active",
			// IsAvailable field doesn't exist
			// IsAvailable: true,
			VehicleID:        ptrString("770e8400-e29b-41d4-a716-446655440004"),
			
			CreatedAt: time.Now().AddDate(0, -3, 0),
			UpdatedAt: time.Now(),
		},

		// Surabaya company drivers (2)
		{
			ID:        "880e8400-e29b-41d4-a716-446655440004",
			CompanyID: "550e8400-e29b-41d4-a716-446655440002",
			FirstName: "Agus",
			LastName:  "Setiawan",
			Email:     "agus.setiawan@transportsby.co.id",
			Phone:     "+62 859-4567-8901",
			DateOfBirth: ptrDate(time.Date(1992, 6, 25, 0, 0, 0, 0, time.UTC)),
			NIK:       "3578062592004444",
			Address:   "Jl. Darmo Permai No. 89",
			City:      "Surabaya",
			Province:  "Jawa Timur",
			PostalCode: "60189",
			
		SIMNumber:     "3578-0625-9200-4444",
		SIMType:           "B2",
		SIMExpiry:         ptrTime(time.Now().AddDate(2, 3, 0)),
		
		HireDate:       ptrDate(time.Now().AddDate(-1, -6, 0)),
		Position:       "permanent",
			
			// BloodType field doesn't exist in model
			// EmergencyBloodType:           "AB",
			EmergencyContact1Name:        "Yuni Setiawan",
			EmergencyContact1Phone:       "+62 815-6543-2109",
			EmergencyContact1Relation:    "Istri",
			
			// TotalTrips:     112,
			// TotalDistance:  9870.2,
			// Rating:         4.6,
			// ViolationsCount: 3,
			
			Status:      "active",
			// IsAvailable field doesn't exist
			// IsAvailable: true,
			VehicleID:        ptrString("770e8400-e29b-41d4-a716-446655440006"),
			
			CreatedAt: time.Now().AddDate(0, -4, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "880e8400-e29b-41d4-a716-446655440005",
			CompanyID: "550e8400-e29b-41d4-a716-446655440002",
			FirstName: "Rudi",
			LastName:  "Wijayanto",
			Email:     "rudi.wijayanto@transportsby.co.id",
			Phone:     "+62 851-5678-9012",
			DateOfBirth: ptrDate(time.Date(1987, 11, 5, 0, 0, 0, 0, time.UTC)),
			NIK:       "3579110587005555",
			Address:   "Jl. Ahmad Yani No. 123",
			City:      "Surabaya",
			Province:  "Jawa Timur",
			PostalCode: "60234",
			
		SIMNumber:     "3579-1105-8700-5555",
		SIMType:           "B2",
		SIMExpiry:         ptrTime(time.Now().AddDate(1, 0, 0)),
		
		HireDate:       ptrDate(time.Now().AddDate(-2, -3, 0)),
		Position:       "permanent",
			
			// BloodType field doesn't exist in model
			// EmergencyBloodType:           "O",
			EmergencyContact1Name:        "Ani Wijayanto",
			EmergencyContact1Phone:       "+62 816-5432-1098",
			EmergencyContact1Relation:    "Istri",
			
			// TotalTrips:     176,
			// TotalDistance:  15230.7,
			// Rating:         4.9,
			// ViolationsCount: 0,
			
			Status:      "active",
			// IsAvailable field doesn't exist
			// IsAvailable: true,
			VehicleID:        ptrString("770e8400-e29b-41d4-a716-446655440007"),
			
			CreatedAt: time.Now().AddDate(0, -5, 0),
			UpdatedAt: time.Now(),
		},
	}

	for _, driver := range drivers {
		var existing models.Driver
		result := db.Where("id = ?", driver.ID).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&driver).Error; err != nil {
				return err
			}
			log.Printf("  ✅ Created driver: %s %s (SIM: %s)", driver.FirstName, driver.LastName, driver.SIMNumber)
		} else {
			log.Printf("  ⏭️  Driver already exists: %s %s", driver.FirstName, driver.LastName)
		}
	}

	return nil
}
