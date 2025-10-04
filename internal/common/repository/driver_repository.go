package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// DriverRepositoryImpl implements the DriverRepository interface
type DriverRepositoryImpl struct {
	*BaseRepository[models.Driver]
}

// NewDriverRepository creates a new driver repository
func NewDriverRepository(db *gorm.DB) DriverRepository {
	return &DriverRepositoryImpl{
		BaseRepository: NewBaseRepository[models.Driver](db),
	}
}

// GetByCompany retrieves drivers by company ID with pagination
func (r *DriverRepositoryImpl) GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Driver, error) {
	var drivers []*models.Driver
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by company: %w", err)
	}
	
	return drivers, nil
}

// GetByVehicle retrieves the driver assigned to a specific vehicle
func (r *DriverRepositoryImpl) GetByVehicle(ctx context.Context, vehicleID string) (*models.Driver, error) {
	var driver models.Driver
	if err := r.db.WithContext(ctx).Where("vehicle_id = ?", vehicleID).First(&driver).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no driver assigned to vehicle: %s", vehicleID)
		}
		return nil, fmt.Errorf("failed to get driver by vehicle: %w", err)
	}
	return &driver, nil
}

// GetByStatus retrieves drivers by status within a company
func (r *DriverRepositoryImpl) GetByStatus(ctx context.Context, companyID string, status string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	if err := r.db.WithContext(ctx).Where("company_id = ? AND status = ?", companyID, status).Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by status: %w", err)
	}
	return drivers, nil
}

// GetBySIMType retrieves drivers by SIM type within a company
func (r *DriverRepositoryImpl) GetBySIMType(ctx context.Context, companyID string, simType string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	if err := r.db.WithContext(ctx).Where("company_id = ? AND sim_type = ?", companyID, simType).Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by SIM type: %w", err)
	}
	return drivers, nil
}

// SearchByNIK searches drivers by NIK within a company
func (r *DriverRepositoryImpl) SearchByNIK(ctx context.Context, nik string, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	if err := r.db.WithContext(ctx).Where("company_id = ? AND nik = ?", companyID, nik).Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to search drivers by NIK: %w", err)
	}
	return drivers, nil
}

// SearchBySIM searches drivers by SIM number within a company
func (r *DriverRepositoryImpl) SearchBySIM(ctx context.Context, simNumber string, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	if err := r.db.WithContext(ctx).Where("company_id = ? AND sim_number = ?", companyID, simNumber).Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to search drivers by SIM: %w", err)
	}
	return drivers, nil
}

// GetAvailableDrivers retrieves drivers that are available for assignment
func (r *DriverRepositoryImpl) GetAvailableDrivers(ctx context.Context, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	if err := r.db.WithContext(ctx).Where("company_id = ? AND status = ? AND vehicle_id IS NULL", companyID, "available").Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get available drivers: %w", err)
	}
	return drivers, nil
}

// GetDriversNeedingTraining retrieves drivers who need training
func (r *DriverRepositoryImpl) GetDriversNeedingTraining(ctx context.Context, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	query := `
		SELECT * FROM drivers 
		WHERE company_id = ? 
		AND (training_completed = false 
			OR next_training_date IS NULL 
			OR next_training_date <= CURRENT_DATE)
		AND status = 'active'
	`
	
	if err := r.db.WithContext(ctx).Raw(query, companyID).Scan(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers needing training: %w", err)
	}
	
	return drivers, nil
}

// GetDriversWithExpiredSIM retrieves drivers with expired SIM
func (r *DriverRepositoryImpl) GetDriversWithExpiredSIM(ctx context.Context, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	query := `
		SELECT * FROM drivers 
		WHERE company_id = ? 
		AND (sim_expiry_date IS NOT NULL 
			AND sim_expiry_date <= CURRENT_DATE)
		AND status = 'active'
	`
	
	if err := r.db.WithContext(ctx).Raw(query, companyID).Scan(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers with expired SIM: %w", err)
	}
	
	return drivers, nil
}

// GetDriversWithExpiredMedicalCheckup retrieves drivers with expired medical checkup
func (r *DriverRepositoryImpl) GetDriversWithExpiredMedicalCheckup(ctx context.Context, companyID string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	query := `
		SELECT * FROM drivers 
		WHERE company_id = ? 
		AND (medical_checkup_expiry IS NOT NULL 
			AND medical_checkup_expiry <= CURRENT_DATE)
		AND status = 'active'
	`
	
	if err := r.db.WithContext(ctx).Raw(query, companyID).Scan(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers with expired medical checkup: %w", err)
	}
	
	return drivers, nil
}

