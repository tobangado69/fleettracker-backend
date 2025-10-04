package models

import (
	"math"
	"time"

	"gorm.io/gorm"
)

// GPSTrack represents GPS tracking data with PostGIS geometry
type GPSTrack struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VehicleID string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	DriverID  *string   `json:"driver_id" gorm:"type:uuid;index"`
	TripID    *string   `json:"trip_id" gorm:"type:uuid;index"`
	
	// GPS Coordinates with PostGIS
	Latitude    float64   `json:"latitude" gorm:"type:decimal(10,8);not null"`
	Longitude   float64   `json:"longitude" gorm:"type:decimal(11,8);not null"`
	Altitude    float64   `json:"altitude" gorm:"type:decimal(8,2)"`
	Heading     float64   `json:"heading" gorm:"type:decimal(5,2)"` // degrees 0-360
	Speed       float64   `json:"speed" gorm:"type:decimal(5,2)"` // km/h
	
	// Location Information
	Location    string    `json:"location" gorm:"type:varchar(255)"`
	Address     string    `json:"address" gorm:"type:text"`
	City        string    `json:"city" gorm:"type:varchar(100)"`
	Province    string    `json:"province" gorm:"type:varchar(100)"`
	Country     string    `json:"country" gorm:"type:varchar(100);default:'Indonesia'"`
	
	// Tracking Data Quality
	Accuracy    float64   `json:"accuracy" gorm:"type:decimal(5,2)"` // meters
	Satellites  int       `json:"satellites" gorm:"default:0"`
	HDOP        float64   `json:"hdop" gorm:"type:decimal(3,1)"` // Horizontal Dilution of Precision
	
	// Vehicle Status
	IgnitionOn  bool      `json:"ignition_on" gorm:"default:false"`
	EngineOn    bool      `json:"engine_on" gorm:"default:false"`
	Moving      bool      `json:"moving" gorm:"default:false"`
	IdleTime    int       `json:"idle_time" gorm:"default:0"` // seconds
	
	// Fuel Information
	FuelLevel   float64   `json:"fuel_level" gorm:"type:decimal(5,2)"` // liters
	FuelConsumption float64 `json:"fuel_consumption" gorm:"type:decimal(8,2)"` // liters
	
	// Distance and Odometer
	Distance    float64   `json:"distance" gorm:"type:decimal(10,2)"` // km from previous point
	Odometer    float64   `json:"odometer" gorm:"type:decimal(10,2)"` // total odometer reading
	
	// Behavior Detection
	HarshBraking       bool    `json:"harsh_braking" gorm:"default:false"`
	RapidAcceleration  bool    `json:"rapid_acceleration" gorm:"default:false"`
	SharpCornering     bool    `json:"sharp_cornering" gorm:"default:false"`
	Speeding           bool    `json:"speeding" gorm:"default:false"`
	Idling             bool    `json:"idling" gorm:"default:false"`
	
	// Speed Limits and Violations
	SpeedLimit         float64 `json:"speed_limit" gorm:"type:decimal(5,2)"` // km/h
	SpeedViolation     bool    `json:"speed_violation" gorm:"default:false"`
	ViolationSeverity  string  `json:"violation_severity" gorm:"type:varchar(20)"` // low, medium, high, critical
	
	// Geofencing
	InGeofence         bool    `json:"in_geofence" gorm:"default:false"`
	GeofenceID         *string `json:"geofence_id" gorm:"type:uuid"`
	GeofenceName       string  `json:"geofence_name" gorm:"type:varchar(255)"`
	GeofenceEvent      string  `json:"geofence_event" gorm:"type:varchar(50)"` // enter, exit, inside
	
	// Additional Data
	RawData     JSON    `json:"raw_data" gorm:"type:jsonb"` // Original GPS data from device
	ProcessedAt time.Time `json:"processed_at"`
	
	// Timestamps
	Timestamp   time.Time `json:"timestamp" gorm:"not null;index"` // GPS timestamp
	CreatedAt   time.Time `json:"created_at" gorm:"index"`

	// Relationships
	Vehicle Vehicle `json:"vehicle,omitempty" gorm:"foreignKey:VehicleID"`
	Driver  *Driver `json:"driver,omitempty" gorm:"foreignKey:DriverID"`
	Trip    *Trip   `json:"trip,omitempty" gorm:"foreignKey:TripID"`
}

