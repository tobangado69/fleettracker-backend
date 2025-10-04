package vehicle

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

// Service handles vehicle operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new vehicle service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// VehicleStatus represents the status of a vehicle
type VehicleStatus string

const (
	StatusActive      VehicleStatus = "active"
	StatusMaintenance VehicleStatus = "maintenance"
	StatusRetired     VehicleStatus = "retired"
	StatusInactive    VehicleStatus = "inactive"
)

// CreateVehicleRequest represents the request to create a vehicle
type CreateVehicleRequest struct {
	Make                    string     `json:"make" validate:"required,min=2,max=100"`
	Model                   string     `json:"model" validate:"required,min=2,max=100"`
	Year                    int        `json:"year" validate:"required,min=1900,max=2030"`
	LicensePlate            string     `json:"license_plate" validate:"required,min=5,max=20"`
	VIN                     string     `json:"vin" validate:"required,len=17"`
	Color                   string     `json:"color" validate:"required,min=2,max=50"`
	FuelType                string     `json:"fuel_type" validate:"required,oneof=gasoline diesel electric hybrid"`
	CurrentOdometer         int        `json:"current_odometer" validate:"min=0"`
	PurchaseDate            *time.Time `json:"purchase_date"`
	DriverID                *string    `json:"driver_id,omitempty"`
	
	// Indonesian Compliance Fields
	STNKNumber              string     `json:"stnk_number" validate:"required,min=10,max=20"`
	BPKBNumber              string     `json:"bpkb_number" validate:"required,min=10,max=20"`
	InsurancePolicyNumber   string     `json:"insurance_policy_number" validate:"required,min=5,max=50"`
	LastInspectionDate      *time.Time `json:"last_inspection_date"`
}

// UpdateVehicleRequest represents the request to update a vehicle
type UpdateVehicleRequest struct {
	Make                    *string    `json:"make,omitempty" validate:"omitempty,min=2,max=100"`
	Model                   *string    `json:"model,omitempty" validate:"omitempty,min=2,max=100"`
	Year                    *int       `json:"year,omitempty" validate:"omitempty,min=1900,max=2030"`
	LicensePlate            *string    `json:"license_plate,omitempty" validate:"omitempty,min=5,max=20"`
	VIN                     *string    `json:"vin,omitempty" validate:"omitempty,len=17"`
	Color                   *string    `json:"color,omitempty" validate:"omitempty,min=2,max=50"`
	FuelType                *string    `json:"fuel_type,omitempty" validate:"omitempty,oneof=gasoline diesel electric hybrid"`
	CurrentOdometer         *int       `json:"current_odometer,omitempty" validate:"omitempty,min=0"`
	PurchaseDate            *time.Time `json:"purchase_date,omitempty"`
	DriverID                *string    `json:"driver_id,omitempty"`
	Status                  *string    `json:"status,omitempty" validate:"omitempty,oneof=active maintenance retired inactive"`
	IsActive                *bool      `json:"is_active,omitempty"`
	IsGPSEnabled            *bool      `json:"is_gps_enabled,omitempty"`
	
	// Indonesian Compliance Fields
	STNKNumber              *string    `json:"stnk_number,omitempty" validate:"omitempty,min=10,max=20"`
	BPKBNumber              *string    `json:"bpkb_number,omitempty" validate:"omitempty,min=10,max=20"`
	InsurancePolicyNumber   *string    `json:"insurance_policy_number,omitempty" validate:"omitempty,min=5,max=50"`
	LastInspectionDate      *time.Time `json:"last_inspection_date,omitempty"`
}

// VehicleFilters represents filters for listing vehicles
type VehicleFilters struct {
	Status      *string `json:"status" form:"status"`
	Make        *string `json:"make" form:"make"`
	Model       *string `json:"model" form:"model"`
	Year        *int    `json:"year" form:"year"`
	FuelType    *string `json:"fuel_type" form:"fuel_type"`
	HasDriver   *bool   `json:"has_driver" form:"has_driver"`
	GPSEnabled  *bool   `json:"gps_enabled" form:"gps_enabled"`
	Search      *string `json:"search" form:"search"`
	
	// Pagination
	Page        int     `json:"page" form:"page" validate:"min=1"`
	Limit       int     `json:"limit" form:"limit" validate:"min=1,max=100"`
	SortBy      string  `json:"sort_by" form:"sort_by" validate:"oneof=created_at updated_at make model year license_plate"`
	SortOrder   string  `json:"sort_order" form:"sort_order" validate:"oneof=asc desc"`
}

