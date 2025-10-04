package tracking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/testutil"
)

func TestService_ProcessGPSData(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, nil) // Redis not needed for basic tests

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	tests := []struct {
		name    string
		request GPSDataRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid GPS data - Jakarta",
			request: GPSDataRequest{
				VehicleID: vehicle.ID,
				DriverID:  driver.ID,
				Latitude:  -6.2088,  // Jakarta
				Longitude: 106.8456, // Jakarta
				Speed:     60.0,     // km/h
				Heading:   180.0,
				Accuracy:  5.0,
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid GPS data - Surabaya",
			request: GPSDataRequest{
				VehicleID: vehicle.ID,
				DriverID:  driver.ID,
				Latitude:  -7.2575,  // Surabaya
				Longitude: 112.7521, // Surabaya
				Speed:     80.0,
				Heading:   90.0,
				Accuracy:  10.0,
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid coordinates - out of range latitude",
			request: GPSDataRequest{
				VehicleID: vehicle.ID,
				DriverID:  driver.ID,
				Latitude:  91.0, // Invalid
				Longitude: 106.8456,
				Speed:     60.0,
				Accuracy:  5.0,
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "invalid coordinates",
		},
		{
			name: "invalid coordinates - out of range longitude",
			request: GPSDataRequest{
				VehicleID: vehicle.ID,
				DriverID:  driver.ID,
				Latitude:  -6.2088,
				Longitude: 181.0, // Invalid
				Speed:     60.0,
				Accuracy:  5.0,
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "invalid coordinates",
		},
		{
			name: "poor accuracy",
			request: GPSDataRequest{
				VehicleID: vehicle.ID,
				DriverID:  driver.ID,
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Speed:     60.0,
				Accuracy:  150.0, // Very poor accuracy
				Timestamp: time.Now(),
			},
			wantErr: true,
			errMsg:  "accuracy",
		},
		{
			name: "speed violation",
			request: GPSDataRequest{
				VehicleID: vehicle.ID,
				DriverID:  driver.ID,
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Speed:     150.0, // High speed
				Heading:   180.0,
				Accuracy:  5.0,
				Timestamp: time.Now(),
			},
			wantErr: false, // Should accept but may create event
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gpsTrack, err := service.ProcessGPSData(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, gpsTrack)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gpsTrack)
				testutil.AssertValidUUID(t, gpsTrack.ID)
				assert.Equal(t, tt.request.VehicleID, gpsTrack.VehicleID)
				assert.Equal(t, tt.request.Latitude, gpsTrack.Latitude)
				assert.Equal(t, tt.request.Longitude, gpsTrack.Longitude)
				assert.Equal(t, tt.request.Speed, gpsTrack.Speed)
			}
		})
	}
}

