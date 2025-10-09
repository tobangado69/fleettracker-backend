package driver

import (
	"context"
	"fmt"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// OptimizedDriverQueries provides optimized database queries for driver operations
type OptimizedDriverQueries struct {
	db *gorm.DB
}

// NewOptimizedDriverQueries creates a new optimized driver queries service
func NewOptimizedDriverQueries(db *gorm.DB) *OptimizedDriverQueries {
	return &OptimizedDriverQueries{db: db}
}

// ListDriversOptimized lists drivers with optimized query and pagination
func (odq *OptimizedDriverQueries) ListDriversOptimized(ctx context.Context, companyID string, filters DriverFilters) ([]*models.Driver, int64, error) {
	var drivers []*models.Driver
	var total int64

	// Build base query with company filter
	query := odq.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ?", companyID)

	// Apply filters with optimized conditions
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.EmploymentStatus != nil {
		query = query.Where("employment_status = ?", *filters.EmploymentStatus)
	}
	if filters.PerformanceGrade != nil {
		query = query.Where("performance_grade = ?", *filters.PerformanceGrade)
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
		query = query.Where("is_available = ?", *filters.IsAvailable)
	}
	if filters.IsCompliant != nil {
		if *filters.IsCompliant {
			query = query.Where("medical_checkup_date > NOW() AND training_expiry > NOW() AND sim_expiry > NOW()")
		} else {
			query = query.Where("(medical_checkup_date <= NOW() OR training_expiry <= NOW() OR sim_expiry <= NOW())")
		}
	}
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + *filters.Search + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR phone_number ILIKE ? OR nik ILIKE ?", 
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Get total count with optimized query
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count drivers: %w", err)
	}

	// Apply sorting with index-friendly ordering
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
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

	// Execute query with selective preloading
	if err := query.Preload("Vehicle", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, license_plate, make, model, status")
	}).Find(&drivers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list drivers: %w", err)
	}

	return drivers, total, nil
}

// GetDriversByStatusOptimized gets drivers by status with optimized query
func (odq *OptimizedDriverQueries) GetDriversByStatusOptimized(ctx context.Context, companyID string, status string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	
	// Use composite index on (company_id, status)
	if err := odq.db.WithContext(ctx).
		Where("company_id = ? AND status = ?", companyID, status).
		Order("created_at DESC").
		Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by status: %w", err)
	}
	
	return drivers, nil
}

// GetActiveDriversOptimized gets active drivers with optimized query
func (odq *OptimizedDriverQueries) GetActiveDriversOptimized(ctx context.Context, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	
	// Use composite index on (company_id, is_active)
	if err := odq.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ?", companyID, true).
		Order("created_at DESC").
		Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get active drivers: %w", err)
	}
	
	return drivers, nil
}

// GetDriversNeedingMedicalCheckupOptimized gets drivers needing medical checkup
func (odq *OptimizedDriverQueries) GetDriversNeedingMedicalCheckupOptimized(ctx context.Context, companyID string, daysBefore int) ([]*models.Driver, error) {
	var drivers []*models.Driver
	cutoffDate := time.Now().AddDate(0, 0, daysBefore)
	
	// Use index on medical_checkup_date with optimized condition
	if err := odq.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ? AND (medical_checkup_date IS NULL OR medical_checkup_date <= ?)", 
			companyID, true, cutoffDate).
		Order("medical_checkup_date ASC NULLS LAST").
		Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers needing medical checkup: %w", err)
	}
	
	return drivers, nil
}

// GetDriversNeedingTrainingOptimized gets drivers needing training
func (odq *OptimizedDriverQueries) GetDriversNeedingTrainingOptimized(ctx context.Context, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	cutoffDate := time.Now()
	
	// Use index on training_expiry with optimized condition
	if err := odq.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ? AND (training_expiry IS NULL OR training_expiry <= ?)", 
			companyID, true, cutoffDate).
		Order("training_expiry ASC NULLS LAST").
		Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers needing training: %w", err)
	}
	
	return drivers, nil
}

