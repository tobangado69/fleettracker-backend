package driver

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	apperrors "github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// Service handles driver operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new driver service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// DriverStatus represents the status of a driver
type DriverStatus string

const (
	StatusAvailable   DriverStatus = "available"
	StatusBusy        DriverStatus = "busy"
	StatusInactive    DriverStatus = "inactive"
	StatusSuspended   DriverStatus = "suspended"
	StatusTerminated  DriverStatus = "terminated"
)

// EmploymentStatus represents the employment status of a driver
type EmploymentStatus string

const (
	EmploymentActive      EmploymentStatus = "active"
	EmploymentSuspended   EmploymentStatus = "suspended"
	EmploymentTerminated  EmploymentStatus = "terminated"
)

// PerformanceGrade represents the performance grade of a driver
type PerformanceGrade string

const (
	GradeA PerformanceGrade = "A" // 90-100
	GradeB PerformanceGrade = "B" // 80-89
	GradeC PerformanceGrade = "C" // 70-79
	GradeD PerformanceGrade = "D" // 60-69
	GradeF PerformanceGrade = "F" // 0-59
)

// CreateDriverRequest represents the request to create a driver
type CreateDriverRequest struct {
	FirstName             string     `json:"first_name" validate:"required,min=2,max=100"`
	LastName              string     `json:"last_name" validate:"required,min=2,max=100"`
	PhoneNumber           string     `json:"phone_number" validate:"required,min=10,max=15"`
	Email                 string     `json:"email" validate:"required,email"`
	Address               string     `json:"address" validate:"required,min=10,max=255"`
	City                  string     `json:"city" validate:"required,min=2,max=100"`
	Province              string     `json:"province" validate:"required,min=2,max=100"`
	DateOfBirth           time.Time  `json:"date_of_birth" validate:"required"`
	HireDate              time.Time  `json:"hire_date" validate:"required"`
	
	// Indonesian Compliance Fields
	NIK                   string     `json:"nik" validate:"required,len=16"`
	SIMNumber             string     `json:"sim_number" validate:"required,min=10,max=20"`
	SIMExpiry             *time.Time `json:"sim_expiry"`
	MedicalCheckupDate    *time.Time `json:"medical_checkup_date"`
	TrainingCompleted     bool       `json:"training_completed"`
	TrainingExpiry        *time.Time `json:"training_expiry"`
}

// UpdateDriverRequest represents the request to update a driver
type UpdateDriverRequest struct {
	FirstName             *string    `json:"first_name,omitempty" validate:"omitempty,min=2,max=100"`
	LastName              *string    `json:"last_name,omitempty" validate:"omitempty,min=2,max=100"`
	PhoneNumber           *string    `json:"phone_number,omitempty" validate:"omitempty,min=10,max=15"`
	Email                 *string    `json:"email,omitempty" validate:"omitempty,email"`
	Address               *string    `json:"address,omitempty" validate:"omitempty,min=10,max=255"`
	City                  *string    `json:"city,omitempty" validate:"omitempty,min=2,max=100"`
	Province              *string    `json:"province,omitempty" validate:"omitempty,min=2,max=100"`
	DateOfBirth           *time.Time `json:"date_of_birth,omitempty"`
	HireDate              *time.Time `json:"hire_date,omitempty"`
	Status                *string    `json:"status,omitempty" validate:"omitempty,oneof=available busy inactive suspended terminated"`
	EmploymentStatus      *string    `json:"employment_status,omitempty" validate:"omitempty,oneof=active suspended terminated"`
	IsActive              *bool      `json:"is_active,omitempty"`
	
	// Indonesian Compliance Fields
	NIK                   *string    `json:"nik,omitempty" validate:"omitempty,len=16"`
	SIMNumber             *string    `json:"sim_number,omitempty" validate:"omitempty,min=10,max=20"`
	SIMExpiry             *time.Time `json:"sim_expiry,omitempty"`
	MedicalCheckupDate    *time.Time `json:"medical_checkup_date,omitempty"`
	TrainingCompleted     *bool      `json:"training_completed,omitempty"`
	TrainingExpiry        *time.Time `json:"training_expiry,omitempty"`
	
	// Performance Fields
	PerformanceScore      *float64   `json:"performance_score,omitempty" validate:"omitempty,min=0,max=100"`
	SafetyScore           *float64   `json:"safety_score,omitempty" validate:"omitempty,min=0,max=100"`
	EfficiencyScore       *float64   `json:"efficiency_score,omitempty" validate:"omitempty,min=0,max=100"`
}