func TestService_ValidateGPSCoordinates(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, nil)

	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		accuracy  float64
		wantErr   bool
	}{
		{
			name:      "valid Jakarta coordinates",
			latitude:  -6.2088,
			longitude: 106.8456,
			accuracy:  5.0,
			wantErr:   false,
		},
		{
			name:      "valid Surabaya coordinates",
			latitude:  -7.2575,
			longitude: 112.7521,
			accuracy:  10.0,
			wantErr:   false,
		},
		{
			name:      "valid Medan coordinates",
			latitude:  3.5952,
			longitude: 98.6722,
			accuracy:  8.0,
			wantErr:   false,
		},
		{
			name:      "latitude too high",
			latitude:  91.0,
			longitude: 106.8456,
			accuracy:  5.0,
			wantErr:   true,
		},
		{
			name:      "latitude too low",
			latitude:  -91.0,
			longitude: 106.8456,
			accuracy:  5.0,
			wantErr:   true,
		},
		{
			name:      "longitude too high",
			longitude: 181.0,
			latitude:  -6.2088,
			accuracy:  5.0,
			wantErr:   true,
		},
		{
			name:      "longitude too low",
			longitude: -181.0,
			latitude:  -6.2088,
			accuracy:  5.0,
			wantErr:   true,
		},
		{
			name:      "poor accuracy",
			latitude:  -6.2088,
			longitude: 106.8456,
			accuracy:  150.0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateGPSCoordinates(tt.latitude, tt.longitude, tt.accuracy)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ProcessDriverEvent(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, nil)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	tests := []struct {
		name    string
		request DriverEventRequest
		wantErr bool
	}{
		{
			name: "speed violation event",
			request: DriverEventRequest{
				VehicleID: vehicle.ID,
				DriverID:  driver.ID,
				EventType: "speed_violation",
				Severity:  "high",
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Timestamp: time.Now(),
				Speed:     120.0,
				Details:   "Exceeded speed limit by 40 km/h",
				Value:     120.0,
			},
			wantErr: false,
		},
		{
			name: "harsh braking event",
			request: DriverEventRequest{
				VehicleID: vehicle.ID,
				DriverID:  driver.ID,
				EventType: "harsh_braking",
				Severity:  "medium",
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Timestamp: time.Now(),
				Speed:     60.0,
				Details:   "Sudden braking detected",
				Value:     -8.5, // Deceleration in m/s²
			},
			wantErr: false,
		},
		{
			name: "rapid acceleration event",
			request: DriverEventRequest{
				VehicleID: vehicle.ID,
				DriverID:  driver.ID,
				EventType: "rapid_acceleration",
				Severity:  "low",
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Timestamp: time.Now(),
				Speed:     80.0,
				Details:   "Rapid acceleration detected",
				Value:     6.0, // Acceleration in m/s²
			},
			wantErr: false,
		},
		{
			name: "invalid event type",
			request: DriverEventRequest{
				VehicleID: vehicle.ID,
				DriverID:  driver.ID,
				EventType: "invalid_type",
				Severity:  "high",
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := service.ProcessDriverEvent(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, event)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, event)
				testutil.AssertValidUUID(t, event.ID)
				assert.Equal(t, tt.request.VehicleID, event.VehicleID)
				assert.Equal(t, tt.request.DriverID, event.DriverID)
				assert.Equal(t, tt.request.EventType, event.EventType)
				assert.Equal(t, tt.request.Severity, event.Severity)
			}
		})
	}
}

func TestService_StartTrip(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, nil)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	odometerStart := 50000.0

	t.Run("start trip successfully", func(t *testing.T) {
		trip, err := service.StartTrip(TripRequest{
			VehicleID: vehicle.ID,
			DriverID:  driver.ID,
			Action:    "start",
			StartLocation: &Location{
				Latitude:  -6.2088,
				Longitude: 106.8456,
				Address:   "Jakarta, Indonesia",
			},
			Timestamp:     time.Now(),
			OdometerStart: &odometerStart,
		})

		assert.NoError(t, err)
		assert.NotNil(t, trip)
		testutil.AssertValidUUID(t, trip.ID)
		assert.Equal(t, vehicle.ID, trip.VehicleID)
		assert.Equal(t, driver.ID, *trip.DriverID)
		assert.Equal(t, "active", trip.Status)
		assert.NotNil(t, trip.StartTime)
		assert.Equal(t, -6.2088, trip.StartLatitude)
		assert.Equal(t, 106.8456, trip.StartLongitude)
	})
}

func TestService_EndTrip(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, nil)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	// Start a trip first
	odometerStart := 50000.0
	trip, err := service.StartTrip(TripRequest{
		VehicleID: vehicle.ID,
		DriverID:  driver.ID,
		Action:    "start",
		StartLocation: &Location{
			Latitude:  -6.2088,
			Longitude: 106.8456,
			Address:   "Jakarta, Indonesia",
		},
		Timestamp:     time.Now().Add(-2 * time.Hour),
		OdometerStart: &odometerStart,
	})
	require.NoError(t, err)

	t.Run("end trip successfully", func(t *testing.T) {
		odometerEnd := 50150.0
		endedTrip, err := service.EndTrip(TripRequest{
			VehicleID: vehicle.ID,
			DriverID:  driver.ID,
			Action:    "end",
			EndLocation: &Location{
				Latitude:  -7.2575,  // Surabaya
				Longitude: 112.7521, // Surabaya
				Address:   "Surabaya, Indonesia",
			},
			Timestamp:   time.Now(),
			OdometerEnd: &odometerEnd,
		})

		assert.NoError(t, err)
		assert.NotNil(t, endedTrip)
		assert.Equal(t, trip.ID, endedTrip.ID)
		assert.Equal(t, "completed", endedTrip.Status)
		assert.NotNil(t, endedTrip.EndTime)
		assert.Equal(t, -7.2575, endedTrip.EndLatitude)
		assert.Equal(t, 112.7521, endedTrip.EndLongitude)
	})
}