// GetDriversByPerformanceScoreOptimized gets drivers by performance score
func (odq *OptimizedDriverQueries) GetDriversByPerformanceScoreOptimized(ctx context.Context, companyID string, minScore float64) ([]*models.Driver, error) {
	var drivers []*models.Driver
	
	// Use index on overall_score with optimized condition
	if err := odq.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ? AND overall_score >= ?", 
			companyID, true, minScore).
		Order("overall_score DESC").
		Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by performance score: %w", err)
	}
	
	return drivers, nil
}

// GetDriverWithVehicleOptimized gets driver with assigned vehicle using optimized join
func (odq *OptimizedDriverQueries) GetDriverWithVehicleOptimized(ctx context.Context, driverID string) (*models.Driver, error) {
	var driver models.Driver
	
	// Use optimized join with selective fields
	if err := odq.db.WithContext(ctx).
		Preload("Vehicle", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, license_plate, make, model, status, is_active")
		}).
		Where("id = ?", driverID).
		First(&driver).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("driver with ID %s not found", driverID)
		}
		return nil, fmt.Errorf("failed to get driver with vehicle: %w", err)
	}
	
	return &driver, nil
}

// GetUnassignedDriversOptimized gets drivers without assigned vehicles
func (odq *OptimizedDriverQueries) GetUnassignedDriversOptimized(ctx context.Context, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	
	// Use optimized condition for unassigned drivers
	if err := odq.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ? AND vehicle_id IS NULL", 
			companyID, true).
		Order("created_at DESC").
		Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get unassigned drivers: %w", err)
	}
	
	return drivers, nil
}

// GetDriverByNIKOptimized gets driver by NIK with optimized query
func (odq *OptimizedDriverQueries) GetDriverByNIKOptimized(ctx context.Context, nik string) (*models.Driver, error) {
	var driver models.Driver
	
	// Use index on NIK
	if err := odq.db.WithContext(ctx).
		Where("nik = ?", nik).
		First(&driver).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("driver with NIK %s not found", nik)
		}
		return nil, fmt.Errorf("failed to get driver by NIK: %w", err)
	}
	
	return &driver, nil
}

// GetDriverBySIMNumberOptimized gets driver by SIM number with optimized query
func (odq *OptimizedDriverQueries) GetDriverBySIMNumberOptimized(ctx context.Context, simNumber string) (*models.Driver, error) {
	var driver models.Driver
	
	// Use index on sim_number
	if err := odq.db.WithContext(ctx).
		Where("sim_number = ?", simNumber).
		First(&driver).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("driver with SIM number %s not found", simNumber)
		}
		return nil, fmt.Errorf("failed to get driver by SIM number: %w", err)
	}
	
	return &driver, nil
}

// UpdatePerformanceScoreOptimized updates driver performance scores
func (odq *OptimizedDriverQueries) UpdatePerformanceScoreOptimized(ctx context.Context, driverID string, scores map[string]float64) error {
	// Calculate overall score
	var overallScore float64
	count := 0
	for _, score := range scores {
		overallScore += score
		count++
	}
	if count > 0 {
		overallScore = overallScore / float64(count)
	}
	
	updates := map[string]interface{}{
		"overall_score": overallScore,
		"updated_at":    time.Now(),
	}
	
	// Add individual scores
	for key, value := range scores {
		updates[key] = value
	}
	
	// Use direct update for better performance
	if err := odq.db.WithContext(ctx).
		Model(&models.Driver{}).
		Where("id = ?", driverID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update performance score: %w", err)
	}
	
	return nil
}

// GetDriverEventsOptimized gets driver events with optimized query
func (odq *OptimizedDriverQueries) GetDriverEventsOptimized(ctx context.Context, driverID string, startTime, endTime time.Time, eventType string) ([]*models.DriverEvent, error) {
	var events []*models.DriverEvent
	
	// Build query with optional event type filter
	query := odq.db.WithContext(ctx).
		Where("driver_id = ? AND timestamp BETWEEN ? AND ?", driverID, startTime, endTime)
	
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	
	if err := query.Order("timestamp DESC").Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get driver events: %w", err)
	}
	
	return events, nil
}

