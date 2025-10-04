package models

import (
	"time"

	"gorm.io/gorm"
)

// Driver represents a fleet driver
type Driver struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID string    `json:"company_id" gorm:"type:uuid;not null;index"`
	VehicleID *string   `json:"vehicle_id" gorm:"type:uuid;index"` // Current assigned vehicle
	
	// Personal Information
	FirstName   string    `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName    string    `json:"last_name" gorm:"type:varchar(100);not null"`
	Email       string    `json:"email" gorm:"type:varchar(255);unique"`
	Phone       string    `json:"phone" gorm:"type:varchar(20);not null"`
	Avatar      string    `json:"avatar" gorm:"type:varchar(500)"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender" gorm:"type:varchar(10)"` // male, female
	
	// Indonesian Fields
	NIK         string    `json:"nik" gorm:"type:varchar(16);unique;not null"` // Indonesian ID
	Address     string    `json:"address" gorm:"type:text"`
	City        string    `json:"city" gorm:"type:varchar(100)"`
	Province    string    `json:"province" gorm:"type:varchar(100)"`
	PostalCode  string    `json:"postal_code" gorm:"type:varchar(10)"`
	
	// Driver License Information (Indonesian SIM)
	SIMNumber   string    `json:"sim_number" gorm:"type:varchar(20);unique;not null"` // Driver's License Number
	SIMType     string    `json:"sim_type" gorm:"type:varchar(10);not null"`         // A, B1, B2, C, D
	SIMExpiry   *time.Time `json:"sim_expiry"`
	SIMClass    string    `json:"sim_class" gorm:"type:varchar(50)"`               // Private, Commercial, etc.
	
	// Employment Information
	EmployeeID    string    `json:"employee_id" gorm:"type:varchar(50);unique"`
	Position      string    `json:"position" gorm:"type:varchar(100)"`           // Driver, Senior Driver, Supervisor
	Department    string    `json:"department" gorm:"type:varchar(100)"`
	HireDate      *time.Time `json:"hire_date"`
	Salary        float64   `json:"salary" gorm:"type:decimal(12,2)"`           // in IDR
	EmploymentStatus string  `json:"employment_status" gorm:"type:varchar(20);default:'active'"` // active, inactive, terminated
	
	// Performance and Scoring
	PerformanceScore float64 `json:"performance_score" gorm:"type:decimal(5,2);default:100.00"` // 0-100 scale
	SafetyScore      float64 `json:"safety_score" gorm:"type:decimal(5,2);default:100.00"`      // 0-100 scale
	EfficiencyScore  float64 `json:"efficiency_score" gorm:"type:decimal(5,2);default:100.00"`  // 0-100 scale
	OverallScore     float64 `json:"overall_score" gorm:"type:decimal(5,2);default:100.00"`     // 0-100 scale
	
	// Driving Statistics
	TotalTrips       int     `json:"total_trips" gorm:"default:0"`
	TotalDistance    float64 `json:"total_distance" gorm:"type:decimal(10,2);default:0"` // in km
	TotalHours       float64 `json:"total_hours" gorm:"type:decimal(8,2);default:0"`    // in hours
	AverageSpeed     float64 `json:"average_speed" gorm:"type:decimal(5,2);default:0"`  // km/h
	
	// Violations and Incidents
	Violations       int     `json:"violations" gorm:"default:0"`
	Accidents        int     `json:"accidents" gorm:"default:0"`
	LastViolationAt  *time.Time `json:"last_violation_at"`
	LastAccidentAt   *time.Time `json:"last_accident_at"`
	
	// Compliance and Training
	TrainingCompleted bool      `json:"training_completed" gorm:"default:false"`
	LastTrainingDate  *time.Time `json:"last_training_date"`
	NextTrainingDate  *time.Time `json:"next_training_date"`
	MedicalCheckupExpiry *time.Time `json:"medical_checkup_expiry"`
	
	// Status and Availability
	Status          string    `json:"status" gorm:"type:varchar(20);default:'available'"` // available, busy, offline, inactive
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	IsVerified      bool      `json:"is_verified" gorm:"default:false"`
	LastActiveAt    *time.Time `json:"last_active_at"`
	
	// Emergency Contacts
	EmergencyContact1Name  string `json:"emergency_contact1_name" gorm:"type:varchar(100)"`
	EmergencyContact1Phone string `json:"emergency_contact1_phone" gorm:"type:varchar(20)"`
	EmergencyContact1Relation string `json:"emergency_contact1_relation" gorm:"type:varchar(50)"`
	EmergencyContact2Name  string `json:"emergency_contact2_name" gorm:"type:varchar(100)"`
	EmergencyContact2Phone string `json:"emergency_contact2_phone" gorm:"type:varchar(20)"`
	EmergencyContact2Relation string `json:"emergency_contact2_relation" gorm:"type:varchar(50)"`
	
	// Settings and Preferences
	Settings        JSON     `json:"settings" gorm:"type:jsonb"`                   // Driver-specific settings
	Preferences     JSON     `json:"preferences" gorm:"type:jsonb"`               // Driver preferences
	
	// Timestamps
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Company       Company       `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
	Vehicle       *Vehicle      `json:"vehicle,omitempty" gorm:"foreignKey:VehicleID"`
	Trips         []Trip        `json:"trips,omitempty" gorm:"foreignKey:DriverID"`
	DriverEvents  []DriverEvent `json:"driver_events,omitempty" gorm:"foreignKey:DriverID"`
	PerformanceLogs []PerformanceLog `json:"performance_logs,omitempty" gorm:"foreignKey:DriverID"`
}

// DriverEvent represents driver behavior events
type DriverEvent struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DriverID    string    `json:"driver_id" gorm:"type:uuid;not null;index"`
	VehicleID   string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	TripID      *string   `json:"trip_id" gorm:"type:uuid;index"`
	
	EventType   string    `json:"event_type" gorm:"type:varchar(50);not null"` // harsh_braking, rapid_acceleration, sharp_cornering, speeding, idling
	Severity    string    `json:"severity" gorm:"type:varchar(20);not null"`   // low, medium, high, critical
	Description string    `json:"description" gorm:"type:text"`
	
	// Location and Context
	Latitude    float64   `json:"latitude" gorm:"type:decimal(10,8)"`
	Longitude   float64   `json:"longitude" gorm:"type:decimal(11,8)"`
	Location    string    `json:"location" gorm:"type:varchar(255)"`
	Speed       float64   `json:"speed" gorm:"type:decimal(5,2)"` // km/h
	
	// Event Data
	Data        JSON      `json:"data" gorm:"type:jsonb"` // Additional event-specific data
	IsProcessed bool      `json:"is_processed" gorm:"default:false"`
	
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	Driver  Driver  `json:"driver,omitempty" gorm:"foreignKey:DriverID"`
	Vehicle Vehicle `json:"vehicle,omitempty" gorm:"foreignKey:VehicleID"`
	Trip    *Trip   `json:"trip,omitempty" gorm:"foreignKey:TripID"`
}

// PerformanceLog represents driver performance tracking
type PerformanceLog struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DriverID        string    `json:"driver_id" gorm:"type:uuid;not null;index"`
	Period          string    `json:"period" gorm:"type:varchar(20);not null"` // daily, weekly, monthly
	
	// Performance Metrics
	PerformanceScore float64 `json:"performance_score" gorm:"type:decimal(5,2)"`
	SafetyScore      float64 `json:"safety_score" gorm:"type:decimal(5,2)"`
	EfficiencyScore  float64 `json:"efficiency_score" gorm:"type:decimal(5,2)"`
	OverallScore     float64 `json:"overall_score" gorm:"type:decimal(5,2)"`
	
	// Driving Statistics
	TripsCompleted   int     `json:"trips_completed" gorm:"default:0"`
	DistanceTraveled float64 `json:"distance_traveled" gorm:"type:decimal(10,2)"`
	HoursDriven      float64 `json:"hours_driven" gorm:"type:decimal(8,2)"`
	AverageSpeed     float64 `json:"average_speed" gorm:"type:decimal(5,2)"`
	
	// Violations and Events
	Violations       int     `json:"violations" gorm:"default:0"`
	Accidents        int     `json:"accidents" gorm:"default:0"`
	HarshBraking     int     `json:"harsh_braking" gorm:"default:0"`
	RapidAcceleration int    `json:"rapid_acceleration" gorm:"default:0"`
	SharpCornering   int     `json:"sharp_cornering" gorm:"default:0"`
	SpeedingEvents   int     `json:"speeding_events" gorm:"default:0"`
	
	// Fuel Efficiency
	FuelConsumed     float64 `json:"fuel_consumed" gorm:"type:decimal(8,2)"` // in liters
	FuelEfficiency   float64 `json:"fuel_efficiency" gorm:"type:decimal(5,2)"` // km/liter
	
	// Period Information
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	CreatedAt        time.Time `json:"created_at"`

	// Relationships
	Driver Driver `json:"driver,omitempty" gorm:"foreignKey:DriverID"`
}

// TableName specifies the table name for the Driver model
func (Driver) TableName() string {
	return "drivers"
}

// TableName specifies the table name for the DriverEvent model
func (DriverEvent) TableName() string {
	return "driver_events"
}

// TableName specifies the table name for the PerformanceLog model
func (PerformanceLog) TableName() string {
	return "performance_logs"
}

// BeforeCreate hook to set default values
func (d *Driver) BeforeCreate(tx *gorm.DB) error {
	if d.Status == "" {
		d.Status = "available"
	}
	if d.EmploymentStatus == "" {
		d.EmploymentStatus = "active"
	}
	if d.PerformanceScore == 0 {
		d.PerformanceScore = 100.00
	}
	if d.SafetyScore == 0 {
		d.SafetyScore = 100.00
	}
	if d.EfficiencyScore == 0 {
		d.EfficiencyScore = 100.00
	}
	if d.OverallScore == 0 {
		d.OverallScore = 100.00
	}
	return nil
}

// GetFullName returns the driver's full name
func (d *Driver) GetFullName() string {
	return d.FirstName + " " + d.LastName
}

// GetDisplayName returns a formatted driver name for display
func (d *Driver) GetDisplayName() string {
	return d.GetFullName() + " (" + d.SIMNumber + ")"
}

// IsValidNIK validates Indonesian NIK format (16 digits)
func (d *Driver) IsValidNIK() bool {
	if len(d.NIK) != 16 {
		return false
	}
	// Additional NIK validation logic can be added here
	return true
}

// IsValidSIM validates Indonesian SIM format
func (d *Driver) IsValidSIM() bool {
	if len(d.SIMNumber) < 10 {
		return false
	}
	// Additional SIM validation logic can be added here
	return true
}

// IsSIMExpired checks if the driver's SIM is expired
func (d *Driver) IsSIMExpired() bool {
	if d.SIMExpiry == nil {
		return false
	}
	return time.Now().After(*d.SIMExpiry)
}

// IsTrainingDue checks if driver needs training
func (d *Driver) IsTrainingDue() bool {
	if d.NextTrainingDate == nil {
		return false
	}
	return time.Now().After(*d.NextTrainingDate)
}

// IsMedicalCheckupExpired checks if medical checkup is expired
func (d *Driver) IsMedicalCheckupExpired() bool {
	if d.MedicalCheckupExpiry == nil {
		return false
	}
	return time.Now().After(*d.MedicalCheckupExpiry)
}

// CanDrive checks if driver is eligible to drive
func (d *Driver) CanDrive() bool {
	return d.IsActive && 
		   d.Status == "available" && 
		   d.EmploymentStatus == "active" &&
		   !d.IsSIMExpired() &&
		   !d.IsMedicalCheckupExpired() &&
		   d.TrainingCompleted
}

// UpdatePerformanceScore updates the driver's performance scores
func (d *Driver) UpdatePerformanceScore(performance, safety, efficiency float64) {
	d.PerformanceScore = performance
	d.SafetyScore = safety
	d.EfficiencyScore = efficiency
	d.OverallScore = (performance + safety + efficiency) / 3
}

// AddViolation adds a violation to the driver's record
func (d *Driver) AddViolation(severity string) {
	d.Violations++
	now := time.Now()
	d.LastViolationAt = &now
	
	// Adjust performance score based on severity
	switch severity {
	case "critical":
		d.SafetyScore -= 10
	case "high":
		d.SafetyScore -= 5
	case "medium":
		d.SafetyScore -= 2
	case "low":
		d.SafetyScore -= 1
	}
	
	if d.SafetyScore < 0 {
		d.SafetyScore = 0
	}
	
	d.OverallScore = (d.PerformanceScore + d.SafetyScore + d.EfficiencyScore) / 3
}

// AddAccident adds an accident to the driver's record
func (d *Driver) AddAccident() {
	d.Accidents++
	now := time.Now()
	d.LastAccidentAt = &now
	d.SafetyScore -= 20 // Significant impact on safety score
	
	if d.SafetyScore < 0 {
		d.SafetyScore = 0
	}
	
	d.OverallScore = (d.PerformanceScore + d.SafetyScore + d.EfficiencyScore) / 3
}

// AssignVehicle assigns a vehicle to the driver
func (d *Driver) AssignVehicle(vehicleID string) {
	d.VehicleID = &vehicleID
	d.Status = "busy"
}

// UnassignVehicle removes the vehicle assignment
func (d *Driver) UnassignVehicle() {
	d.VehicleID = nil
	d.Status = "available"
}

// UpdateLastActive updates the last active timestamp
func (d *Driver) UpdateLastActive() {
	now := time.Now()
	d.LastActiveAt = &now
}

// GetPerformanceGrade returns a letter grade based on overall score
func (d *Driver) GetPerformanceGrade() string {
	switch {
	case d.OverallScore >= 90:
		return "A"
	case d.OverallScore >= 80:
		return "B"
	case d.OverallScore >= 70:
		return "C"
	case d.OverallScore >= 60:
		return "D"
	default:
		return "F"
	}
}

// IsHighPerformer checks if driver is a high performer
func (d *Driver) IsHighPerformer() bool {
	return d.OverallScore >= 85
}

// NeedsAttention checks if driver needs management attention
func (d *Driver) NeedsAttention() bool {
	return d.OverallScore < 70 || d.Violations > 5 || d.Accidents > 0
}
