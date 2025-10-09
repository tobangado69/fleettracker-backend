package vehicle

import (
	"context"
	"fmt"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// OptimizedVehicleQueries provides optimized database queries for vehicle operations
type OptimizedVehicleQueries struct {
	db *gorm.DB
}

// NewOptimizedVehicleQueries creates a new optimized vehicle queries service
func NewOptimizedVehicleQueries(db *gorm.DB) *OptimizedVehicleQueries {
	return &OptimizedVehicleQueries{db: db}
}

// ListVehiclesOptimized lists vehicles with optimized query and pagination
func (ovq *OptimizedVehicleQueries) ListVehiclesOptimized(ctx context.Context, companyID string, filters VehicleFilters) ([]*models.Vehicle, int64, error) {
	var vehicles []*models.Vehicle
	var total int64

	// Build base query with company filter
	query := ovq.db.WithContext(ctx).Model(&models.Vehicle{}).Where("company_id = ?", companyID)

	// Apply filters with optimized conditions
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
		// Use full-text search for better performance
		query = query.Where("make ILIKE ? OR model ILIKE ? OR license_plate ILIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Get total count with optimized query
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count vehicles: %w", err)
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
	if err := query.Preload("Driver", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email, phone, status")
	}).Find(&vehicles).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list vehicles: %w", err)
	}

	return vehicles, total, nil
}

// GetVehiclesByStatusOptimized gets vehicles by status with optimized query
func (ovq *OptimizedVehicleQueries) GetVehiclesByStatusOptimized(ctx context.Context, companyID string, status string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	
	// Use composite index on (company_id, status)
	if err := ovq.db.WithContext(ctx).
		Where("company_id = ? AND status = ?", companyID, status).
		Order("created_at DESC").
		Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by status: %w", err)
	}
	
	return vehicles, nil
}

// GetActiveVehiclesOptimized gets active vehicles with optimized query
func (ovq *OptimizedVehicleQueries) GetActiveVehiclesOptimized(ctx context.Context, companyID string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	
	// Use composite index on (company_id, is_active)
	if err := ovq.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ?", companyID, true).
		Order("created_at DESC").
		Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get active vehicles: %w", err)
	}
	
	return vehicles, nil
}

// GetVehiclesNeedingInspectionOptimized gets vehicles needing inspection with optimized query
func (ovq *OptimizedVehicleQueries) GetVehiclesNeedingInspectionOptimized(ctx context.Context, companyID string, daysBefore int) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	cutoffDate := time.Now().AddDate(0, 0, daysBefore)
	
	// Use index on inspection_date with optimized condition
	if err := ovq.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ? AND (inspection_date IS NULL OR inspection_date <= ?)", 
			companyID, true, cutoffDate).
		Order("inspection_date ASC NULLS LAST").
		Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles needing inspection: %w", err)
	}
	
	return vehicles, nil
}

// GetVehiclesWithDriverOptimized gets vehicles with assigned drivers using optimized join
func (ovq *OptimizedVehicleQueries) GetVehiclesWithDriverOptimized(ctx context.Context, companyID string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	
	// Use optimized join with selective fields
	if err := ovq.db.WithContext(ctx).
		Select("vehicles.*").
		Joins("INNER JOIN drivers ON drivers.vehicle_id = vehicles.id AND drivers.is_active = ?", true).
		Where("vehicles.company_id = ? AND vehicles.is_active = ?", companyID, true).
		Order("vehicles.created_at DESC").
		Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles with driver: %w", err)
	}
	
	return vehicles, nil
}

// GetVehiclesWithoutDriverOptimized gets vehicles without assigned drivers using optimized query
func (ovq *OptimizedVehicleQueries) GetVehiclesWithoutDriverOptimized(ctx context.Context, companyID string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	
	// Use NOT EXISTS for better performance than NOT IN
	if err := ovq.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ? AND NOT EXISTS (SELECT 1 FROM drivers WHERE drivers.vehicle_id = vehicles.id AND drivers.is_active = ?)", 
			companyID, true, true).
		Order("created_at DESC").
		Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles without driver: %w", err)
	}
	
	return vehicles, nil
}

// GetVehicleByLicensePlateOptimized gets vehicle by license plate with optimized query
func (ovq *OptimizedVehicleQueries) GetVehicleByLicensePlateOptimized(ctx context.Context, licensePlate string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	
	// Use index on license_plate with case-insensitive search
	if err := ovq.db.WithContext(ctx).
		Where("LOWER(license_plate) = LOWER(?)", licensePlate).
		First(&vehicle).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("vehicle with license plate %s not found", licensePlate)
		}
		return nil, fmt.Errorf("failed to get vehicle by license plate: %w", err)
	}
	
	return &vehicle, nil
}

// GetVehicleByVINOptimized gets vehicle by VIN with optimized query
func (ovq *OptimizedVehicleQueries) GetVehicleByVINOptimized(ctx context.Context, vin string) (*models.Vehicle, error) {
	var vehicle models.Vehicle
	
	// Use index on VIN
	if err := ovq.db.WithContext(ctx).
		Where("vin = ?", vin).
		First(&vehicle).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("vehicle with VIN %s not found", vin)
		}
		return nil, fmt.Errorf("failed to get vehicle by VIN: %w", err)
	}
	
	return &vehicle, nil
}