// VehicleResponse represents the response for vehicle data
type VehicleResponse struct {
	ID                      string     `json:"id"`
	CompanyID               string     `json:"company_id"`
	DriverID                *string    `json:"driver_id"`
	Make                    string     `json:"make"`
	Model                   string     `json:"model"`
	Year                    int        `json:"year"`
	LicensePlate            string     `json:"license_plate"`
	VIN                     string     `json:"vin"`
	Color                   string     `json:"color"`
	FuelType                string     `json:"fuel_type"`
	CurrentOdometer         int        `json:"current_odometer"`
	LastMaintenanceOdometer int        `json:"last_maintenance_odometer"`
	PurchaseDate            *time.Time `json:"purchase_date"`
	Status                  string     `json:"status"`
	IsActive                bool       `json:"is_active"`
	IsGPSEnabled            bool       `json:"is_gps_enabled"`
	
	// Indonesian Compliance Fields
	STNKNumber              string     `json:"stnk_number"`
	BPKBNumber              string     `json:"bpkb_number"`
	InsurancePolicyNumber   string     `json:"insurance_policy_number"`
	LastInspectionDate      *time.Time `json:"last_inspection_date"`
	NextInspectionDate      *time.Time `json:"next_inspection_date"`
	
	// Relationships
	Driver                  *models.Driver `json:"driver,omitempty"`
	
	// Timestamps
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// CreateVehicle creates a new vehicle
func (s *Service) CreateVehicle(companyID string, req CreateVehicleRequest) (*models.Vehicle, error) {
	// Validate Indonesian compliance fields
	if err := s.validateIndonesianCompliance(req.STNKNumber, req.BPKBNumber, req.LicensePlate); err != nil {
		return nil, err
	}

	// Check if license plate already exists
	var existingVehicle models.Vehicle
	if err := s.db.Where("license_plate = ?", req.LicensePlate).First(&existingVehicle).Error; err == nil {
		return nil, apperrors.NewConflictError("Vehicle with this license plate already exists")
	}

	// Check if VIN already exists
	if err := s.db.Where("vin = ?", req.VIN).First(&existingVehicle).Error; err == nil {
		return nil, apperrors.NewConflictError("Vehicle with this VIN already exists")
	}

	// Check if STNK number already exists
	if err := s.db.Where("stnk = ?", req.STNKNumber).First(&existingVehicle).Error; err == nil {
		return nil, apperrors.NewConflictError("Vehicle with this STNK number already exists")
	}

	// Validate driver assignment if provided
	if req.DriverID != nil {
		if err := s.validateDriverAssignment(companyID, *req.DriverID); err != nil {
			return nil, err
		}
	}

	// Create vehicle
	vehicle := &models.Vehicle{
		CompanyID:               companyID,
		DriverID:                req.DriverID,
		Make:                    req.Make,
		Model:                   req.Model,
		Year:                    req.Year,
		LicensePlate:            req.LicensePlate,
		VIN:                     req.VIN,
		Color:                   req.Color,
		FuelType:                req.FuelType,
		OdometerReading:         float64(req.CurrentOdometer),
		Status:                  string(StatusActive),
		IsActive:                true,
		IsGPSEnabled:            true,
		STNK:                    req.STNKNumber,
		BPKB:                    req.BPKBNumber,
		InsuranceNumber:         req.InsurancePolicyNumber,
		LastServiceDate:         req.LastInspectionDate,
	}

	// Calculate next inspection date if last inspection date is provided
	if req.LastInspectionDate != nil {
		nextInspection := req.LastInspectionDate.AddDate(1, 0, 0) // 1 year later
		vehicle.NextServiceDate = &nextInspection
	}

	// Save to database
	if err := s.db.Create(vehicle).Error; err != nil {
		return nil, apperrors.NewInternalError("Failed to create vehicle").WithInternal(err)
	}

	return vehicle, nil
}

// GetVehicle retrieves a vehicle by ID
func (s *Service) GetVehicle(companyID, vehicleID string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	
	if err := s.db.Preload("Driver").Where("company_id = ? AND id = ?", companyID, vehicleID).First(&vehicle).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError("Vehicle")
		}
		return nil, apperrors.NewInternalError("Failed to fetch vehicle").WithInternal(err)
	}

	return &vehicle, nil
}