// GetDriverStatsOptimized gets driver statistics with optimized query
func (odq *OptimizedDriverQueries) GetDriverStatsOptimized(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats struct {
		TotalDrivers        int64 `gorm:"column:total_drivers"`
		ActiveDrivers       int64 `gorm:"column:active_drivers"`
		DriversWithVehicle  int64 `gorm:"column:drivers_with_vehicle"`
		DriversWithoutVehicle int64 `gorm:"column:drivers_without_vehicle"`
		DriversNeedingMedical int64 `gorm:"column:drivers_needing_medical"`
		DriversNeedingTraining int64 `gorm:"column:drivers_needing_training"`
		HighPerformers      int64 `gorm:"column:high_performers"`
	}
	
	// Use single query with conditional aggregation for better performance
	query := `
		SELECT 
			COUNT(*) as total_drivers,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_drivers,
			COUNT(CASE WHEN vehicle_id IS NOT NULL THEN 1 END) as drivers_with_vehicle,
			COUNT(CASE WHEN vehicle_id IS NULL THEN 1 END) as drivers_without_vehicle,
			COUNT(CASE WHEN medical_checkup_date IS NULL OR medical_checkup_date <= NOW() THEN 1 END) as drivers_needing_medical,
			COUNT(CASE WHEN training_expiry IS NULL OR training_expiry <= NOW() THEN 1 END) as drivers_needing_training,
			COUNT(CASE WHEN overall_score >= 80 THEN 1 END) as high_performers
		FROM drivers 
		WHERE company_id = ?
	`
	
	if err := odq.db.WithContext(ctx).Raw(query, companyID).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get driver stats: %w", err)
	}
	
	return map[string]interface{}{
		"total_drivers":         stats.TotalDrivers,
		"active_drivers":        stats.ActiveDrivers,
		"drivers_with_vehicle":  stats.DriversWithVehicle,
		"drivers_without_vehicle": stats.DriversWithoutVehicle,
		"drivers_needing_medical": stats.DriversNeedingMedical,
		"drivers_needing_training": stats.DriversNeedingTraining,
		"high_performers":       stats.HighPerformers,
	}, nil
}

// GetDriversByLicenseTypeOptimized gets drivers by license type
func (odq *OptimizedDriverQueries) GetDriversByLicenseTypeOptimized(ctx context.Context, companyID string, licenseType string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	
	// Use index on license_type
	if err := odq.db.WithContext(ctx).
		Where("company_id = ? AND license_type = ?", companyID, licenseType).
		Order("created_at DESC").
		Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by license type: %w", err)
	}
	
	return drivers, nil
}

// GetDriversByLicenseExpiryOptimized gets drivers with expiring licenses
func (odq *OptimizedDriverQueries) GetDriversByLicenseExpiryOptimized(ctx context.Context, companyID string, daysBefore int) ([]*models.Driver, error) {
	var drivers []*models.Driver
	cutoffDate := time.Now().AddDate(0, 0, daysBefore)
	
	// Use index on license_expiry with optimized condition
	if err := odq.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ? AND license_expiry IS NOT NULL AND license_expiry <= ?", 
			companyID, true, cutoffDate).
		Order("license_expiry ASC").
		Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by license expiry: %w", err)
	}
	
	return drivers, nil
}

// BatchUpdateDriverStatusOptimized updates multiple driver statuses in a single transaction
func (odq *OptimizedDriverQueries) BatchUpdateDriverStatusOptimized(ctx context.Context, driverIDs []string, status string, reason string) error {
	if len(driverIDs) == 0 {
		return nil
	}
	
	// Use batch update for better performance
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	
	if err := odq.db.WithContext(ctx).
		Model(&models.Driver{}).
		Where("id IN ?", driverIDs).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to batch update driver status: %w", err)
	}
	
	return nil
}

// GetDriverPerformanceHistoryOptimized gets driver performance history
func (odq *OptimizedDriverQueries) GetDriverPerformanceHistoryOptimized(ctx context.Context, driverID string, startTime, endTime time.Time) ([]*models.PerformanceLog, error) {
	var logs []*models.PerformanceLog
	
	// Use index on (driver_id, log_date)
	if err := odq.db.WithContext(ctx).
		Where("driver_id = ? AND log_date BETWEEN ? AND ?", driverID, startTime, endTime).
		Order("log_date DESC").
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get driver performance history: %w", err)
	}
	
	return logs, nil
}