func TestService_CalculateDistance(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, nil)

	tests := []struct {
		name      string
		lat1      float64
		lng1      float64
		lat2      float64
		lng2      float64
		minDist   float64 // Expected minimum distance
		maxDist   float64 // Expected maximum distance
	}{
		{
			name:    "Jakarta to Surabaya",
			lat1:    -6.2088,
			lng1:    106.8456,
			lat2:    -7.2575,
			lng2:    112.7521,
			minDist: 650.0, // Approximately 660-680 km
			maxDist: 700.0,
		},
		{
			name:    "same location",
			lat1:    -6.2088,
			lng1:    106.8456,
			lat2:    -6.2088,
			lng2:    106.8456,
			minDist: 0.0,
			maxDist: 0.1,
		},
		{
			name:    "short distance - within city",
			lat1:    -6.2088,
			lng1:    106.8456,
			lat2:    -6.2188,
			lng2:    106.8556,
			minDist: 1.0,
			maxDist: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := service.calculateDistance(tt.lat1, tt.lng1, tt.lat2, tt.lng2)

			assert.GreaterOrEqual(t, distance, tt.minDist, "Distance should be >= minimum")
			assert.LessOrEqual(t, distance, tt.maxDist, "Distance should be <= maximum")
		})
	}
}

func TestService_GetSpeedViolationSeverity(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, nil)

	tests := []struct {
		name     string
		speed    float64
		expected string
	}{
		{
			name:     "normal speed",
			speed:    60.0,
			expected: "low",
		},
		{
			name:     "moderate speeding",
			speed:    95.0,
			expected: "medium",
		},
		{
			name:     "high speeding",
			speed:    125.0,
			expected: "high",
		},
		{
			name:     "critical speeding",
			speed:    155.0,
			expected: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := service.getSpeedViolationSeverity(tt.speed)
			assert.Equal(t, tt.expected, severity)
		})
	}
}

func TestService_CreateGeofence(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, nil)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	tests := []struct {
		name    string
		request GeofenceRequest
		wantErr bool
	}{
		{
			name: "create zone geofence",
			request: GeofenceRequest{
				CompanyID:   company.ID,
				Name:        "Office Area",
				Description: "Main office geofence",
				Type:        "zone",
				CenterLat:   -6.2088,
				CenterLng:   106.8456,
				Radius:      500.0, // 500 meters
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "create large geofence",
			request: GeofenceRequest{
				CompanyID:   company.ID,
				Name:        "City Coverage",
				Description: "Jakarta coverage area",
				Type:        "zone",
				CenterLat:   -6.2088,
				CenterLng:   106.8456,
				Radius:      5000.0, // 5 km
				IsActive:    true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geofence, err := service.CreateGeofence(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, geofence)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, geofence)
				testutil.AssertValidUUID(t, geofence.ID)
				assert.Equal(t, tt.request.CompanyID, geofence.CompanyID)
				assert.Equal(t, tt.request.Name, geofence.Name)
				assert.Equal(t, tt.request.Type, geofence.Type)
			}
		})
	}
}

func TestService_GetLocationHistory(t *testing.T) {
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	service := NewService(db, nil)

	// Create test data
	company := testutil.NewTestCompany()
	require.NoError(t, db.Create(company).Error)

	vehicle := testutil.NewTestVehicle(company.ID)
	require.NoError(t, db.Create(vehicle).Error)

	driver := testutil.NewTestDriver(company.ID)
	require.NoError(t, db.Create(driver).Error)

	// Create multiple GPS tracks
	now := time.Now()
	for i := 0; i < 5; i++ {
		gpsTrack := testutil.NewTestGPSTrack(vehicle.ID)
		gpsTrack.Timestamp = now.Add(time.Duration(-i) * time.Minute)
		require.NoError(t, db.Create(gpsTrack).Error)
	}

	t.Run("get location history with filters", func(t *testing.T) {
		startTime := now.Add(-10 * time.Minute)
		endTime := now
		
		filters := GPSFilters{
			VehicleID: &vehicle.ID,
			StartTime: &startTime,
			EndTime:   &endTime,
			Page:      1,
			Limit:     10,
			SortBy:    "timestamp",
			SortOrder: "desc",
		}

		tracks, total, err := service.GetLocationHistory(vehicle.ID, filters)

		assert.NoError(t, err)
		assert.NotEmpty(t, tracks)
		assert.Greater(t, total, int64(0))
		assert.LessOrEqual(t, len(tracks), 10)
	})
}