// UpdateVehicle updates a vehicle
func (s *Service) UpdateVehicle(companyID, vehicleID string, req UpdateVehicleRequest) (*models.Vehicle, error) {
	// Get existing vehicle
	vehicle, err := s.GetVehicle(companyID, vehicleID)
	if err != nil {
		return nil, err
	}

	// Validate Indonesian compliance fields if provided
	if req.STNKNumber != nil || req.BPKBNumber != nil || req.LicensePlate != nil {
		stnk := vehicle.STNK
		bpkb := vehicle.BPKB
		plate := vehicle.LicensePlate
		
		if req.STNKNumber != nil {
			stnk = *req.STNKNumber
		}
		if req.BPKBNumber != nil {
			bpkb = *req.BPKBNumber
		}
		if req.LicensePlate != nil {
			plate = *req.LicensePlate
		}
		
		if err := s.validateIndonesianCompliance(stnk, bpkb, plate); err != nil {
			return nil, err
		}
	}

	// Check for duplicate license plate if being updated
	if req.LicensePlate != nil && *req.LicensePlate != vehicle.LicensePlate {
		var existingVehicle models.Vehicle
		if err := s.db.Where("license_plate = ? AND id != ?", *req.LicensePlate, vehicleID).First(&existingVehicle).Error; err == nil {
			return nil, apperrors.NewConflictError("Vehicle with this license plate already exists")
		}
	}

	// Check for duplicate VIN if being updated
	if req.VIN != nil && *req.VIN != vehicle.VIN {
		var existingVehicle models.Vehicle
		if err := s.db.Where("vin = ? AND id != ?", *req.VIN, vehicleID).First(&existingVehicle).Error; err == nil {
			return nil, apperrors.NewConflictError("Vehicle with this VIN already exists")
		}
	}

	// Check for duplicate STNK if being updated
	if req.STNKNumber != nil && *req.STNKNumber != vehicle.STNK {
		var existingVehicle models.Vehicle
		if err := s.db.Where("stnk = ? AND id != ?", *req.STNKNumber, vehicleID).First(&existingVehicle).Error; err == nil {
			return nil, apperrors.NewConflictError("Vehicle with this STNK number already exists")
		}
	}

	// Validate driver assignment if provided
	if req.DriverID != nil {
		if err := s.validateDriverAssignment(companyID, *req.DriverID); err != nil {
			return nil, err
		}
	}

	// Update fields
	if req.Make != nil {
		vehicle.Make = *req.Make
	}
	if req.Model != nil {
		vehicle.Model = *req.Model
	}
	if req.Year != nil {
		vehicle.Year = *req.Year
	}
	if req.LicensePlate != nil {
		vehicle.LicensePlate = *req.LicensePlate
	}
	if req.VIN != nil {
		vehicle.VIN = *req.VIN
	}
	if req.Color != nil {
		vehicle.Color = *req.Color
	}
	if req.FuelType != nil {
		vehicle.FuelType = *req.FuelType
	}
	if req.CurrentOdometer != nil {
		vehicle.OdometerReading = float64(*req.CurrentOdometer)
	}
	if req.PurchaseDate != nil {
		// Note: Vehicle model doesn't have PurchaseDate field, using LastServiceDate for now
		vehicle.LastServiceDate = req.PurchaseDate
	}
	if req.DriverID != nil {
		vehicle.DriverID = req.DriverID
	}
	if req.Status != nil {
		vehicle.Status = *req.Status
	}
	if req.IsActive != nil {
		vehicle.IsActive = *req.IsActive
	}
	if req.IsGPSEnabled != nil {
		vehicle.IsGPSEnabled = *req.IsGPSEnabled
	}
	if req.STNKNumber != nil {
		vehicle.STNK = *req.STNKNumber
	}
	if req.BPKBNumber != nil {
		vehicle.BPKB = *req.BPKBNumber
	}
	if req.InsurancePolicyNumber != nil {
		vehicle.InsuranceNumber = *req.InsurancePolicyNumber
	}
	if req.LastInspectionDate != nil {
		vehicle.LastServiceDate = req.LastInspectionDate
		// Calculate next inspection date
		nextInspection := req.LastInspectionDate.AddDate(1, 0, 0) // 1 year later
		vehicle.NextServiceDate = &nextInspection
	}

	// Save changes
	if err := s.db.Save(vehicle).Error; err != nil {
		return nil, apperrors.NewInternalError("Failed to update vehicle").WithInternal(err)
	}

	return vehicle, nil
}

