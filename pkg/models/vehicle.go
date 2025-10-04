package models

import (
	"time"

	"gorm.io/gorm"
)

// Vehicle represents a fleet vehicle
type Vehicle struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID string    `json:"company_id" gorm:"type:uuid;not null;index"`
	DriverID  *string   `json:"driver_id" gorm:"type:uuid;index"` // Optional current driver
	
	// Vehicle Identification
	LicensePlate string    `json:"license_plate" gorm:"type:varchar(20);unique;not null"`
	VIN          string    `json:"vin" gorm:"type:varchar(17);unique"` // Vehicle Identification Number
	ChassisNumber string   `json:"chassis_number" gorm:"type:varchar(50)"`
	EngineNumber  string   `json:"engine_number" gorm:"type:varchar(50)"`
	
	// Vehicle Information
	Make         string    `json:"make" gorm:"type:varchar(100);not null"`          // Toyota, Honda, etc.
	Model        string    `json:"model" gorm:"type:varchar(100);not null"`         // Camry, Civic, etc.
	Year         int       `json:"year" gorm:"not null"`
	Color        string    `json:"color" gorm:"type:varchar(50)"`
	Type         string    `json:"type" gorm:"type:varchar(50);not null"`           // sedan, truck, bus, motorcycle
	Category     string    `json:"category" gorm:"type:varchar(50)"`                // commercial, passenger, cargo
	
	// Indonesian Vehicle Registration
	STNK         string    `json:"stnk" gorm:"type:varchar(50)"`                    // Vehicle Registration Certificate
	BPKB         string    `json:"bpkb" gorm:"type:varchar(50)"`                    // Vehicle Ownership Certificate
	Pajak        string    `json:"pajak" gorm:"type:varchar(50)"`                   // Tax Certificate
	STNKExpiry   *time.Time `json:"stnk_expiry"`
	PajakExpiry  *time.Time `json:"pajak_expiry"`
	
	// Vehicle Specifications
	EngineCapacity    float64 `json:"engine_capacity" gorm:"type:decimal(5,2)"`     // in liters
	FuelType          string  `json:"fuel_type" gorm:"type:varchar(20)"`           // petrol, diesel, electric, hybrid
	TankCapacity      float64 `json:"tank_capacity" gorm:"type:decimal(8,2)"`       // in liters
	SeatingCapacity   int     `json:"seating_capacity" gorm:"default:1"`
	CargoCapacity     float64 `json:"cargo_capacity" gorm:"type:decimal(8,2)"`      // in kg
	Weight           float64 `json:"weight" gorm:"type:decimal(8,2)"`              // in kg
	
	// Status and Management
	Status          string    `json:"status" gorm:"type:varchar(20);default:'active'"` // active, maintenance, retired, inactive
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	IsGPSEnabled    bool      `json:"is_gps_enabled" gorm:"default:true"`
	IsFuelTrackingEnabled bool `json:"is_fuel_tracking_enabled" gorm:"default:true"`
	
	// Location and Tracking
	LastLatitude    float64   `json:"last_latitude" gorm:"type:decimal(10,8)"`
	LastLongitude   float64   `json:"last_longitude" gorm:"type:decimal(11,8)"`
	LastLocation    string    `json:"last_location" gorm:"type:varchar(255)"`
	LastUpdatedAt   *time.Time `json:"last_updated_at"`
	
	// Fuel Information
	CurrentFuelLevel float64 `json:"current_fuel_level" gorm:"type:decimal(5,2)"`    // in liters
	FuelEfficiency   float64 `json:"fuel_efficiency" gorm:"type:decimal(5,2)"`       // km/liter
	TotalDistance    float64 `json:"total_distance" gorm:"type:decimal(10,2)"`       // in km
	OdometerReading  float64 `json:"odometer_reading" gorm:"type:decimal(10,2)"`     // in km
	
	// Maintenance
	LastServiceDate  *time.Time `json:"last_service_date"`
	NextServiceDate  *time.Time `json:"next_service_date"`
	ServiceInterval  int        `json:"service_interval" gorm:"default:10000"`       // km
	MaintenanceCost  float64    `json:"maintenance_cost" gorm:"type:decimal(12,2)"` // in IDR
	
	// Insurance
	InsuranceProvider string    `json:"insurance_provider" gorm:"type:varchar(100)"`
	InsuranceNumber   string    `json:"insurance_number" gorm:"type:varchar(50)"`
	InsuranceExpiry   *time.Time `json:"insurance_expiry"`
	
	// Cost and Pricing
	PurchasePrice     float64 `json:"purchase_price" gorm:"type:decimal(15,2)"`     // in IDR
	CurrentValue      float64 `json:"current_value" gorm:"type:decimal(15,2)"`      // in IDR
	DepreciationRate  float64 `json:"depreciation_rate" gorm:"type:decimal(5,2)"`   // percentage per year
	
	// Settings and Configuration
	Settings          JSON    `json:"settings" gorm:"type:jsonb"`                   // Vehicle-specific settings
	GeofenceSettings  JSON    `json:"geofence_settings" gorm:"type:jsonb"`         // Geofencing configuration
	
	// Timestamps
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Company    Company    `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Driver     *Driver    `json:"driver,omitempty" gorm:"foreignKey:DriverID"`
	GPSTracks  []GPSTrack `json:"gps_tracks,omitempty" gorm:"foreignKey:VehicleID"`
	Trips      []Trip     `json:"trips,omitempty" gorm:"foreignKey:VehicleID"`
	MaintenanceLogs []MaintenanceLog `json:"maintenance_logs,omitempty" gorm:"foreignKey:VehicleID"`
	FuelLogs   []FuelLog  `json:"fuel_logs,omitempty" gorm:"foreignKey:VehicleID"`
}

// MaintenanceLog represents vehicle maintenance records
type MaintenanceLog struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VehicleID   string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	MaintenanceType string `json:"maintenance_type" gorm:"type:varchar(50);not null"` // service, repair, inspection
	Description string    `json:"description" gorm:"type:text"`
	Cost        float64   `json:"cost" gorm:"type:decimal(12,2)"` // in IDR
	OdometerReading float64 `json:"odometer_reading" gorm:"type:decimal(10,2)"`
	PerformedBy string    `json:"performed_by" gorm:"type:varchar(100)"`
	Location    string    `json:"location" gorm:"type:varchar(255)"`
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	Vehicle Vehicle `json:"vehicle,omitempty" gorm:"foreignKey:VehicleID"`
}

// FuelLog represents fuel consumption records
type FuelLog struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VehicleID       string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	FuelType        string    `json:"fuel_type" gorm:"type:varchar(20);not null"`
	Quantity        float64   `json:"quantity" gorm:"type:decimal(8,2);not null"` // in liters
	Cost            float64   `json:"cost" gorm:"type:decimal(12,2)"` // in IDR
	PricePerLiter   float64   `json:"price_per_liter" gorm:"type:decimal(8,2)"` // in IDR
	StationName     string    `json:"station_name" gorm:"type:varchar(255)"`
	Location        string    `json:"location" gorm:"type:varchar(255)"`
	OdometerReading float64   `json:"odometer_reading" gorm:"type:decimal(10,2)"`
	CreatedAt       time.Time `json:"created_at"`

	// Relationships
	Vehicle Vehicle `json:"vehicle,omitempty" gorm:"foreignKey:VehicleID"`
}

// TableName specifies the table name for the Vehicle model
func (Vehicle) TableName() string {
	return "vehicles"
}

// TableName specifies the table name for the MaintenanceLog model
func (MaintenanceLog) TableName() string {
	return "maintenance_logs"
}

// TableName specifies the table name for the FuelLog model
func (FuelLog) TableName() string {
	return "fuel_logs"
}

// BeforeCreate hook to set default values
func (v *Vehicle) BeforeCreate(tx *gorm.DB) error {
	if v.Status == "" {
		v.Status = "active"
	}
	if v.ServiceInterval == 0 {
		v.ServiceInterval = 10000 // 10,000 km default
	}
	return nil
}

// GetDisplayName returns a formatted vehicle name for display
func (v *Vehicle) GetDisplayName() string {
	return v.LicensePlate + " - " + v.Make + " " + v.Model + " (" + string(rune(v.Year)) + ")"
}

// GetFullName returns the complete vehicle identification
func (v *Vehicle) GetFullName() string {
	return v.Make + " " + v.Model + " " + v.LicensePlate
}

// IsValidSTNK validates Indonesian STNK format
func (v *Vehicle) IsValidSTNK() bool {
	if len(v.STNK) < 10 {
		return false
	}
	// Additional STNK validation logic can be added here
	return true
}

// IsValidBPKB validates Indonesian BPKB format
func (v *Vehicle) IsValidBPKB() bool {
	if len(v.BPKB) < 10 {
		return false
	}
	// Additional BPKB validation logic can be added here
	return true
}

// IsMaintenanceDue checks if vehicle is due for maintenance
func (v *Vehicle) IsMaintenanceDue() bool {
	if v.NextServiceDate != nil {
		return time.Now().After(*v.NextServiceDate)
	}
	if v.LastServiceDate != nil && v.OdometerReading > 0 {
		nextServiceOdometer := v.OdometerReading + float64(v.ServiceInterval)
		return v.TotalDistance >= nextServiceOdometer
	}
	return false
}

// UpdateLocation updates the vehicle's current location
func (v *Vehicle) UpdateLocation(latitude, longitude float64, location string) {
	v.LastLatitude = latitude
	v.LastLongitude = longitude
	v.LastLocation = location
	now := time.Now()
	v.LastUpdatedAt = &now
}

// AddDistance adds distance to the total distance and updates odometer
func (v *Vehicle) AddDistance(distance float64) {
	v.TotalDistance += distance
	v.OdometerReading += distance
}

// UpdateFuelLevel updates the current fuel level
func (v *Vehicle) UpdateFuelLevel(level float64) {
	v.CurrentFuelLevel = level
}

// CalculateFuelEfficiency calculates fuel efficiency based on distance and fuel consumed
func (v *Vehicle) CalculateFuelEfficiency(distance, fuelConsumed float64) {
	if fuelConsumed > 0 {
		v.FuelEfficiency = distance / fuelConsumed
	}
}

// GetCurrentValue calculates the current value based on depreciation
func (v *Vehicle) GetCurrentValue() float64 {
	if v.PurchasePrice == 0 || v.DepreciationRate == 0 {
		return v.CurrentValue
	}
	
	years := time.Since(v.CreatedAt).Hours() / (24 * 365)
	depreciation := v.PurchasePrice * (v.DepreciationRate / 100) * years
	currentValue := v.PurchasePrice - depreciation
	
	if currentValue < 0 {
		currentValue = 0
	}
	
	return currentValue
}

// CanTrackGPS checks if GPS tracking is enabled for the vehicle
func (v *Vehicle) CanTrackGPS() bool {
	return v.IsGPSEnabled && v.IsActive && v.Status == "active"
}

// CanBeAssigned checks if vehicle can be assigned to a driver
func (v *Vehicle) CanBeAssigned() bool {
	return v.IsActive && v.Status == "active" && v.DriverID == nil
}

// AssignDriver assigns a driver to the vehicle
func (v *Vehicle) AssignDriver(driverID string) {
	v.DriverID = &driverID
}

// UnassignDriver removes the current driver assignment
func (v *Vehicle) UnassignDriver() {
	v.DriverID = nil
}
