package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// VehicleRepositoryImpl implements the VehicleRepository interface
type VehicleRepositoryImpl struct {
	*BaseRepository[models.Vehicle]
}

// NewVehicleRepository creates a new vehicle repository
func NewVehicleRepository(db *gorm.DB) VehicleRepository {
	return &VehicleRepositoryImpl{
		BaseRepository: NewBaseRepository[models.Vehicle](db),
	}
}

// GetByCompany retrieves vehicles by company ID with pagination
func (r *VehicleRepositoryImpl) GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by company: %w", err)
	}
	
	return vehicles, nil
}

// GetByDriver retrieves the vehicle assigned to a specific driver
func (r *VehicleRepositoryImpl) GetByDriver(ctx context.Context, driverID string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	if err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).First(&vehicle).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no vehicle assigned to driver: %s", driverID)
		}
		return nil, fmt.Errorf("failed to get vehicle by driver: %w", err)
	}
	return &vehicle, nil
}

// GetByStatus retrieves vehicles by status within a company
func (r *VehicleRepositoryImpl) GetByStatus(ctx context.Context, companyID string, status string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	if err := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, status).Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by status: %w", err)
	}
	return vehicles, nil
}

// GetByType retrieves vehicles by type within a company
func (r *VehicleRepositoryImpl) GetByType(ctx context.Context, companyID string, vehicleType string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	if err := r.db.WithContext(ctx).Where("company_id = ? AND vehicle_type = ?", companyID, vehicleType).Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by type: %w", err)
	}
	return vehicles, nil
}

// SearchByLicensePlate searches vehicles by license plate within a company
func (r *VehicleRepositoryImpl) SearchByLicensePlate(ctx context.Context, licensePlate string, companyID string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	searchPattern := "%" + strings.ToUpper(licensePlate) + "%"
	
	if err := r.db.WithContext(ctx).Where("company_id = ? AND UPPER(license_plate) LIKE ?", companyID, searchPattern).Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to search vehicles by license plate: %w", err)
	}
	return vehicles, nil
}

// SearchByVIN searches vehicles by VIN within a company
func (r *VehicleRepositoryImpl) SearchByVIN(ctx context.Context, vin string, companyID string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	searchPattern := "%" + strings.ToUpper(vin) + "%"
	
	if err := r.db.WithContext(ctx).Where("company_id = ? AND UPPER(vin) LIKE ?", companyID, searchPattern).Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to search vehicles by VIN: %w", err)
	}
	return vehicles, nil
}

// GetAvailableVehicles retrieves vehicles that are available for assignment
func (r *VehicleRepositoryImpl) GetAvailableVehicles(ctx context.Context, companyID string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	if err := r.db.WithContext(ctx).Where("company_id = ? AND status = ? AND driver_id IS NULL", companyID, "active").Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get available vehicles: %w", err)
	}
	return vehicles, nil
}

// GetVehiclesNeedingInspection retrieves vehicles that need inspection
func (r *VehicleRepositoryImpl) GetVehiclesNeedingInspection(ctx context.Context, companyID string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	query := `
		SELECT * FROM vehicles 
		WHERE company_id = ? 
		AND (next_service_date IS NULL 
			OR next_service_date <= CURRENT_DATE + INTERVAL '30 days')
		AND status = 'active'
	`
	
	if err := r.db.WithContext(ctx).Raw(query, companyID).Scan(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles needing inspection: %w", err)
	}
	
	return vehicles, nil
}

// UpdateStatus updates the status of a vehicle
func (r *VehicleRepositoryImpl) UpdateStatus(ctx context.Context, vehicleID string, status string) error {
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("id = ?", vehicleID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update vehicle status: %w", err)
	}
	return nil
}

// AssignDriver assigns a driver to a vehicle
func (r *VehicleRepositoryImpl) AssignDriver(ctx context.Context, vehicleID string, driverID *string) error {
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("id = ?", vehicleID).Update("driver_id", driverID).Error; err != nil {
		return fmt.Errorf("failed to assign driver to vehicle: %w", err)
	}
	return nil
}

// UpdateOdometer updates the odometer reading of a vehicle
func (r *VehicleRepositoryImpl) UpdateOdometer(ctx context.Context, vehicleID string, odometer float64) error {
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("id = ?", vehicleID).Update("odometer_reading", odometer).Error; err != nil {
		return fmt.Errorf("failed to update vehicle odometer: %w", err)
	}
	return nil
}

// GetVehiclesByFuelType retrieves vehicles by fuel type within a company
func (r *VehicleRepositoryImpl) GetVehiclesByFuelType(ctx context.Context, companyID string, fuelType string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	if err := r.db.WithContext(ctx).Where("company_id = ? AND fuel_type = ?", companyID, fuelType).Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by fuel type: %w", err)
	}
	return vehicles, nil
}

// GetVehiclesByYear retrieves vehicles by year range within a company
func (r *VehicleRepositoryImpl) GetVehiclesByYear(ctx context.Context, companyID string, startYear, endYear int) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	if startYear > 0 {
		query = query.Where("year >= ?", startYear)
	}
	if endYear > 0 {
		query = query.Where("year <= ?", endYear)
	}
	
	if err := query.Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by year: %w", err)
	}
	
	return vehicles, nil
}

