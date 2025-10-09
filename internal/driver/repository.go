package driver

import (
	"context"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// DriverRepository defines the interface for driver data operations
type DriverRepository interface {
	// CRUD operations
	Create(ctx context.Context, driver *models.Driver) error
	FindByID(ctx context.Context, id, companyID string) (*models.Driver, error)
	FindByCompanyID(ctx context.Context, companyID string, filters DriverFilters) ([]*models.Driver, int64, error)
	Update(ctx context.Context, driver *models.Driver) error
	Delete(ctx context.Context, id string) error
	
	// Query operations
	FindByStatus(ctx context.Context, companyID string, status string) ([]*models.Driver, error)
	FindActive(ctx context.Context, companyID string) ([]*models.Driver, error)
	FindByNIK(ctx context.Context, nik string) (*models.Driver, error)
	FindBySIM(ctx context.Context, simNumber string) (*models.Driver, error)
	FindByPerformanceGrade(ctx context.Context, companyID string, grade string) ([]*models.Driver, error)
	FindWithVehicle(ctx context.Context, companyID string) ([]*models.Driver, error)
	FindWithoutVehicle(ctx context.Context, companyID string) ([]*models.Driver, error)
	FindAvailable(ctx context.Context, companyID string) ([]*models.Driver, error)
	FindExpiringSIM(ctx context.Context, companyID string, daysBefore int) ([]*models.Driver, error)
	FindExpiringMedical(ctx context.Context, companyID string, daysBefore int) ([]*models.Driver, error)
	FindNonCompliant(ctx context.Context, companyID string) ([]*models.Driver, error)
	
	// Performance operations
	UpdatePerformanceGrade(ctx context.Context, driverID string, grade string) error
	GetPerformanceStats(ctx context.Context, driverID string, startDate, endDate time.Time) (map[string]interface{}, error)
	
	// Statistics
	GetStats(ctx context.Context, companyID string) (map[string]interface{}, error)
	
	// Bulk operations
	BatchUpdateStatus(ctx context.Context, driverIDs []string, status string) error
}

// driverRepository implements DriverRepository interface
type driverRepository struct {
	db              *gorm.DB
	optimizedQueries *OptimizedDriverQueries
}

// NewDriverRepository creates a new driver repository
func NewDriverRepository(db *gorm.DB) DriverRepository {
	return &driverRepository{
		db:              db,
		optimizedQueries: NewOptimizedDriverQueries(db),
	}
}

// Create creates a new driver
func (r *driverRepository) Create(ctx context.Context, driver *models.Driver) error {
	return r.db.WithContext(ctx).Create(driver).Error
}

// FindByID finds a driver by ID with company isolation
// For super-admin access (cross-company), pass empty string for companyID
func (r *driverRepository) FindByID(ctx context.Context, id, companyID string) (*models.Driver, error) {
	var driver models.Driver
	query := r.db.WithContext(ctx).Preload("Vehicle")
	
	// Company isolation: filter by company_id if provided
	// Empty companyID allows super-admin to access any company's driver
	if companyID != "" {
		query = query.Where("id = ? AND company_id = ?", id, companyID)
	} else {
		query = query.Where("id = ?", id)
	}
	
	if err := query.First(&driver).Error; err != nil {
		return nil, err
	}
	return &driver, nil
}

// FindByCompanyID finds drivers by company ID with filters and pagination
func (r *driverRepository) FindByCompanyID(ctx context.Context, companyID string, filters DriverFilters) ([]*models.Driver, int64, error) {
	return r.optimizedQueries.ListDriversOptimized(ctx, companyID, filters)
}

// Update updates a driver
func (r *driverRepository) Update(ctx context.Context, driver *models.Driver) error {
	return r.db.WithContext(ctx).Save(driver).Error
}

// Delete soft deletes a driver
func (r *driverRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Driver{}, "id = ?", id).Error
}

// FindByStatus finds drivers by status
func (r *driverRepository) FindByStatus(ctx context.Context, companyID string, status string) ([]*models.Driver, error) {
	return r.optimizedQueries.GetDriversByStatusOptimized(ctx, companyID, status)
}

// FindActive finds active drivers
func (r *driverRepository) FindActive(ctx context.Context, companyID string) ([]*models.Driver, error) {
	return r.optimizedQueries.GetActiveDriversOptimized(ctx, companyID)
}