// DriverFilters represents filters for listing drivers
type DriverFilters struct {
	Status            *string `json:"status" form:"status"`
	EmploymentStatus  *string `json:"employment_status" form:"employment_status"`
	PerformanceGrade  *string `json:"performance_grade" form:"performance_grade"`
	City              *string `json:"city" form:"city"`
	Province          *string `json:"province" form:"province"`
	HasVehicle        *bool   `json:"has_vehicle" form:"has_vehicle"`
	IsAvailable       *bool   `json:"is_available" form:"is_available"`
	IsCompliant       *bool   `json:"is_compliant" form:"is_compliant"`
	Search            *string `json:"search" form:"search"`
	
	// Pagination
	Page              int     `json:"page" form:"page" validate:"min=1"`
	Limit             int     `json:"limit" form:"limit" validate:"min=1,max=100"`
	SortBy            string  `json:"sort_by" form:"sort_by" validate:"oneof=created_at updated_at first_name last_name performance_score hire_date"`
	SortOrder         string  `json:"sort_order" form:"sort_order" validate:"oneof=asc desc"`
}

// DriverResponse represents the response for driver data
type DriverResponse struct {
	ID                    string     `json:"id"`
	CompanyID             string     `json:"company_id"`
	UserID                *string    `json:"user_id"`
	FirstName             string     `json:"first_name"`
	LastName              string     `json:"last_name"`
	PhoneNumber           string     `json:"phone_number"`
	Email                 string     `json:"email"`
	Address               string     `json:"address"`
	City                  string     `json:"city"`
	Province              string     `json:"province"`
	DateOfBirth           time.Time  `json:"date_of_birth"`
	HireDate              time.Time  `json:"hire_date"`
	Status                string     `json:"status"`
	EmploymentStatus      string     `json:"employment_status"`
	IsActive              bool       `json:"is_active"`
	
	// Indonesian Compliance Fields
	NIK                   string     `json:"nik"`
	SIMNumber             string     `json:"sim_number"`
	SIMExpiry             *time.Time `json:"sim_expiry"`
	MedicalCheckupDate    *time.Time `json:"medical_checkup_date"`
	TrainingCompleted     bool       `json:"training_completed"`
	TrainingExpiry        *time.Time `json:"training_expiry"`
	
	// Performance Fields
	PerformanceScore      float64    `json:"performance_score"`
	SafetyScore           float64    `json:"safety_score"`
	EfficiencyScore       float64    `json:"efficiency_score"`
	OverallScore          float64    `json:"overall_score"`
	PerformanceGrade      string     `json:"performance_grade"`
	
	// Relationships
	Vehicle               *models.Vehicle `json:"vehicle,omitempty"`
	
	// Timestamps
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// CreateDriver creates a new driver
func (s *Service) CreateDriver(companyID string, req CreateDriverRequest) (*models.Driver, error) {
	// Validate Indonesian compliance fields
	if err := s.validateIndonesianCompliance(req.NIK, req.SIMNumber, req.DateOfBirth); err != nil {
		return nil, err
	}

	// Check if NIK already exists
	var existingDriver models.Driver
	if err := s.db.Where("nik = ?", req.NIK).First(&existingDriver).Error; err == nil {
		return nil, apperrors.NewConflictError("driver with this NIK already exists")
	}

	// Check if SIM number already exists
	if err := s.db.Where("sim_number = ?", req.SIMNumber).First(&existingDriver).Error; err == nil {
		return nil, apperrors.NewConflictError("driver with this SIM number already exists")
	}

	// Check if email already exists
	if err := s.db.Where("email = ?", req.Email).First(&existingDriver).Error; err == nil {
		return nil, apperrors.NewConflictError("driver with this email already exists")
	}

	// Calculate age from date of birth
	age := time.Now().Year() - req.DateOfBirth.Year()
	if age < 18 {
		return nil, apperrors.NewValidationError("driver must be at least 18 years old")
	}

	// Create driver
	driver := &models.Driver{
		CompanyID:             companyID,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Phone:                 req.PhoneNumber,
		Email:                 req.Email,
		Address:               req.Address,
		City:                  req.City,
		Province:              req.Province,
		DateOfBirth:           &req.DateOfBirth,
		HireDate:              &req.HireDate,
		Status:                string(StatusAvailable),
		EmploymentStatus:      string(EmploymentActive),
		IsActive:              true,
		NIK:                   req.NIK,
		SIMNumber:             req.SIMNumber,
		SIMExpiry:             req.SIMExpiry,
		MedicalCheckupExpiry:  req.MedicalCheckupDate,
		TrainingCompleted:     req.TrainingCompleted,
		NextTrainingDate:      req.TrainingExpiry,
		PerformanceScore:      100.0, // Start with perfect score
		SafetyScore:           100.0,
		EfficiencyScore:       100.0,
		OverallScore:          100.0,
	}

	// Save to database
	if err := s.db.Create(driver).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to create driver")
	}

	return driver, nil
}