// GetVehiclesByMakeAndModel retrieves vehicles by make and model within a company
func (r *VehicleRepositoryImpl) GetVehiclesByMakeAndModel(ctx context.Context, companyID string, make, model string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	if make != "" {
		query = query.Where("LOWER(make) = ?", strings.ToLower(make))
	}
	if model != "" {
		query = query.Where("LOWER(model) = ?", strings.ToLower(model))
	}
	
	if err := query.Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by make and model: %w", err)
	}
	
	return vehicles, nil
}

// GetVehiclesWithLowFuel retrieves vehicles with low fuel levels
func (r *VehicleRepositoryImpl) GetVehiclesWithLowFuel(ctx context.Context, companyID string, fuelThreshold float64) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	// Assuming we have a fuel_level field in the vehicle model
	if err := r.db.WithContext(ctx).Where("company_id = ? AND fuel_level <= ?", companyID, fuelThreshold).Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles with low fuel: %w", err)
	}
	return vehicles, nil
}

// GetVehicleStatistics retrieves vehicle statistics for a company
func (r *VehicleRepositoryImpl) GetVehicleStatistics(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats struct {
		TotalVehicles     int64 `json:"total_vehicles"`
		ActiveVehicles    int64 `json:"active_vehicles"`
		MaintenanceVehicles int64 `json:"maintenance_vehicles"`
		RetiredVehicles   int64 `json:"retired_vehicles"`
		AssignedVehicles  int64 `json:"assigned_vehicles"`
		UnassignedVehicles int64 `json:"unassigned_vehicles"`
		GPSEnabledVehicles int64 `json:"gps_enabled_vehicles"`
	}

	// Get total vehicles
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("company_id = ?", companyID).Count(&stats.TotalVehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to count total vehicles: %w", err)
	}

	// Get active vehicles
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("company_id = ? AND status = ?", companyID, "active").Count(&stats.ActiveVehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to count active vehicles: %w", err)
	}

	// Get maintenance vehicles
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("company_id = ? AND status = ?", companyID, "maintenance").Count(&stats.MaintenanceVehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to count maintenance vehicles: %w", err)
	}

	// Get retired vehicles
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("company_id = ? AND status = ?", companyID, "retired").Count(&stats.RetiredVehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to count retired vehicles: %w", err)
	}

	// Get assigned vehicles
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("company_id = ? AND driver_id IS NOT NULL", companyID).Count(&stats.AssignedVehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to count assigned vehicles: %w", err)
	}

	// Get unassigned vehicles
	stats.UnassignedVehicles = stats.TotalVehicles - stats.AssignedVehicles

	// Get GPS enabled vehicles
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("company_id = ? AND is_gps_enabled = true", companyID).Count(&stats.GPSEnabledVehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to count GPS enabled vehicles: %w", err)
	}

	return map[string]interface{}{
		"total_vehicles":        stats.TotalVehicles,
		"active_vehicles":       stats.ActiveVehicles,
		"maintenance_vehicles":  stats.MaintenanceVehicles,
		"retired_vehicles":      stats.RetiredVehicles,
		"assigned_vehicles":     stats.AssignedVehicles,
		"unassigned_vehicles":   stats.UnassignedVehicles,
		"gps_enabled_vehicles":  stats.GPSEnabledVehicles,
	}, nil
}

// GetVehicleMaintenanceSchedule retrieves vehicles with upcoming maintenance
func (r *VehicleRepositoryImpl) GetVehicleMaintenanceSchedule(ctx context.Context, companyID string, days int) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	query := `
		SELECT * FROM vehicles 
		WHERE company_id = ? 
		AND (next_service_date IS NOT NULL 
			AND next_service_date <= CURRENT_DATE + INTERVAL '%d days')
		ORDER BY next_service_date ASC
	`
	
	if err := r.db.WithContext(ctx).Raw(fmt.Sprintf(query, days), companyID).Scan(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicle maintenance schedule: %w", err)
	}
	
	return vehicles, nil
}

// BulkUpdateStatus updates the status of multiple vehicles
func (r *VehicleRepositoryImpl) BulkUpdateStatus(ctx context.Context, vehicleIDs []string, status string) error {
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("id IN ?", vehicleIDs).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to bulk update vehicle status: %w", err)
	}
	return nil
}

// GetVehiclesByInsuranceExpiry retrieves vehicles with expiring insurance
func (r *VehicleRepositoryImpl) GetVehiclesByInsuranceExpiry(ctx context.Context, companyID string, days int) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	query := `
		SELECT * FROM vehicles 
		WHERE company_id = ? 
		AND (insurance_expiry_date IS NOT NULL 
			AND insurance_expiry_date <= CURRENT_DATE + INTERVAL '%d days')
		ORDER BY insurance_expiry_date ASC
	`
	
	if err := r.db.WithContext(ctx).Raw(fmt.Sprintf(query, days), companyID).Scan(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by insurance expiry: %w", err)
	}
	
	return vehicles, nil
}

// SearchVehicles performs a comprehensive search across multiple vehicle fields
func (r *VehicleRepositoryImpl) SearchVehicles(ctx context.Context, companyID string, searchQuery string, pagination Pagination) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	query = query.Where(
		"LOWER(license_plate) LIKE ? OR LOWER(make) LIKE ? OR LOWER(model) LIKE ? OR LOWER(vin) LIKE ? OR LOWER(color) LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
	)
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to search vehicles: %w", err)
	}
	
	return vehicles, nil
}
