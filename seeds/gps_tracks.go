package seeds

import (
	"log"
	"math"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// Jakarta route: Monas to Blok M (approximately 7km)
var jakartaRoute = []struct {
	lat, lon float64
	name     string
}{
	{-6.1751, 106.8272, "Monas"},
	{-6.1800, 106.8250, "Jl. Medan Merdeka Selatan"},
	{-6.1900, 106.8230, "Jl. M.H. Thamrin"},
	{-6.2000, 106.8210, "Bundaran HI"},
	{-6.2100, 106.8190, "Jl. Jend. Sudirman"},
	{-6.2200, 106.8170, "Plaza Senayan"},
	{-6.2350, 106.8000, "Blok M"},
}

// Surabaya route: Tugu Pahlawan to Delta Plaza (approximately 5km)
var surabayaRoute = []struct {
	lat, lon float64
	name     string
}{
	{-7.2458, 112.7378, "Tugu Pahlawan"},
	{-7.2500, 112.7400, "Jl. Pahlawan"},
	{-7.2550, 112.7450, "Jl. Basuki Rahmat"},
	{-7.2600, 112.7500, "Jl. Pemuda"},
	{-7.2650, 112.7550, "Jl. Ahmad Yani"},
	{-7.2700, 112.7600, "Delta Plaza"},
}

// SeedGPSTracks generates realistic GPS tracking data for vehicles
func SeedGPSTracks(db *gorm.DB) error {
	log.Println("📍 Seeding GPS tracks...")

	// Jakarta vehicles
	jakartaVehicles := []string{
		"770e8400-e29b-41d4-a716-446655440001",
		"770e8400-e29b-41d4-a716-446655440002",
		"770e8400-e29b-41d4-a716-446655440003",
	}

	// Surabaya vehicles
	surabayaVehicles := []string{
		"770e8400-e29b-41d4-a716-446655440006",
		"770e8400-e29b-41d4-a716-446655440007",
	}

	totalTracks := 0

	// Generate tracks for Jakarta vehicles (last 24 hours)
	for _, vehicleID := range jakartaVehicles {
		tracks := generateRouteGPSTracks(vehicleID, jakartaRoute, 10) // 10 trips
		for _, track := range tracks {
			var existing models.GPSTrack
			result := db.Where("vehicle_id = ? AND timestamp = ?", track.VehicleID, track.Timestamp).First(&existing)
			
			if result.Error == gorm.ErrRecordNotFound {
				if err := db.Create(&track).Error; err != nil {
					log.Printf("  ⚠️  Failed to create GPS track: %v", err)
					continue
				}
				totalTracks++
			}
		}
	}

	// Generate tracks for Surabaya vehicles
	for _, vehicleID := range surabayaVehicles {
		tracks := generateRouteGPSTracks(vehicleID, surabayaRoute, 10) // 10 trips
		for _, track := range tracks {
			var existing models.GPSTrack
			result := db.Where("vehicle_id = ? AND timestamp = ?", track.VehicleID, track.Timestamp).First(&existing)
			
			if result.Error == gorm.ErrRecordNotFound {
				if err := db.Create(&track).Error; err != nil {
					log.Printf("  ⚠️  Failed to create GPS track: %v", err)
					continue
				}
				totalTracks++
			}
		}
	}

	log.Printf("  ✅ Created %d GPS tracks", totalTracks)
	return nil
}

// generateRouteGPSTracks creates GPS points along a route
func generateRouteGPSTracks(vehicleID string, route []struct{ lat, lon float64; name string }, tripCount int) []models.GPSTrack {
	tracks := []models.GPSTrack{}
	now := time.Now()

	// Generate tracks for the last 24 hours
	for trip := 0; trip < tripCount; trip++ {
		// Space trips throughout the day
		tripStartTime := now.Add(time.Duration(-24+trip*2) * time.Hour)
		
		// Generate points along the route
		for i := 0; i < len(route)-1; i++ {
			start := route[i]
			end := route[i+1]
			
			// Generate 10 points between each waypoint
			for j := 0; j <= 10; j++ {
				progress := float64(j) / 10.0
				
				// Interpolate position
				lat := interpolate(start.lat, end.lat, progress)
				lon := interpolate(start.lon, end.lon, progress)
				
				// Calculate realistic speed (20-60 km/h in city traffic)
				baseSpeed := 35.0
				speedVariation := RandomFloat(-15.0, 15.0)
				speed := math.Max(5.0, baseSpeed+speedVariation)
				
				// Time offset (approximately 1 minute between points)
				timeOffset := time.Duration(i*10+j) * time.Minute
				timestamp := tripStartTime.Add(timeOffset)
				
				// Calculate heading (bearing)
				heading := calculateBearing(start.lat, start.lon, end.lat, end.lon)
				
				track := models.GPSTrack{
					VehicleID: vehicleID,
					Latitude:  lat,
					Longitude: lon,
					Altitude:  RandomFloat(10.0, 50.0), // Sea level + building height
					Heading:   heading,
					Speed:     speed,
					Location:  getLocationName(i, start.name, end.name),
					Accuracy:  RandomFloat(3.0, 15.0),
					Satellites: 8 + int(RandomFloat(0, 5)),
					IgnitionOn: true,
					EngineOn:   true,
					Moving:     speed > 5.0,
					FuelLevel:  RandomFloat(30.0, 90.0),
					Timestamp:  timestamp,
					CreatedAt:  timestamp,
				}
				
				tracks = append(tracks, track)
			}
		}
	}

	return tracks
}

// interpolate linearly between two values
func interpolate(start, end, progress float64) float64 {
	return start + (end-start)*progress
}

// calculateBearing calculates the bearing between two GPS coordinates
func calculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lonDiff := (lon2 - lon1) * math.Pi / 180

	y := math.Sin(lonDiff) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) - math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(lonDiff)
	
	bearing := math.Atan2(y, x) * 180 / math.Pi
	
	// Normalize to 0-360
	if bearing < 0 {
		bearing += 360
	}
	
	return bearing
}

// getLocationName generates a location description
func getLocationName(segmentIndex int, startName, endName string) string {
	locations := []string{
		startName,
		"En route",
		"Approaching " + endName,
		endName,
	}
	
	if segmentIndex < len(locations) {
		return locations[segmentIndex]
	}
	return "En route"
}