// GetDriver retrieves a driver by ID
func (s *Service) GetDriver(companyID, driverID string) (*models.Driver, error) {
	var driver models.Driver
	
	if err := s.db.Preload("Vehicle").Where("company_id = ? AND id = ?", companyID, driverID).First(&driver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("driver")
		}
		return nil, apperrors.Wrap(err, "failed to get driver")
	}

	return &driver, nil
}

// UpdateDriver updates a driver
func (s *Service) UpdateDriver(companyID, driverID string, req UpdateDriverRequest) (*models.Driver, error) {
	// Get existing driver
	driver, err := s.GetDriver(companyID, driverID)
	if err != nil {
		return nil, err
	}

	// Validate Indonesian compliance fields if provided
	if req.NIK != nil || req.SIMNumber != nil || req.DateOfBirth != nil {
		nik := driver.NIK
		sim := driver.SIMNumber
		dob := time.Time{}
		if driver.DateOfBirth != nil {
			dob = *driver.DateOfBirth
		}
		
		if req.NIK != nil {
			nik = *req.NIK
		}
		if req.SIMNumber != nil {
			sim = *req.SIMNumber
		}
		if req.DateOfBirth != nil {
			dob = *req.DateOfBirth
		}
		
		if err := s.validateIndonesianCompliance(nik, sim, dob); err != nil {
			return nil, err
		}
	}

	// Check for duplicate NIK if being updated
	if req.NIK != nil && *req.NIK != driver.NIK {
		var existingDriver models.Driver
		if err := s.db.Where("nik = ? AND id != ?", *req.NIK, driverID).First(&existingDriver).Error; err == nil {
			return nil, apperrors.NewConflictError("driver with this NIK already exists")
		}
	}

	// Check for duplicate SIM if being updated
	if req.SIMNumber != nil && *req.SIMNumber != driver.SIMNumber {
		var existingDriver models.Driver
		if err := s.db.Where("sim_number = ? AND id != ?", *req.SIMNumber, driverID).First(&existingDriver).Error; err == nil {
			return nil, apperrors.NewConflictError("driver with this SIM number already exists")
		}
	}

	// Check for duplicate email if being updated
	if req.Email != nil && *req.Email != driver.Email {
		var existingDriver models.Driver
		if err := s.db.Where("email = ? AND id != ?", *req.Email, driverID).First(&existingDriver).Error; err == nil {
			return nil, apperrors.NewConflictError("driver with this email already exists")
		}
	}

	// Update fields
	if req.FirstName != nil {
		driver.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		driver.LastName = *req.LastName
	}
	if req.PhoneNumber != nil {
		driver.Phone = *req.PhoneNumber
	}
	if req.Email != nil {
		driver.Email = *req.Email
	}
	if req.Address != nil {
		driver.Address = *req.Address
	}
	if req.City != nil {
		driver.City = *req.City
	}
	if req.Province != nil {
		driver.Province = *req.Province
	}
	if req.DateOfBirth != nil {
		driver.DateOfBirth = req.DateOfBirth
	}
	if req.HireDate != nil {
		driver.HireDate = req.HireDate
	}
	if req.Status != nil {
		driver.Status = *req.Status
	}
	if req.EmploymentStatus != nil {
		driver.EmploymentStatus = *req.EmploymentStatus
	}
	if req.IsActive != nil {
		driver.IsActive = *req.IsActive
	}
	if req.NIK != nil {
		driver.NIK = *req.NIK
	}
	if req.SIMNumber != nil {
		driver.SIMNumber = *req.SIMNumber
	}
	if req.SIMExpiry != nil {
		driver.SIMExpiry = req.SIMExpiry
	}
	if req.MedicalCheckupDate != nil {
		driver.MedicalCheckupExpiry = req.MedicalCheckupDate
	}
	if req.TrainingCompleted != nil {
		driver.TrainingCompleted = *req.TrainingCompleted
	}
	if req.TrainingExpiry != nil {
		driver.NextTrainingDate = req.TrainingExpiry
	}
	if req.PerformanceScore != nil {
		driver.PerformanceScore = *req.PerformanceScore
	}
	if req.SafetyScore != nil {
		driver.SafetyScore = *req.SafetyScore
	}
	if req.EfficiencyScore != nil {
		driver.EfficiencyScore = *req.EfficiencyScore
	}

	// Recalculate overall score if any performance score changed
	if req.PerformanceScore != nil || req.SafetyScore != nil || req.EfficiencyScore != nil {
		driver.OverallScore = (driver.PerformanceScore + driver.SafetyScore + driver.EfficiencyScore) / 3
	}

	// Save changes
	if err := s.db.Save(driver).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to update driver")
	}

	return driver, nil
}