// UpdateVehicleLocationOptimized updates vehicle location with optimized query
func (ovq *OptimizedVehicleQueries) UpdateVehicleLocationOptimized(ctx context.Context, vehicleID string, lat, lng float64, timestamp time.Time) error {
	updates := map[string]interface{}{
		"last_latitude":   lat,
		"last_longitude":  lng,
		"last_updated_at": timestamp,
		"updated_at":      time.Now(),
	}

	// Use direct update for better performance
	if err := ovq.db.WithContext(ctx).
		Model(&models.Vehicle{}).
		Where("id = ?", vehicleID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update vehicle location: %w", err)
	}
	
	return nil
}

// GetVehicleStatsOptimized gets vehicle statistics with optimized query
func (ovq *OptimizedVehicleQueries) GetVehicleStatsOptimized(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats struct {
		TotalVehicles      int64 `gorm:"column:total_vehicles"`
		ActiveVehicles     int64 `gorm:"column:active_vehicles"`
		VehiclesWithDriver int64 `gorm:"column:vehicles_with_driver"`
		VehiclesWithoutDriver int64 `gorm:"column:vehicles_without_driver"`
		GPSEnabledVehicles int64 `gorm:"column:gps_enabled_vehicles"`
		MaintenanceVehicles int64 `gorm:"column:maintenance_vehicles"`
	}
	
	// Use single query with conditional aggregation for better performance
	query := `
		SELECT 
			COUNT(*) as total_vehicles,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_vehicles,
			COUNT(CASE WHEN driver_id IS NOT NULL THEN 1 END) as vehicles_with_driver,
			COUNT(CASE WHEN driver_id IS NULL THEN 1 END) as vehicles_without_driver,
			COUNT(CASE WHEN is_gps_enabled = true THEN 1 END) as gps_enabled_vehicles,
			COUNT(CASE WHEN status = 'maintenance' THEN 1 END) as maintenance_vehicles
		FROM vehicles 
		WHERE company_id = ?
	`
	
	if err := ovq.db.WithContext(ctx).Raw(query, companyID).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicle stats: %w", err)
	}
	
	return map[string]interface{}{
		"total_vehicles":        stats.TotalVehicles,
		"active_vehicles":       stats.ActiveVehicles,
		"vehicles_with_driver":  stats.VehiclesWithDriver,
		"vehicles_without_driver": stats.VehiclesWithoutDriver,
		"gps_enabled_vehicles":  stats.GPSEnabledVehicles,
		"maintenance_vehicles":  stats.MaintenanceVehicles,
	}, nil
}

// GetVehiclesByMakeModelOptimized gets vehicles by make and model with optimized query
func (ovq *OptimizedVehicleQueries) GetVehiclesByMakeModelOptimized(ctx context.Context, companyID string, make, model string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	
	// Use composite index on (make, model)
	if err := ovq.db.WithContext(ctx).
		Where("company_id = ? AND make = ? AND model = ?", companyID, make, model).
		Order("year DESC, created_at DESC").
		Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by make and model: %w", err)
	}
	
	return vehicles, nil
}

// GetVehiclesByYearRangeOptimized gets vehicles by year range with optimized query
func (ovq *OptimizedVehicleQueries) GetVehiclesByYearRangeOptimized(ctx context.Context, companyID string, startYear, endYear int) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	
	// Use index on year with range condition
	if err := ovq.db.WithContext(ctx).
		Where("company_id = ? AND year BETWEEN ? AND ?", companyID, startYear, endYear).
		Order("year DESC, created_at DESC").
		Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by year range: %w", err)
	}
	
	return vehicles, nil
}

// GetVehiclesByFuelTypeOptimized gets vehicles by fuel type with optimized query
func (ovq *OptimizedVehicleQueries) GetVehiclesByFuelTypeOptimized(ctx context.Context, companyID string, fuelType string) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	
	// Use index on fuel_type
	if err := ovq.db.WithContext(ctx).
		Where("company_id = ? AND fuel_type = ?", companyID, fuelType).
		Order("created_at DESC").
		Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicles by fuel type: %w", err)
	}
	
	return vehicles, nil
}

// BatchUpdateVehicleStatusOptimized updates multiple vehicle statuses in a single transaction
func (ovq *OptimizedVehicleQueries) BatchUpdateVehicleStatusOptimized(ctx context.Context, vehicleIDs []string, status string, reason string) error {
	if len(vehicleIDs) == 0 {
		return nil
	}
	
	// Use batch update for better performance
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	
	if err := ovq.db.WithContext(ctx).
		Model(&models.Vehicle{}).
		Where("id IN ?", vehicleIDs).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to batch update vehicle status: %w", err)
	}
	
	return nil
}

// GetVehicleMaintenanceScheduleOptimized gets vehicles with upcoming maintenance
func (ovq *OptimizedVehicleQueries) GetVehicleMaintenanceScheduleOptimized(ctx context.Context, companyID string, daysAhead int) ([]*models.Vehicle, error) {
	var vehicles []*models.Vehicle
	cutoffDate := time.Now().AddDate(0, 0, daysAhead)
	
	// Use index on inspection_date with optimized condition
	if err := ovq.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ? AND inspection_date IS NOT NULL AND inspection_date <= ?", 
			companyID, true, cutoffDate).
		Order("inspection_date ASC").
		Find(&vehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to get vehicle maintenance schedule: %w", err)
	}
	
	return vehicles, nil
}