// Trip represents a vehicle trip/journey
type Trip struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID string    `json:"company_id" gorm:"type:uuid;not null;index"`
	VehicleID string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	DriverID  *string   `json:"driver_id" gorm:"type:uuid;index"`
	
	// Trip Information
	Name        string    `json:"name" gorm:"type:varchar(255)"`
	Description string    `json:"description" gorm:"type:text"`
	Purpose     string    `json:"purpose" gorm:"type:varchar(100)"` // delivery, pickup, service, etc.
	
	// Trip Status
	Status      string    `json:"status" gorm:"type:varchar(20);default:'planned'"` // planned, active, completed, cancelled
	
	// Start Information
	StartLatitude  float64   `json:"start_latitude" gorm:"type:decimal(10,8)"`
	StartLongitude float64   `json:"start_longitude" gorm:"type:decimal(11,8)"`
	StartLocation  string    `json:"start_location" gorm:"type:varchar(255)"`
	StartTime      *time.Time `json:"start_time"`
	
	// End Information
	EndLatitude    float64   `json:"end_latitude" gorm:"type:decimal(10,8)"`
	EndLongitude   float64   `json:"end_longitude" gorm:"type:decimal(11,8)"`
	EndLocation    string    `json:"end_location" gorm:"type:varchar(255)"`
	EndTime        *time.Time `json:"end_time"`
	
	// Planned vs Actual
	PlannedStartTime *time.Time `json:"planned_start_time"`
	PlannedEndTime   *time.Time `json:"planned_end_time"`
	PlannedDistance  float64    `json:"planned_distance" gorm:"type:decimal(10,2)"` // km
	
	// Trip Statistics
	TotalDistance    float64 `json:"total_distance" gorm:"type:decimal(10,2);default:0"` // km
	TotalDuration    int     `json:"total_duration" gorm:"default:0"` // seconds
	AverageSpeed     float64 `json:"average_speed" gorm:"type:decimal(5,2);default:0"` // km/h
	MaxSpeed         float64 `json:"max_speed" gorm:"type:decimal(5,2);default:0"` // km/h
	IdleTime         int     `json:"idle_time" gorm:"default:0"` // seconds
	
	// Fuel Information
	FuelConsumed     float64 `json:"fuel_consumed" gorm:"type:decimal(8,2);default:0"` // liters
	FuelEfficiency   float64 `json:"fuel_efficiency" gorm:"type:decimal(5,2);default:0"` // km/liter
	StartFuelLevel   float64 `json:"start_fuel_level" gorm:"type:decimal(5,2)"` // liters
	EndFuelLevel     float64 `json:"end_fuel_level" gorm:"type:decimal(5,2)"` // liters
	
	// Violations and Events
	Violations       int     `json:"violations" gorm:"default:0"`
	HarshBraking     int     `json:"harsh_braking" gorm:"default:0"`
	RapidAcceleration int    `json:"rapid_acceleration" gorm:"default:0"`
	SharpCornering   int     `json:"sharp_cornering" gorm:"default:0"`
	SpeedingEvents   int     `json:"speeding_events" gorm:"default:0"`
	
	// Route Information
	RouteData    JSON    `json:"route_data" gorm:"type:jsonb"` // Route waypoints and data
	RouteOptimized bool  `json:"route_optimized" gorm:"default:false"`
	
	// Additional Data
	Notes        string    `json:"notes" gorm:"type:text"`
	WeatherData  JSON      `json:"weather_data" gorm:"type:jsonb"` // Weather conditions during trip
	
	// Timestamps
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Company   Company   `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Vehicle   Vehicle   `json:"vehicle,omitempty" gorm:"foreignKey:VehicleID"`
	Driver    *Driver   `json:"driver,omitempty" gorm:"foreignKey:DriverID"`
	GPSTracks []GPSTrack `json:"gps_tracks,omitempty" gorm:"foreignKey:TripID"`
}

// Geofence represents a geographical boundary
type Geofence struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID string    `json:"company_id" gorm:"type:uuid;not null;index"`
	
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Type        string    `json:"type" gorm:"type:varchar(50);not null"` // circle, polygon, rectangle
	
	// Geofence Geometry (PostGIS)
	CenterLatitude  float64 `json:"center_latitude" gorm:"type:decimal(10,8)"`
	CenterLongitude float64 `json:"center_longitude" gorm:"type:decimal(11,8)"`
	Radius          float64 `json:"radius" gorm:"type:decimal(8,2)"` // meters
	PolygonData     JSON    `json:"polygon_data" gorm:"type:jsonb"` // For complex polygons
	
	// Geofence Settings
	IsActive        bool    `json:"is_active" gorm:"default:true"`
	AlertOnEnter    bool    `json:"alert_on_enter" gorm:"default:true"`
	AlertOnExit     bool    `json:"alert_on_exit" gorm:"default:true"`
	SpeedLimit      float64 `json:"speed_limit" gorm:"type:decimal(5,2)"` // km/h
	
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Company Company `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
}