// UpdateStatus updates the status of a driver
func (r *DriverRepositoryImpl) UpdateStatus(ctx context.Context, driverID string, status string) error {
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("id = ?", driverID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update driver status: %w", err)
	}
	return nil
}

// UpdatePerformance updates the performance of a driver
func (r *DriverRepositoryImpl) UpdatePerformance(ctx context.Context, driverID string, performance models.PerformanceLog) error {
	// Update driver performance scores
	updates := map[string]interface{}{
		"performance_score":    performance.OverallScore,
		"safety_score":        performance.SafetyScore,
		"efficiency_score":    performance.EfficiencyScore,
	}
	
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("id = ?", driverID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update driver performance: %w", err)
	}
	
	// Create performance log entry
	if err := r.db.WithContext(ctx).Create(&performance).Error; err != nil {
		return fmt.Errorf("failed to create performance log: %w", err)
	}
	
	return nil
}

// AssignVehicle assigns a vehicle to a driver
func (r *DriverRepositoryImpl) AssignVehicle(ctx context.Context, driverID string, vehicleID *string) error {
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("id = ?", driverID).Update("vehicle_id", vehicleID).Error; err != nil {
		return fmt.Errorf("failed to assign vehicle to driver: %w", err)
	}
	return nil
}

// GetDriversByAge retrieves drivers by age range within a company
func (r *DriverRepositoryImpl) GetDriversByAge(ctx context.Context, companyID string, minAge, maxAge int) ([]*models.Driver, error) {
	var drivers []*models.Driver
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	if minAge > 0 {
		query = query.Where("EXTRACT(YEAR FROM AGE(date_of_birth)) >= ?", minAge)
	}
	if maxAge > 0 {
		query = query.Where("EXTRACT(YEAR FROM AGE(date_of_birth)) <= ?", maxAge)
	}
	
	if err := query.Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by age: %w", err)
	}
	
	return drivers, nil
}

// GetDriversByHireDate retrieves drivers by hire date range within a company
func (r *DriverRepositoryImpl) GetDriversByHireDate(ctx context.Context, companyID string, startDate, endDate string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	if startDate != "" {
		query = query.Where("hire_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("hire_date <= ?", endDate)
	}
	
	if err := query.Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by hire date: %w", err)
	}
	
	return drivers, nil
}

// GetDriversByPerformanceRange retrieves drivers by performance score range
func (r *DriverRepositoryImpl) GetDriversByPerformanceRange(ctx context.Context, companyID string, minScore, maxScore float64) ([]*models.Driver, error) {
	var drivers []*models.Driver
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	if minScore >= 0 {
		query = query.Where("performance_score >= ?", minScore)
	}
	if maxScore <= 100 {
		query = query.Where("performance_score <= ?", maxScore)
	}
	
	if err := query.Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by performance range: %w", err)
	}
	
	return drivers, nil
}

