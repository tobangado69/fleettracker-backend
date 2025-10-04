package seeds

import (
	"log"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// SeedTrips creates realistic trip records with fuel consumption
func SeedTrips(db *gorm.DB) error {
	log.Println("🗺️  Seeding trips...")

	// Jakarta trips
	jakartaTrips := []models.Trip{
		{
			ID:          "990e8400-e29b-41d4-a716-446655440001",
			CompanyID:   "550e8400-e29b-41d4-a716-446655440001",
			VehicleID:   "770e8400-e29b-41d4-a716-446655440001",
			DriverID:    ptrString("880e8400-e29b-41d4-a716-446655440001"),
			Name:        "JKT-001-" + time.Now().Format("20060102"),
			Purpose:     "Pengiriman barang ke Blok M",
			
			StartLocation:  "Monas, Jakarta Pusat",
			StartLatitude:  -6.1751,
			StartLongitude: 106.8272,
			EndLocation:    "Blok M, Jakarta Selatan",
			EndLatitude:    -6.2350,
			EndLongitude:   106.8000,
			
			StartTime:       ptrTime(time.Now().Add(-20 * time.Hour)),
			EndTime:         ptrTime(time.Now().Add(-19 * time.Hour)),
			
			TotalDistance:   7.2,
			TotalDuration:   3900, // 65 minutes in seconds
			AverageSpeed:    35.5,
			MaxSpeed:        62.0,
			IdleTime:        8,
			
			FuelConsumed:    0.9,
			FuelEfficiency:  8.0,
			StartFuelLevel:  75.0,
			EndFuelLevel:    74.1,
			
			Violations:         0,
			HarshBraking:       1,
			RapidAcceleration:  2,
			SharpCornering:     0,
			SpeedingEvents:     0,
			
			Status:    "completed",
			CreatedAt: time.Now().Add(-20 * time.Hour),
			UpdatedAt: time.Now().Add(-19 * time.Hour),
		},
		{
			ID:          "990e8400-e29b-41d4-a716-446655440002",
			CompanyID:   "550e8400-e29b-41d4-a716-446655440001",
			VehicleID:   "770e8400-e29b-41d4-a716-446655440001",
			DriverID:    ptrString("880e8400-e29b-41d4-a716-446655440001"),
			Name:        "JKT-002-" + time.Now().Format("20060102"),
			Purpose:     "Pengiriman barang ke Sudirman",
			
			StartLocation:  "Blok M, Jakarta Selatan",
			StartLatitude:  -6.2350,
			StartLongitude: 106.8000,
			EndLocation:    "Sudirman, Jakarta Pusat",
			EndLatitude:    -6.2100,
			EndLongitude:   106.8190,
			
			StartTime:       ptrTime(time.Now().Add(-18 * time.Hour)),
			EndTime:         ptrTime(time.Now().Add(-17 * time.Hour)),
			
			TotalDistance:   5.8,
			TotalDuration:   3000, // 50 minutes in seconds
			AverageSpeed:    42.0,
			MaxSpeed:        58.0,
			IdleTime:        5,
			
			FuelConsumed:    0.7,
			FuelEfficiency:  8.3,
			StartFuelLevel:  74.1,
			EndFuelLevel:    73.4,
			
			Violations:         0,
			HarshBraking:       0,
			RapidAcceleration:  1,
			SharpCornering:     1,
			SpeedingEvents:     0,
			
			Status:    "completed",
			CreatedAt: time.Now().Add(-18 * time.Hour),
			UpdatedAt: time.Now().Add(-17 * time.Hour),
		},
		// More Jakarta trips...
		{
			ID:          "990e8400-e29b-41d4-a716-446655440003",
			CompanyID:   "550e8400-e29b-41d4-a716-446655440001",
			VehicleID:   "770e8400-e29b-41d4-a716-446655440002",
			DriverID:    ptrString("880e8400-e29b-41d4-a716-446655440002"),
			Name:        "JKT-003-" + time.Now().Format("20060102"),
			Purpose:     "Pengambilan barang dari Thamrin",
			
			StartLocation:  "Thamrin, Jakarta Pusat",
			StartLatitude:  -6.1900,
			StartLongitude: 106.8230,
			EndLocation:    "Senayan, Jakarta Selatan",
			EndLatitude:    -6.2200,
			EndLongitude:   106.8170,
			
			StartTime:       ptrTime(time.Now().Add(-16 * time.Hour)),
			EndTime:         ptrTime(time.Now().Add(-15 * time.Hour)),
			
			TotalDistance:   4.5,
			TotalDuration:   2520, // 42 minutes in seconds
			AverageSpeed:    38.0,
			MaxSpeed:        55.0,
			IdleTime:        6,
			
			FuelConsumed:    0.5,
			FuelEfficiency:  9.0,
			StartFuelLevel:  80.0,
			EndFuelLevel:    79.5,
			
			Violations:         0,
			HarshBraking:       0,
			RapidAcceleration:  0,
			SharpCornering:     0,
			SpeedingEvents:     0,
			
			Status:    "completed",
			CreatedAt: time.Now().Add(-16 * time.Hour),
			UpdatedAt: time.Now().Add(-15 * time.Hour),
		},
	}

	// Surabaya trips
	surabayaTrips := []models.Trip{
		{
			ID:          "990e8400-e29b-41d4-a716-446655440011",
			CompanyID:   "550e8400-e29b-41d4-a716-446655440002",
			VehicleID:   "770e8400-e29b-41d4-a716-446655440006",
			DriverID:    ptrString("880e8400-e29b-41d4-a716-446655440004"),
			Name:        "SBY-001-" + time.Now().Format("20060102"),
			Purpose:     "Pengiriman ke Delta Plaza",
			
			StartLocation:  "Tugu Pahlawan, Surabaya",
			StartLatitude:  -7.2458,
			StartLongitude: 112.7378,
			EndLocation:    "Delta Plaza, Surabaya",
			EndLatitude:    -7.2700,
			EndLongitude:   112.7600,
			
			StartTime:       ptrTime(time.Now().Add(-22 * time.Hour)),
			EndTime:         ptrTime(time.Now().Add(-21 * time.Hour)),
			
			TotalDistance:   5.2,
			TotalDuration:   3120, // 52 minutes in seconds
			AverageSpeed:    32.0,
			MaxSpeed:        48.0,
			IdleTime:        7,
			
			FuelConsumed:    0.6,
			FuelEfficiency:  8.7,
			StartFuelLevel:  85.0,
			EndFuelLevel:    84.4,
			
			Violations:         0,
			HarshBraking:       1,
			RapidAcceleration:  1,
			SharpCornering:     0,
			SpeedingEvents:     0,
			
			Status:    "completed",
			CreatedAt: time.Now().Add(-22 * time.Hour),
			UpdatedAt: time.Now().Add(-21 * time.Hour),
		},
		{
			ID:          "990e8400-e29b-41d4-a716-446655440012",
			CompanyID:   "550e8400-e29b-41d4-a716-446655440002",
			VehicleID:   "770e8400-e29b-41d4-a716-446655440007",
			DriverID:    ptrString("880e8400-e29b-41d4-a716-446655440005"),
			Name:        "SBY-002-" + time.Now().Format("20060102"),
			Purpose:     "Pengiriman ke Gubeng",
			
			StartLocation:  "Basuki Rahmat, Surabaya",
			StartLatitude:  -7.2550,
			StartLongitude: 112.7450,
			EndLocation:    "Gubeng, Surabaya",
			EndLatitude:    -7.2650,
			EndLongitude:   112.7550,
			
			StartTime:       ptrTime(time.Now().Add(-19 * time.Hour)),
			EndTime:         ptrTime(time.Now().Add(-18 * time.Hour)),
			
			TotalDistance:   3.8,
			TotalDuration:   2280, // 38 minutes in seconds
			AverageSpeed:    40.0,
			MaxSpeed:        52.0,
			IdleTime:        4,
			
			FuelConsumed:    0.4,
			FuelEfficiency:  9.5,
			StartFuelLevel:  70.0,
			EndFuelLevel:    69.6,
			
			Violations:         0,
			HarshBraking:       0,
			RapidAcceleration:  0,
			SharpCornering:     1,
			SpeedingEvents:     0,
			
			Status:    "completed",
			CreatedAt: time.Now().Add(-19 * time.Hour),
			UpdatedAt: time.Now().Add(-18 * time.Hour),
		},
	}

	allTrips := append(jakartaTrips, surabayaTrips...)
	
	for _, trip := range allTrips {
		var existing models.Trip
		result := db.Where("id = ?", trip.ID).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&trip).Error; err != nil {
				return err
			}
			log.Printf("  ✅ Created trip: %s (%s)", trip.Name, trip.Purpose)
		} else {
			log.Printf("  ⏭️  Trip already exists: %s", trip.Name)
		}
	}

	return nil
}