// TableName specifies the table name for the GPSTrack model
func (GPSTrack) TableName() string {
	return "gps_tracks"
}

// TableName specifies the table name for the Trip model
func (Trip) TableName() string {
	return "trips"
}

// TableName specifies the table name for the Geofence model
func (Geofence) TableName() string {
	return "geofences"
}

// BeforeCreate hook for GPSTrack
func (g *GPSTrack) BeforeCreate(tx *gorm.DB) error {
	if g.Country == "" {
		g.Country = "Indonesia"
	}
	g.ProcessedAt = time.Now()
	return nil
}

// BeforeCreate hook for Trip
func (t *Trip) BeforeCreate(tx *gorm.DB) error {
	if t.Status == "" {
		t.Status = "planned"
	}
	return nil
}

// BeforeCreate hook for Geofence
func (gf *Geofence) BeforeCreate(tx *gorm.DB) error {
	if gf.Type == "" {
		gf.Type = "circle"
	}
	return nil
}

// IsValidGPSTrack checks if GPS track data is valid
func (g *GPSTrack) IsValidGPSTrack() bool {
	// Basic coordinate validation
	if g.Latitude < -90 || g.Latitude > 90 {
		return false
	}
	if g.Longitude < -180 || g.Longitude > 180 {
		return false
	}
	if g.Speed < 0 || g.Speed > 300 { // Reasonable speed limit
		return false
	}
	return true
}

// CalculateDistance calculates distance from previous point
func (g *GPSTrack) CalculateDistance(prevLat, prevLon float64) float64 {
	// Haversine formula for distance calculation
	const earthRadius = 6371.0 // Earth radius in kilometers
	
	lat1Rad := prevLat * 3.14159265359 / 180
	lat2Rad := g.Latitude * 3.14159265359 / 180
	deltaLat := (g.Latitude - prevLat) * 3.14159265359 / 180
	deltaLon := (g.Longitude - prevLon) * 3.14159265359 / 180
	
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
		math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return earthRadius * c
}

// IsIdling checks if vehicle is idling
func (g *GPSTrack) IsIdling() bool {
	return g.Speed < 5 && g.IgnitionOn // Less than 5 km/h and ignition on
}

// IsMoving checks if vehicle is moving
func (g *GPSTrack) IsMoving() bool {
	return g.Speed > 5 // More than 5 km/h
}