// DeleteDriver deletes a driver (soft delete)
func (s *Service) DeleteDriver(companyID, driverID string) error {
	// Check if driver exists
	driver, err := s.GetDriver(companyID, driverID)
	if err != nil {
		return err
	}

	// Check if driver is assigned to a vehicle
	if driver.VehicleID != nil {
		return apperrors.NewBadRequestError("cannot delete driver that is assigned to a vehicle")
	}

	// Soft delete
	if err := s.db.Delete(driver).Error; err != nil {
		return apperrors.Wrap(err, "failed to delete driver")
	}

	return nil
}

// ListDrivers lists drivers with filters and pagination
func (s *Service) ListDrivers(companyID string, filters DriverFilters) ([]models.Driver, int64, error) {
	var drivers []models.Driver
	var total int64

	// Build query
	query := s.db.Model(&models.Driver{}).Where("company_id = ?", companyID)

	// Apply filters
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.EmploymentStatus != nil {
		query = query.Where("employment_status = ?", *filters.EmploymentStatus)
	}
	if filters.PerformanceGrade != nil {
		// Convert grade to score range
		scoreRange := s.getScoreRangeForGrade(PerformanceGrade(*filters.PerformanceGrade))
		query = query.Where("overall_score >= ? AND overall_score <= ?", scoreRange.Min, scoreRange.Max)
	}
	if filters.City != nil {
		query = query.Where("city ILIKE ?", "%"+*filters.City+"%")
	}
	if filters.Province != nil {
		query = query.Where("province ILIKE ?", "%"+*filters.Province+"%")
	}
	if filters.HasVehicle != nil {
		if *filters.HasVehicle {
			query = query.Where("vehicle_id IS NOT NULL")
		} else {
			query = query.Where("vehicle_id IS NULL")
		}
	}
	if filters.IsAvailable != nil {
		if *filters.IsAvailable {
			query = query.Where("status = ? AND is_active = ?", string(StatusAvailable), true)
		} else {
			query = query.Where("status != ? OR is_active = ?", string(StatusAvailable), false)
		}
	}
	if filters.IsCompliant != nil {
		if *filters.IsCompliant {
			// Check compliance: valid SIM, recent medical checkup, completed training
			query = query.Where("sim_expiry > ? AND medical_checkup_expiry > ? AND training_completed = ?", 
				time.Now(), time.Now().AddDate(-1, 0, 0), true)
		} else {
			query = query.Where("sim_expiry <= ? OR medical_checkup_expiry <= ? OR training_completed = ?", 
				time.Now(), time.Now().AddDate(-1, 0, 0), false)
		}
	}
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + *filters.Search + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR nik ILIKE ? OR sim_number ILIKE ?", 
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to count drivers")
	}

	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := strings.ToUpper(filters.SortOrder)
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Execute query with preload
	if err := query.Preload("Vehicle").Find(&drivers).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list drivers")
	}

	return drivers, total, nil
}