// DeleteVehicle deletes a vehicle (soft delete)
func (s *Service) DeleteVehicle(companyID, vehicleID string) error {
	// Check if vehicle exists
	vehicle, err := s.GetVehicle(companyID, vehicleID)
	if err != nil {
		return err
	}

	// Check if vehicle is assigned to a driver
	if vehicle.DriverID != nil {
		return apperrors.NewBadRequestError("Cannot delete vehicle that is assigned to a driver")
	}

	// Soft delete
	if err := s.db.Delete(vehicle).Error; err != nil {
		return apperrors.NewInternalError("Failed to delete vehicle").WithInternal(err)
	}

	return nil
}

// ListVehicles lists vehicles with filters and pagination
func (s *Service) ListVehicles(companyID string, filters VehicleFilters) ([]models.Vehicle, int64, error) {
	var vehicles []models.Vehicle
	var total int64

	// Build query
	query := s.db.Model(&models.Vehicle{}).Where("company_id = ?", companyID)

	// Apply filters
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.Make != nil {
		query = query.Where("make ILIKE ?", "%"+*filters.Make+"%")
	}
	if filters.Model != nil {
		query = query.Where("model ILIKE ?", "%"+*filters.Model+"%")
	}
	if filters.Year != nil {
		query = query.Where("year = ?", *filters.Year)
	}
	if filters.FuelType != nil {
		query = query.Where("fuel_type = ?", *filters.FuelType)
	}
	if filters.HasDriver != nil {
		if *filters.HasDriver {
			query = query.Where("driver_id IS NOT NULL")
		} else {
			query = query.Where("driver_id IS NULL")
		}
	}
	if filters.GPSEnabled != nil {
		query = query.Where("is_gps_enabled = ?", *filters.GPSEnabled)
	}
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + *filters.Search + "%"
		query = query.Where("make ILIKE ? OR model ILIKE ? OR license_plate ILIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.NewInternalError("Failed to count vehicles").WithInternal(err)
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
	if err := query.Preload("Driver").Find(&vehicles).Error; err != nil {
		return nil, 0, apperrors.NewInternalError("Failed to list vehicles").WithInternal(err)
	}

	return vehicles, total, nil
}

// UpdateVehicleStatus updates the status of a vehicle
func (s *Service) UpdateVehicleStatus(companyID, vehicleID string, status VehicleStatus, reason string) error {
	vehicle, err := s.GetVehicle(companyID, vehicleID)
	if err != nil {
		return err
	}

	vehicle.Status = string(status)
	
	if err := s.db.Save(vehicle).Error; err != nil {
		return apperrors.NewInternalError("Failed to update vehicle status").WithInternal(err)
	}

	// TODO: Add status change history tracking
	// TODO: Add status change notifications

	return nil
}

// AssignDriver assigns a driver to a vehicle
func (s *Service) AssignDriver(companyID, vehicleID, driverID string) error {
	// Validate driver assignment
	if err := s.validateDriverAssignment(companyID, driverID); err != nil {
		return err
	}

	// Check if driver is already assigned to another vehicle
	var existingVehicle models.Vehicle
	if err := s.db.Where("company_id = ? AND driver_id = ?", companyID, driverID).First(&existingVehicle).Error; err == nil {
		return apperrors.NewConflictError("Driver is already assigned to another vehicle")
	}

	// Update vehicle
	if err := s.db.Model(&models.Vehicle{}).Where("company_id = ? AND id = ?", companyID, vehicleID).Update("driver_id", driverID).Error; err != nil {
		return apperrors.NewInternalError("Failed to assign driver").WithInternal(err)
	}

	// TODO: Add assignment history tracking
	// TODO: Add assignment notifications

	return nil
}

// UnassignDriver removes driver assignment from a vehicle
func (s *Service) UnassignDriver(companyID, vehicleID string) error {
	// Update vehicle
	if err := s.db.Model(&models.Vehicle{}).Where("company_id = ? AND id = ?", companyID, vehicleID).Update("driver_id", nil).Error; err != nil {
		return apperrors.NewInternalError("Failed to unassign driver").WithInternal(err)
	}

	// TODO: Add unassignment history tracking
	// TODO: Add unassignment notifications

	return nil
}

// GetVehicleDriver gets the driver assigned to a vehicle
func (s *Service) GetVehicleDriver(companyID, vehicleID string) (*models.Driver, error) {
	var driver models.Driver
	
	if err := s.db.Joins("JOIN vehicles ON drivers.id = vehicles.driver_id").
		Where("vehicles.company_id = ? AND vehicles.id = ?", companyID, vehicleID).
		First(&driver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Driver")
		}
		return nil, apperrors.NewInternalError("Failed to get vehicle driver").WithInternal(err)
	}

	return &driver, nil
}