// FindByNIK finds a driver by NIK
func (r *driverRepository) FindByNIK(ctx context.Context, nik string) (*models.Driver, error) {
	return r.optimizedQueries.GetDriverByNIKOptimized(ctx, nik)
}

// FindBySIM finds a driver by SIM number
func (r *driverRepository) FindBySIM(ctx context.Context, simNumber string) (*models.Driver, error) {
	return r.optimizedQueries.GetDriverBySIMNumberOptimized(ctx, simNumber)
}

// FindByPerformanceGrade finds drivers by performance grade (minimum score)
func (r *driverRepository) FindByPerformanceGrade(ctx context.Context, companyID string, grade string) ([]*models.Driver, error) {
	// Convert grade to minimum score
	minScore := 0.0
	switch grade {
	case "A":
		minScore = 90.0
	case "B":
		minScore = 80.0
	case "C":
		minScore = 70.0
	}
	return r.optimizedQueries.GetDriversByPerformanceScoreOptimized(ctx, companyID, minScore)
}

// FindWithVehicle finds drivers with assigned vehicles
func (r *driverRepository) FindWithVehicle(ctx context.Context, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	if err := r.db.WithContext(ctx).
		Where("company_id = ? AND vehicle_id IS NOT NULL", companyID).
		Preload("Vehicle").
		Find(&drivers).Error; err != nil {
		return nil, err
	}
	return drivers, nil
}

// FindWithoutVehicle finds drivers without assigned vehicles
func (r *driverRepository) FindWithoutVehicle(ctx context.Context, companyID string) ([]*models.Driver, error) {
	return r.optimizedQueries.GetUnassignedDriversOptimized(ctx, companyID)
}

// FindAvailable finds available drivers
func (r *driverRepository) FindAvailable(ctx context.Context, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	if err := r.db.WithContext(ctx).
		Where("company_id = ? AND is_available = ? AND is_active = ?", companyID, true, true).
		Find(&drivers).Error; err != nil {
		return nil, err
	}
	return drivers, nil
}

// FindExpiringSIM finds drivers with expiring SIM
func (r *driverRepository) FindExpiringSIM(ctx context.Context, companyID string, daysBefore int) ([]*models.Driver, error) {
	return r.optimizedQueries.GetDriversByLicenseExpiryOptimized(ctx, companyID, daysBefore)
}

// FindExpiringMedical finds drivers with expiring medical certificates
func (r *driverRepository) FindExpiringMedical(ctx context.Context, companyID string, daysBefore int) ([]*models.Driver, error) {
	return r.optimizedQueries.GetDriversNeedingMedicalCheckupOptimized(ctx, companyID, daysBefore)
}

// FindNonCompliant finds non-compliant drivers
func (r *driverRepository) FindNonCompliant(ctx context.Context, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Where("company_id = ? AND (medical_checkup_date <= ? OR training_expiry <= ? OR sim_expiry <= ?)", 
			companyID, now, now, now).
		Find(&drivers).Error; err != nil {
		return nil, err
	}
	return drivers, nil
}

// UpdatePerformanceGrade updates driver performance grade
func (r *driverRepository) UpdatePerformanceGrade(ctx context.Context, driverID string, grade string) error {
	return r.db.WithContext(ctx).
		Model(&models.Driver{}).
		Where("id = ?", driverID).
		Update("performance_grade", grade).Error
}

// GetPerformanceStats gets driver performance statistics
func (r *driverRepository) GetPerformanceStats(ctx context.Context, driverID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	var performanceLogs []*models.PerformanceLog
	if err := r.db.WithContext(ctx).
		Where("driver_id = ? AND date BETWEEN ? AND ?", driverID, startDate, endDate).
		Order("date DESC").
		Find(&performanceLogs).Error; err != nil {
		return nil, err
	}
	
	// Calculate stats from performance logs
	return map[string]interface{}{
		"total_logs": len(performanceLogs),
	}, nil
}

// GetStats gets driver statistics
func (r *driverRepository) GetStats(ctx context.Context, companyID string) (map[string]interface{}, error) {
	return r.optimizedQueries.GetDriverStatsOptimized(ctx, companyID)
}

// BatchUpdateStatus updates status for multiple drivers
func (r *driverRepository) BatchUpdateStatus(ctx context.Context, driverIDs []string, status string) error {
	return r.optimizedQueries.BatchUpdateDriverStatusOptimized(ctx, driverIDs, status, "")
}