// UpdateDriverStatus updates the status of a driver
func (s *Service) UpdateDriverStatus(companyID, driverID string, status DriverStatus, reason string) error {
	driver, err := s.GetDriver(companyID, driverID)
	if err != nil {
		return err
	}

	driver.Status = string(status)
	
	if err := s.db.Save(driver).Error; err != nil {
		return apperrors.Wrap(err, "failed to update driver status")
	}

	// TODO: Add status change history tracking
	// TODO: Add status change notifications

	return nil
}

// UpdateDriverPerformance updates the performance scores of a driver
func (s *Service) UpdateDriverPerformance(companyID, driverID string, performance, safety, efficiency float64) error {
	driver, err := s.GetDriver(companyID, driverID)
	if err != nil {
		return err
	}

	driver.PerformanceScore = performance
	driver.SafetyScore = safety
	driver.EfficiencyScore = efficiency
	driver.OverallScore = (performance + safety + efficiency) / 3

	if err := s.db.Save(driver).Error; err != nil {
		return apperrors.Wrap(err, "failed to update driver performance")
	}

	return nil
}

// GetDriverPerformance gets the performance data of a driver
func (s *Service) GetDriverPerformance(companyID, driverID string) (*models.Driver, error) {
	driver, err := s.GetDriver(companyID, driverID)
	if err != nil {
		return nil, err
	}

	// TODO: Add performance history and trends
	// TODO: Add performance analytics

	return driver, nil
}

// AssignVehicle assigns a driver to a vehicle
func (s *Service) AssignVehicle(companyID, driverID, vehicleID string) error {
	// Validate driver can be assigned
	if err := s.validateDriverAssignment(companyID, driverID); err != nil {
		return err
	}

	// Check if vehicle exists and is available
	var vehicle models.Vehicle
	if err := s.db.Where("company_id = ? AND id = ?", companyID, vehicleID).First(&vehicle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("vehicle")
		}
		return apperrors.Wrap(err, "failed to validate vehicle")
	}

	// Check if vehicle is already assigned
	if vehicle.DriverID != nil {
		return apperrors.NewConflictError("vehicle is already assigned to another driver")
	}

	// Update driver
	if err := s.db.Model(&models.Driver{}).Where("company_id = ? AND id = ?", companyID, driverID).Update("vehicle_id", vehicleID).Error; err != nil {
		return apperrors.Wrap(err, "failed to assign vehicle to driver")
	}

	// Update vehicle
	if err := s.db.Model(&models.Vehicle{}).Where("company_id = ? AND id = ?", companyID, vehicleID).Update("driver_id", driverID).Error; err != nil {
		return apperrors.Wrap(err, "failed to assign driver to vehicle")
	}

	// TODO: Add assignment history tracking
	// TODO: Add assignment notifications

	return nil
}

// UnassignVehicle removes vehicle assignment from a driver
func (s *Service) UnassignVehicle(companyID, driverID string) error {
	// Update driver
	if err := s.db.Model(&models.Driver{}).Where("company_id = ? AND id = ?", companyID, driverID).Update("vehicle_id", nil).Error; err != nil {
		return apperrors.Wrap(err, "failed to unassign vehicle from driver")
	}

	// Update vehicle
	if err := s.db.Model(&models.Vehicle{}).Where("company_id = ? AND driver_id = ?", companyID, driverID).Update("driver_id", nil).Error; err != nil {
		return apperrors.Wrap(err, "failed to unassign driver from vehicle")
	}

	// TODO: Add unassignment history tracking
	// TODO: Add unassignment notifications

	return nil
}

// GetDriverVehicle gets the vehicle assigned to a driver
func (s *Service) GetDriverVehicle(companyID, driverID string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	
	if err := s.db.Joins("JOIN drivers ON vehicles.driver_id = drivers.id").
		Where("drivers.company_id = ? AND drivers.id = ?", companyID, driverID).
		First(&vehicle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("vehicle assigned to this driver")
		}
		return nil, apperrors.Wrap(err, "failed to get driver vehicle")
	}

	return &vehicle, nil
}