// GetDriverStatistics retrieves driver statistics for a company
func (r *DriverRepositoryImpl) GetDriverStatistics(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats struct {
		TotalDrivers           int64   `json:"total_drivers"`
		ActiveDrivers          int64   `json:"active_drivers"`
		AvailableDrivers       int64   `json:"available_drivers"`
		BusyDrivers            int64   `json:"busy_drivers"`
		OfflineDrivers         int64   `json:"offline_drivers"`
		AssignedDrivers        int64   `json:"assigned_drivers"`
		UnassignedDrivers      int64   `json:"unassigned_drivers"`
		AveragePerformance     float64 `json:"average_performance"`
		DriversWithExpiredSIM  int64   `json:"drivers_with_expired_sim"`
		DriversNeedingTraining int64   `json:"drivers_needing_training"`
	}

	// Get total drivers
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ?", companyID).Count(&stats.TotalDrivers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total drivers: %w", err)
	}

	// Get active drivers
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ? AND status = ?", companyID, "active").Count(&stats.ActiveDrivers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active drivers: %w", err)
	}

	// Get available drivers
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ? AND status = ?", companyID, "available").Count(&stats.AvailableDrivers).Error; err != nil {
		return nil, fmt.Errorf("failed to count available drivers: %w", err)
	}

	// Get busy drivers
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ? AND status = ?", companyID, "busy").Count(&stats.BusyDrivers).Error; err != nil {
		return nil, fmt.Errorf("failed to count busy drivers: %w", err)
	}

	// Get offline drivers
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ? AND status = ?", companyID, "offline").Count(&stats.OfflineDrivers).Error; err != nil {
		return nil, fmt.Errorf("failed to count offline drivers: %w", err)
	}

	// Get assigned drivers
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ? AND vehicle_id IS NOT NULL", companyID).Count(&stats.AssignedDrivers).Error; err != nil {
		return nil, fmt.Errorf("failed to count assigned drivers: %w", err)
	}

	// Get unassigned drivers
	stats.UnassignedDrivers = stats.TotalDrivers - stats.AssignedDrivers

	// Get average performance
	var avgPerformance struct {
		Average float64
	}
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ?", companyID).Select("AVG(performance_score) as average").Scan(&avgPerformance).Error; err != nil {
		return nil, fmt.Errorf("failed to get average performance: %w", err)
	}
	stats.AveragePerformance = avgPerformance.Average

	// Get drivers with expired SIM
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ? AND sim_expiry_date <= CURRENT_DATE", companyID).Count(&stats.DriversWithExpiredSIM).Error; err != nil {
		return nil, fmt.Errorf("failed to count drivers with expired SIM: %w", err)
	}

	// Get drivers needing training
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ? AND (training_completed = false OR next_training_date <= CURRENT_DATE)", companyID).Count(&stats.DriversNeedingTraining).Error; err != nil {
		return nil, fmt.Errorf("failed to count drivers needing training: %w", err)
	}

	return map[string]interface{}{
		"total_drivers":             stats.TotalDrivers,
		"active_drivers":            stats.ActiveDrivers,
		"available_drivers":         stats.AvailableDrivers,
		"busy_drivers":              stats.BusyDrivers,
		"offline_drivers":           stats.OfflineDrivers,
		"assigned_drivers":          stats.AssignedDrivers,
		"unassigned_drivers":        stats.UnassignedDrivers,
		"average_performance":       stats.AveragePerformance,
		"drivers_with_expired_sim":  stats.DriversWithExpiredSIM,
		"drivers_needing_training":  stats.DriversNeedingTraining,
	}, nil
}

// SearchDrivers performs a comprehensive search across multiple driver fields
func (r *DriverRepositoryImpl) SearchDrivers(ctx context.Context, companyID string, searchQuery string, pagination Pagination) ([]*models.Driver, error) {
	var drivers []*models.Driver
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	query = query.Where(
		"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(nik) LIKE ? OR LOWER(sim_number) LIKE ? OR LOWER(phone) LIKE ? OR LOWER(email) LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
	)
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to search drivers: %w", err)
	}
	
	return drivers, nil
}

// BulkUpdateStatus updates the status of multiple drivers
func (r *DriverRepositoryImpl) BulkUpdateStatus(ctx context.Context, driverIDs []string, status string) error {
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("id IN ?", driverIDs).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to bulk update driver status: %w", err)
	}
	return nil
}

// GetDriversByCompliance retrieves drivers by compliance status
func (r *DriverRepositoryImpl) GetDriversByCompliance(ctx context.Context, companyID string, complianceType string) ([]*models.Driver, error) {
	var drivers []*models.Driver
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	switch complianceType {
	case "sim_expired":
		query = query.Where("sim_expiry_date <= CURRENT_DATE")
	case "medical_expired":
		query = query.Where("medical_checkup_expiry <= CURRENT_DATE")
	case "training_needed":
		query = query.Where("training_completed = false OR next_training_date <= CURRENT_DATE")
	case "compliant":
		query = query.Where("sim_expiry_date > CURRENT_DATE AND medical_checkup_expiry > CURRENT_DATE AND training_completed = true")
	}
	
	if err := query.Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get drivers by compliance: %w", err)
	}
	
	return drivers, nil
}

// GetTopPerformers retrieves the top performing drivers
func (r *DriverRepositoryImpl) GetTopPerformers(ctx context.Context, companyID string, limit int) ([]*models.Driver, error) {
	var drivers []*models.Driver
	if err := r.db.WithContext(ctx).Where("company_id = ?", companyID).Order("performance_score DESC").Limit(limit).Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to get top performers: %w", err)
	}
	return drivers, nil
}