// UpdateInspectionDate updates the inspection date of a vehicle
func (s *Service) UpdateInspectionDate(companyID, vehicleID string, inspectionDate time.Time) error {
	vehicle, err := s.GetVehicle(companyID, vehicleID)
	if err != nil {
		return err
	}

	vehicle.LastServiceDate = &inspectionDate
	// Calculate next inspection date (1 year later)
	nextInspection := inspectionDate.AddDate(1, 0, 0)
	vehicle.NextServiceDate = &nextInspection

	if err := s.db.Save(vehicle).Error; err != nil {
		return apperrors.NewInternalError("Failed to update inspection date").WithInternal(err)
	}

	return nil
}

// validateIndonesianCompliance validates Indonesian compliance fields
func (s *Service) validateIndonesianCompliance(stnkNumber, bpkbNumber, licensePlate string) error {
	// Validate STNK number format
	if err := s.validateSTNKNumber(stnkNumber); err != nil {
		return apperrors.NewValidationError("STNK validation failed").WithInternal(err)
	}

	// Validate license plate format
	if err := s.validateIndonesianLicensePlate(licensePlate); err != nil {
		return apperrors.NewValidationError("License plate validation failed").WithInternal(err)
	}

	// BPKB number validation (basic format check)
	if len(bpkbNumber) < 10 || len(bpkbNumber) > 20 {
		return apperrors.NewValidationError("BPKB number must be between 10 and 20 characters")
	}

	return nil
}

// validateSTNKNumber validates Indonesian STNK number format
func (s *Service) validateSTNKNumber(stnkNumber string) error {
	// Indonesian STNK format: XXXX-XXXX-XXXX-XXXX
	pattern := `^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$`
	matched, err := regexp.MatchString(pattern, stnkNumber)
	if err != nil {
		return apperrors.NewValidationError("Failed to validate STNK number").WithInternal(err)
	}
	if !matched {
		return apperrors.NewValidationError("Invalid STNK number format, expected format: XXXX-XXXX-XXXX-XXXX")
	}
	return nil
}

// validateIndonesianLicensePlate validates Indonesian license plate format
func (s *Service) validateIndonesianLicensePlate(plate string) error {
	// Indonesian license plate format: B 1234 ABC (region number letters)
	pattern := `^[A-Z]{1,2}\s[0-9]{1,4}\s[A-Z]{1,3}$`
	matched, err := regexp.MatchString(pattern, plate)
	if err != nil {
		return apperrors.NewValidationError("Failed to validate license plate").WithInternal(err)
	}
	if !matched {
		return apperrors.NewValidationError("Invalid Indonesian license plate format, expected format: B 1234 ABC")
	}
	return nil
}

// validateDriverAssignment validates if a driver can be assigned
func (s *Service) validateDriverAssignment(companyID, driverID string) error {
	var driver models.Driver
	
	if err := s.db.Where("company_id = ? AND id = ? AND is_active = ?", companyID, driverID, true).First(&driver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("Driver")
		}
		return apperrors.NewInternalError("Failed to validate driver").WithInternal(err)
	}

	// Check if driver can drive (includes license validation)
	if !driver.CanDrive() {
		return apperrors.NewValidationError("Driver license is expired or invalid, or driver is not available")
	}

	return nil
}