// UpdateMedicalCheckup updates the medical checkup date of a driver
func (s *Service) UpdateMedicalCheckup(companyID, driverID string, checkupDate time.Time) error {
	driver, err := s.GetDriver(companyID, driverID)
	if err != nil {
		return err
	}

	driver.MedicalCheckupExpiry = &checkupDate

	if err := s.db.Save(driver).Error; err != nil {
		return apperrors.Wrap(err, "failed to update medical checkup")
	}

	return nil
}

// UpdateTrainingStatus updates the training status of a driver
func (s *Service) UpdateTrainingStatus(companyID, driverID string, completed bool, expiryDate *time.Time) error {
	driver, err := s.GetDriver(companyID, driverID)
	if err != nil {
		return err
	}

	driver.TrainingCompleted = completed
	driver.NextTrainingDate = expiryDate

	if err := s.db.Save(driver).Error; err != nil {
		return apperrors.Wrap(err, "failed to update training status")
	}

	return nil
}

// ScoreRange represents a score range for performance grades
type ScoreRange struct {
	Min float64
	Max float64
}

// getScoreRangeForGrade returns the score range for a performance grade
func (s *Service) getScoreRangeForGrade(grade PerformanceGrade) ScoreRange {
	switch grade {
	case GradeA:
		return ScoreRange{Min: 90, Max: 100}
	case GradeB:
		return ScoreRange{Min: 80, Max: 89}
	case GradeC:
		return ScoreRange{Min: 70, Max: 79}
	case GradeD:
		return ScoreRange{Min: 60, Max: 69}
	case GradeF:
		return ScoreRange{Min: 0, Max: 59}
	default:
		return ScoreRange{Min: 0, Max: 100}
	}
}

// validateIndonesianCompliance validates Indonesian compliance fields
func (s *Service) validateIndonesianCompliance(nik, simNumber string, dateOfBirth time.Time) error {
	// Validate NIK format
	if err := s.validateNIK(nik); err != nil {
		return fmt.Errorf("NIK validation failed: %w", err)
	}

	// Validate SIM number format
	if err := s.validateSIMNumber(simNumber); err != nil {
		return fmt.Errorf("SIM validation failed: %w", err)
	}

	// Validate age (must be at least 18)
	age := time.Now().Year() - dateOfBirth.Year()
	if age < 18 {
		return apperrors.NewValidationError("driver must be at least 18 years old")
	}

	return nil
}

// validateNIK validates Indonesian NIK format
func (s *Service) validateNIK(nik string) error {
	// Indonesian NIK format: 16 digits
	pattern := `^[0-9]{16}$`
	matched, err := regexp.MatchString(pattern, nik)
	if err != nil {
		return fmt.Errorf("failed to validate NIK: %w", err)
	}
	if !matched {
		return apperrors.NewValidationError("invalid NIK format, expected 16 digits")
	}
	return nil
}

// validateSIMNumber validates Indonesian SIM number format
func (s *Service) validateSIMNumber(simNumber string) error {
	// Indonesian SIM format: alphanumeric, 10-20 characters
	pattern := `^[A-Z0-9]{10,20}$`
	matched, err := regexp.MatchString(pattern, simNumber)
	if err != nil {
		return fmt.Errorf("failed to validate SIM number: %w", err)
	}
	if !matched {
		return apperrors.NewValidationError("invalid Indonesian SIM number format, expected 10-20 alphanumeric characters")
	}
	return nil
}

// validateDriverAssignment validates if a driver can be assigned to a vehicle
func (s *Service) validateDriverAssignment(companyID, driverID string) error {
	var driver models.Driver
	
	if err := s.db.Where("company_id = ? AND id = ? AND is_active = ?", companyID, driverID, true).First(&driver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("driver or driver is inactive")
		}
		return apperrors.Wrap(err, "failed to validate driver")
	}

	// Check if driver can drive (includes license validation)
	if !driver.CanDrive() {
		return apperrors.NewBadRequestError("driver license is expired or invalid, or driver is not available")
	}

	// Check if driver is already assigned to another vehicle
	if driver.VehicleID != nil {
		return apperrors.NewConflictError("driver is already assigned to another vehicle")
	}

	return nil
}