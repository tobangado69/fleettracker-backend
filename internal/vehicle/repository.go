package vehicle

import (
	"context"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// VehicleRepository defines the interface for vehicle data operations
type VehicleRepository interface {
	// CRUD operations
	Create(ctx context.Context, vehicle *models.Vehicle) error
	FindByID(ctx context.Context, id, companyID string) (*models.Vehicle, error)
	FindByCompanyID(ctx context.Context, companyID string, filters VehicleFilters) ([]*models.Vehicle, int64, error)
	Update(ctx context.Context, vehicle *models.Vehicle) error
	Delete(ctx context.Context, id string) error
	
	// Query operations
	FindByStatus(ctx context.Context, companyID string, status string) ([]*models.Vehicle, error)
	FindActive(ctx context.Context, companyID string) ([]*models.Vehicle, error)
	FindByLicensePlate(ctx context.Context, licensePlate string) (*models.Vehicle, error)
	FindByVIN(ctx context.Context, vin string) (*models.Vehicle, error)
	FindByMakeModel(ctx context.Context, companyID, make, model string) ([]*models.Vehicle, error)
	FindByYearRange(ctx context.Context, companyID string, startYear, endYear int) ([]*models.Vehicle, error)
	FindByFuelType(ctx context.Context, companyID string, fuelType string) ([]*models.Vehicle, error)
	FindWithDriver(ctx context.Context, companyID string) ([]*models.Vehicle, error)
	FindWithoutDriver(ctx context.Context, companyID string) ([]*models.Vehicle, error)
	FindNeedingInspection(ctx context.Context, companyID string, daysBefore int) ([]*models.Vehicle, error)
	
	// Bulk operations
	BatchUpdateStatus(ctx context.Context, vehicleIDs []string, status string) error
	
	// Statistics
	GetStats(ctx context.Context, companyID string) (map[string]interface{}, error)
	
	// Location updates
	UpdateLocation(ctx context.Context, vehicleID string, lat, lng float64, timestamp time.Time) error
}

// vehicleRepository implements VehicleRepository interface
type vehicleRepository struct {
	db              *gorm.DB
	optimizedQueries *OptimizedVehicleQueries
}

// NewVehicleRepository creates a new vehicle repository
func NewVehicleRepository(db *gorm.DB) VehicleRepository {
	return &vehicleRepository{
		db:              db,
		optimizedQueries: NewOptimizedVehicleQueries(db),
	}
}

// Create creates a new vehicle
func (r *vehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	return r.db.WithContext(ctx).Create(vehicle).Error
}

// FindByID finds a vehicle by ID with company isolation
// For super-admin access (cross-company), pass empty string for companyID
func (r *vehicleRepository) FindByID(ctx context.Context, id, companyID string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	query := r.db.WithContext(ctx).Preload("Driver")
	
	// Company isolation: filter by company_id if provided
	// Empty companyID allows super-admin to access any company's vehicle
	if companyID != "" {
		query = query.Where("id = ? AND company_id = ?", id, companyID)
	} else {
		query = query.Where("id = ?", id)
	}
	
	if err := query.First(&vehicle).Error; err != nil {
		return nil, err
	}
	return &vehicle, nil
}

// FindByCompanyID finds vehicles by company ID with filters and pagination
func (r *vehicleRepository) FindByCompanyID(ctx context.Context, companyID string, filters VehicleFilters) ([]*models.Vehicle, int64, error) {
	return r.optimizedQueries.ListVehiclesOptimized(ctx, companyID, filters)
}

// Update updates a vehicle
func (r *vehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	return r.db.WithContext(ctx).Save(vehicle).Error
}

// Delete soft deletes a vehicle
func (r *vehicleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Vehicle{}, "id = ?", id).Error
}

// FindByStatus finds vehicles by status
func (r *vehicleRepository) FindByStatus(ctx context.Context, companyID string, status string) ([]*models.Vehicle, error) {
	return r.optimizedQueries.GetVehiclesByStatusOptimized(ctx, companyID, status)
}

// FindActive finds active vehicles
func (r *vehicleRepository) FindActive(ctx context.Context, companyID string) ([]*models.Vehicle, error) {
	return r.optimizedQueries.GetActiveVehiclesOptimized(ctx, companyID)
}

// FindByLicensePlate finds a vehicle by license plate
func (r *vehicleRepository) FindByLicensePlate(ctx context.Context, licensePlate string) (*models.Vehicle, error) {
	return r.optimizedQueries.GetVehicleByLicensePlateOptimized(ctx, licensePlate)
}

// FindByVIN finds a vehicle by VIN
func (r *vehicleRepository) FindByVIN(ctx context.Context, vin string) (*models.Vehicle, error) {
	return r.optimizedQueries.GetVehicleByVINOptimized(ctx, vin)
}

// FindByMakeModel finds vehicles by make and model
func (r *vehicleRepository) FindByMakeModel(ctx context.Context, companyID, make, model string) ([]*models.Vehicle, error) {
	return r.optimizedQueries.GetVehiclesByMakeModelOptimized(ctx, companyID, make, model)
}

// FindByYearRange finds vehicles by year range
func (r *vehicleRepository) FindByYearRange(ctx context.Context, companyID string, startYear, endYear int) ([]*models.Vehicle, error) {
	return r.optimizedQueries.GetVehiclesByYearRangeOptimized(ctx, companyID, startYear, endYear)
}

// FindByFuelType finds vehicles by fuel type
func (r *vehicleRepository) FindByFuelType(ctx context.Context, companyID string, fuelType string) ([]*models.Vehicle, error) {
	return r.optimizedQueries.GetVehiclesByFuelTypeOptimized(ctx, companyID, fuelType)
}

// FindWithDriver finds vehicles with assigned drivers
func (r *vehicleRepository) FindWithDriver(ctx context.Context, companyID string) ([]*models.Vehicle, error) {
	return r.optimizedQueries.GetVehiclesWithDriverOptimized(ctx, companyID)
}

// FindWithoutDriver finds vehicles without assigned drivers
func (r *vehicleRepository) FindWithoutDriver(ctx context.Context, companyID string) ([]*models.Vehicle, error) {
	return r.optimizedQueries.GetVehiclesWithoutDriverOptimized(ctx, companyID)
}

// FindNeedingInspection finds vehicles needing inspection
func (r *vehicleRepository) FindNeedingInspection(ctx context.Context, companyID string, daysBefore int) ([]*models.Vehicle, error) {
	return r.optimizedQueries.GetVehiclesNeedingInspectionOptimized(ctx, companyID, daysBefore)
}

// BatchUpdateStatus updates status for multiple vehicles
func (r *vehicleRepository) BatchUpdateStatus(ctx context.Context, vehicleIDs []string, status string) error {
	return r.optimizedQueries.BatchUpdateVehicleStatusOptimized(ctx, vehicleIDs, status, "")
}

// GetStats gets vehicle statistics
func (r *vehicleRepository) GetStats(ctx context.Context, companyID string) (map[string]interface{}, error) {
	return r.optimizedQueries.GetVehicleStatsOptimized(ctx, companyID)
}

// UpdateLocation updates vehicle location
func (r *vehicleRepository) UpdateLocation(ctx context.Context, vehicleID string, lat, lng float64, timestamp time.Time) error {
	return r.optimizedQueries.UpdateVehicleLocationOptimized(ctx, vehicleID, lat, lng, timestamp)
}