// IsSpeeding checks if vehicle is speeding
func (g *GPSTrack) IsSpeeding() bool {
	return g.SpeedLimit > 0 && g.Speed > g.SpeedLimit
}

// GetViolationSeverity returns violation severity based on speed
func (g *GPSTrack) GetViolationSeverity() string {
	if !g.IsSpeeding() {
		return ""
	}
	
	overSpeed := g.Speed - g.SpeedLimit
	speedPercentage := (overSpeed / g.SpeedLimit) * 100
	
	switch {
	case speedPercentage >= 50:
		return "critical"
	case speedPercentage >= 30:
		return "high"
	case speedPercentage >= 15:
		return "medium"
	default:
		return "low"
	}
}

// Trip methods
// IsActive checks if trip is currently active
func (t *Trip) IsActive() bool {
	return t.Status == "active"
}

// IsCompleted checks if trip is completed
func (t *Trip) IsCompleted() bool {
	return t.Status == "completed"
}

// GetDuration returns trip duration in seconds
func (t *Trip) GetDuration() int {
	if t.StartTime == nil || t.EndTime == nil {
		return 0
	}
	return int(t.EndTime.Sub(*t.StartTime).Seconds())
}

// GetEfficiency returns fuel efficiency for the trip
func (t *Trip) GetEfficiency() float64 {
	if t.FuelConsumed <= 0 {
		return 0
	}
	return t.TotalDistance / t.FuelConsumed
}

// AddViolation adds a violation to trip statistics
func (t *Trip) AddViolation(violationType string) {
	t.Violations++
	switch violationType {
	case "harsh_braking":
		t.HarshBraking++
	case "rapid_acceleration":
		t.RapidAcceleration++
	case "sharp_cornering":
		t.SharpCornering++
	case "speeding":
		t.SpeedingEvents++
	}
}

// StartTrip starts the trip
func (t *Trip) StartTrip() {
	t.Status = "active"
	now := time.Now()
	t.StartTime = &now
}

// EndTrip ends the trip
func (t *Trip) EndTrip() {
	t.Status = "completed"
	now := time.Now()
	t.EndTime = &now
	t.TotalDuration = t.GetDuration()
	t.FuelEfficiency = t.GetEfficiency()
}

// Geofence methods
// IsPointInside checks if a point is inside the geofence
func (gf *Geofence) IsPointInside(lat, lon float64) bool {
	if !gf.IsActive {
		return false
	}
	
	switch gf.Type {
	case "circle":
		return gf.isPointInCircle(lat, lon)
	case "polygon":
		return gf.isPointInPolygon(lat, lon)
	case "rectangle":
		return gf.isPointInRectangle(lat, lon)
	default:
		return false
	}
}

// isPointInCircle checks if point is inside circle geofence
func (gf *Geofence) isPointInCircle(lat, lon float64) bool {
	// Haversine formula to calculate distance
	const earthRadius = 6371000 // Earth radius in meters
	
	lat1Rad := gf.CenterLatitude * 3.14159265359 / 180
	lat2Rad := lat * 3.14159265359 / 180
	deltaLat := (lat - gf.CenterLatitude) * 3.14159265359 / 180
	deltaLon := (lon - gf.CenterLongitude) * 3.14159265359 / 180
	
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
		math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	distance := earthRadius * c
	return distance <= gf.Radius
}

// isPointInPolygon checks if point is inside polygon geofence
func (gf *Geofence) isPointInPolygon(lat, lon float64) bool {
	// Implementation for polygon checking
	// This would require parsing the polygon data from JSON
	return false // Placeholder
}

// isPointInRectangle checks if point is inside rectangle geofence
func (gf *Geofence) isPointInRectangle(lat, lon float64) bool {
	// Implementation for rectangle checking
	// This would require parsing rectangle bounds from polygon data
	return false // Placeholder
}
