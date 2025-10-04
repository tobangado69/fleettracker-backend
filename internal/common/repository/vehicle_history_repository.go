package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// VehicleHistoryRepository defines the interface for vehicle history operations
type VehicleHistoryRepository interface {
	Repository[models.VehicleHistory]
	
	// Vehicle-specific history operations
	GetByVehicle(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.VehicleHistory, error)
	GetByVehicleAndType(ctx context.Context, vehicleID, eventType string, pagination Pagination) ([]*models.VehicleHistory, error)
	GetByVehicleAndCategory(ctx context.Context, vehicleID, eventCategory string, pagination Pagination) ([]*models.VehicleHistory, error)
	GetMaintenanceHistory(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.VehicleHistory, error)
	GetRepairHistory(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.VehicleHistory, error)
	GetInspectionHistory(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.VehicleHistory, error)
	
	// Company-wide operations
	GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.VehicleHistory, error)
	GetUpcomingMaintenance(ctx context.Context, companyID string, days int) ([]*models.VehicleHistory, error)
	GetOverdueMaintenance(ctx context.Context, companyID string) ([]*models.VehicleHistory, error)
	GetMaintenanceByProvider(ctx context.Context, companyID, serviceProvider string, pagination Pagination) ([]*models.VehicleHistory, error)
	
	// Cost and analytics operations
	GetCostSummary(ctx context.Context, vehicleID string, startDate, endDate time.Time) (*CostSummary, error)
	GetCompanyCostSummary(ctx context.Context, companyID string, startDate, endDate time.Time) (*CostSummary, error)
	GetMaintenanceTrends(ctx context.Context, vehicleID string, months int) ([]*MaintenanceTrend, error)
	
	// Date range operations
	GetByDateRange(ctx context.Context, vehicleID string, startDate, endDate time.Time, pagination Pagination) ([]*models.VehicleHistory, error)
	GetByMileageRange(ctx context.Context, vehicleID string, minMileage, maxMileage int, pagination Pagination) ([]*models.VehicleHistory, error)
	
	// Search and filtering
	SearchByTitle(ctx context.Context, vehicleID, searchTerm string, pagination Pagination) ([]*models.VehicleHistory, error)
	SearchByDescription(ctx context.Context, vehicleID, searchTerm string, pagination Pagination) ([]*models.VehicleHistory, error)
	GetByCostRange(ctx context.Context, vehicleID string, minCost, maxCost float64, pagination Pagination) ([]*models.VehicleHistory, error)
	
	// Document operations
	GetWithDocuments(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.VehicleHistory, error)
	UpdateDocuments(ctx context.Context, historyID string, documents []map[string]interface{}) error
	
	// Maintenance scheduling
	GetNextMaintenanceDue(ctx context.Context, vehicleID string) (*models.VehicleHistory, error)
	UpdateMaintenanceSchedule(ctx context.Context, historyID string, nextDueDate time.Time) error
	GetMaintenanceSchedule(ctx context.Context, vehicleID string) ([]*models.VehicleHistory, error)
}

// CostSummary represents cost analytics data
type CostSummary struct {
	TotalCost     float64 `json:"total_cost"`
	MaintenanceCost float64 `json:"maintenance_cost"`
	RepairCost    float64 `json:"repair_cost"`
	InspectionCost float64 `json:"inspection_cost"`
	OtherCost     float64 `json:"other_cost"`
	EventCount    int     `json:"event_count"`
	AverageCost   float64 `json:"average_cost"`
	Currency      string  `json:"currency"`
}

// MaintenanceTrend represents maintenance trend data
type MaintenanceTrend struct {
	Month       string  `json:"month"`
	Year        int     `json:"year"`
	EventCount  int     `json:"event_count"`
	TotalCost   float64 `json:"total_cost"`
	AverageCost float64 `json:"average_cost"`
}

// VehicleHistoryRepositoryImpl implements VehicleHistoryRepository
type VehicleHistoryRepositoryImpl struct {
	BaseRepository[models.VehicleHistory]
}

// NewVehicleHistoryRepository creates a new VehicleHistoryRepository
func NewVehicleHistoryRepository(db *gorm.DB) VehicleHistoryRepository {
	return &VehicleHistoryRepositoryImpl{
		BaseRepository: BaseRepository[models.VehicleHistory]{
			db: db,
		},
	}
}

// GetByVehicle retrieves vehicle history for a specific vehicle
func (r *VehicleHistoryRepositoryImpl) GetByVehicle(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ?", vehicleID).
		Preload("Creator").
		Order("created_at DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetByVehicleAndType retrieves vehicle history filtered by event type
func (r *VehicleHistoryRepositoryImpl) GetByVehicleAndType(ctx context.Context, vehicleID, eventType string, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND event_type = ?", vehicleID, eventType).
		Preload("Creator").
		Order("created_at DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetByVehicleAndCategory retrieves vehicle history filtered by event category
func (r *VehicleHistoryRepositoryImpl) GetByVehicleAndCategory(ctx context.Context, vehicleID, eventCategory string, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND event_category = ?", vehicleID, eventCategory).
		Preload("Creator").
		Order("created_at DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetMaintenanceHistory retrieves maintenance history for a vehicle
func (r *VehicleHistoryRepositoryImpl) GetMaintenanceHistory(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.VehicleHistory, error) {
	return r.GetByVehicleAndType(ctx, vehicleID, models.EventTypeMaintenance, pagination)
}

// GetRepairHistory retrieves repair history for a vehicle
func (r *VehicleHistoryRepositoryImpl) GetRepairHistory(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.VehicleHistory, error) {
	return r.GetByVehicleAndType(ctx, vehicleID, models.EventTypeRepair, pagination)
}

// GetInspectionHistory retrieves inspection history for a vehicle
func (r *VehicleHistoryRepositoryImpl) GetInspectionHistory(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.VehicleHistory, error) {
	return r.GetByVehicleAndType(ctx, vehicleID, models.EventTypeInspection, pagination)
}

// GetByCompany retrieves all vehicle history for a company
func (r *VehicleHistoryRepositoryImpl) GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Preload("Vehicle").
		Preload("Creator").
		Order("created_at DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetUpcomingMaintenance retrieves upcoming maintenance for a company
func (r *VehicleHistoryRepositoryImpl) GetUpcomingMaintenance(ctx context.Context, companyID string, days int) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	futureDate := time.Now().AddDate(0, 0, days)
	
	query := r.db.WithContext(ctx).
		Where("company_id = ? AND next_service_due IS NOT NULL AND next_service_due <= ? AND next_service_due > ?", 
			companyID, futureDate, time.Now()).
		Preload("Vehicle").
		Preload("Creator").
		Order("next_service_due ASC")
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetOverdueMaintenance retrieves overdue maintenance for a company
func (r *VehicleHistoryRepositoryImpl) GetOverdueMaintenance(ctx context.Context, companyID string) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("company_id = ? AND next_service_due IS NOT NULL AND next_service_due < ?", 
			companyID, time.Now()).
		Preload("Vehicle").
		Preload("Creator").
		Order("next_service_due ASC")
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetMaintenanceByProvider retrieves maintenance history by service provider
func (r *VehicleHistoryRepositoryImpl) GetMaintenanceByProvider(ctx context.Context, companyID, serviceProvider string, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("company_id = ? AND service_provider = ?", companyID, serviceProvider).
		Preload("Vehicle").
		Preload("Creator").
		Order("created_at DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetCostSummary calculates cost summary for a vehicle
func (r *VehicleHistoryRepositoryImpl) GetCostSummary(ctx context.Context, vehicleID string, startDate, endDate time.Time) (*CostSummary, error) {
	var summary CostSummary
	
	// Get total cost
	if err := r.db.WithContext(ctx).
		Model(&models.VehicleHistory{}).
		Select("COALESCE(SUM(cost), 0) as total_cost, COUNT(*) as event_count").
		Where("vehicle_id = ? AND created_at BETWEEN ? AND ?", vehicleID, startDate, endDate).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	
	// Get maintenance cost
	if err := r.db.WithContext(ctx).
		Model(&models.VehicleHistory{}).
		Select("COALESCE(SUM(cost), 0) as maintenance_cost").
		Where("vehicle_id = ? AND event_type = ? AND created_at BETWEEN ? AND ?", 
			vehicleID, models.EventTypeMaintenance, startDate, endDate).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	
	// Get repair cost
	if err := r.db.WithContext(ctx).
		Model(&models.VehicleHistory{}).
		Select("COALESCE(SUM(cost), 0) as repair_cost").
		Where("vehicle_id = ? AND event_type = ? AND created_at BETWEEN ? AND ?", 
			vehicleID, models.EventTypeRepair, startDate, endDate).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	
	// Get inspection cost
	if err := r.db.WithContext(ctx).
		Model(&models.VehicleHistory{}).
		Select("COALESCE(SUM(cost), 0) as inspection_cost").
		Where("vehicle_id = ? AND event_type = ? AND created_at BETWEEN ? AND ?", 
			vehicleID, models.EventTypeInspection, startDate, endDate).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	
	// Calculate other costs and average
	summary.OtherCost = summary.TotalCost - summary.MaintenanceCost - summary.RepairCost - summary.InspectionCost
	if summary.EventCount > 0 {
		summary.AverageCost = summary.TotalCost / float64(summary.EventCount)
	}
	summary.Currency = "IDR"
	
	return &summary, nil
}

// GetCompanyCostSummary calculates cost summary for a company
func (r *VehicleHistoryRepositoryImpl) GetCompanyCostSummary(ctx context.Context, companyID string, startDate, endDate time.Time) (*CostSummary, error) {
	var summary CostSummary
	
	// Get total cost
	if err := r.db.WithContext(ctx).
		Model(&models.VehicleHistory{}).
		Select("COALESCE(SUM(cost), 0) as total_cost, COUNT(*) as event_count").
		Where("company_id = ? AND created_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	
	// Get maintenance cost
	if err := r.db.WithContext(ctx).
		Model(&models.VehicleHistory{}).
		Select("COALESCE(SUM(cost), 0) as maintenance_cost").
		Where("company_id = ? AND event_type = ? AND created_at BETWEEN ? AND ?", 
			companyID, models.EventTypeMaintenance, startDate, endDate).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	
	// Get repair cost
	if err := r.db.WithContext(ctx).
		Model(&models.VehicleHistory{}).
		Select("COALESCE(SUM(cost), 0) as repair_cost").
		Where("company_id = ? AND event_type = ? AND created_at BETWEEN ? AND ?", 
			companyID, models.EventTypeRepair, startDate, endDate).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	
	// Get inspection cost
	if err := r.db.WithContext(ctx).
		Model(&models.VehicleHistory{}).
		Select("COALESCE(SUM(cost), 0) as inspection_cost").
		Where("company_id = ? AND event_type = ? AND created_at BETWEEN ? AND ?", 
			companyID, models.EventTypeInspection, startDate, endDate).
		Scan(&summary).Error; err != nil {
		return nil, err
	}
	
	// Calculate other costs and average
	summary.OtherCost = summary.TotalCost - summary.MaintenanceCost - summary.RepairCost - summary.InspectionCost
	if summary.EventCount > 0 {
		summary.AverageCost = summary.TotalCost / float64(summary.EventCount)
	}
	summary.Currency = "IDR"
	
	return &summary, nil
}

// GetMaintenanceTrends retrieves maintenance trends for a vehicle
func (r *VehicleHistoryRepositoryImpl) GetMaintenanceTrends(ctx context.Context, vehicleID string, months int) ([]*MaintenanceTrend, error) {
	var trends []*MaintenanceTrend
	
	startDate := time.Now().AddDate(0, -months, 0)
	
	query := `
		SELECT 
			TO_CHAR(created_at, 'Mon') as month,
			EXTRACT(YEAR FROM created_at) as year,
			COUNT(*) as event_count,
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(AVG(cost), 0) as average_cost
		FROM vehicle_histories 
		WHERE vehicle_id = ? AND event_type = ? AND created_at >= ?
		GROUP BY TO_CHAR(created_at, 'Mon'), EXTRACT(YEAR FROM created_at)
		ORDER BY year DESC, 
			CASE TO_CHAR(created_at, 'Mon')
				WHEN 'Jan' THEN 1 WHEN 'Feb' THEN 2 WHEN 'Mar' THEN 3
				WHEN 'Apr' THEN 4 WHEN 'May' THEN 5 WHEN 'Jun' THEN 6
				WHEN 'Jul' THEN 7 WHEN 'Aug' THEN 8 WHEN 'Sep' THEN 9
				WHEN 'Oct' THEN 10 WHEN 'Nov' THEN 11 WHEN 'Dec' THEN 12
			END DESC
	`
	
	if err := r.db.WithContext(ctx).Raw(query, vehicleID, models.EventTypeMaintenance, startDate).Scan(&trends).Error; err != nil {
		return nil, err
	}
	
	return trends, nil
}

// GetByDateRange retrieves vehicle history within a date range
func (r *VehicleHistoryRepositoryImpl) GetByDateRange(ctx context.Context, vehicleID string, startDate, endDate time.Time, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND created_at BETWEEN ? AND ?", vehicleID, startDate, endDate).
		Preload("Creator").
		Order("created_at DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetByMileageRange retrieves vehicle history within a mileage range
func (r *VehicleHistoryRepositoryImpl) GetByMileageRange(ctx context.Context, vehicleID string, minMileage, maxMileage int, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND mileage_at_event BETWEEN ? AND ?", vehicleID, minMileage, maxMileage).
		Preload("Creator").
		Order("mileage_at_event DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// SearchByTitle searches vehicle history by title
func (r *VehicleHistoryRepositoryImpl) SearchByTitle(ctx context.Context, vehicleID, searchTerm string, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND title ILIKE ?", vehicleID, "%"+searchTerm+"%").
		Preload("Creator").
		Order("created_at DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// SearchByDescription searches vehicle history by description
func (r *VehicleHistoryRepositoryImpl) SearchByDescription(ctx context.Context, vehicleID, searchTerm string, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND description ILIKE ?", vehicleID, "%"+searchTerm+"%").
		Preload("Creator").
		Order("created_at DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetByCostRange retrieves vehicle history within a cost range
func (r *VehicleHistoryRepositoryImpl) GetByCostRange(ctx context.Context, vehicleID string, minCost, maxCost float64, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND cost BETWEEN ? AND ?", vehicleID, minCost, maxCost).
		Preload("Creator").
		Order("cost DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// GetWithDocuments retrieves vehicle history entries that have documents
func (r *VehicleHistoryRepositoryImpl) GetWithDocuments(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND documents IS NOT NULL AND documents != 'null' AND documents != '{}'", vehicleID).
		Preload("Creator").
		Order("created_at DESC")
	
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}

// UpdateDocuments updates documents for a history entry
func (r *VehicleHistoryRepositoryImpl) UpdateDocuments(ctx context.Context, historyID string, documents []map[string]interface{}) error {
	// Convert documents to JSON
	documentsJSON, err := json.Marshal(documents)
	if err != nil {
		return err
	}
	
	return r.db.WithContext(ctx).
		Model(&models.VehicleHistory{}).
		Where("id = ?", historyID).
		Update("documents", documentsJSON).Error
}

// GetNextMaintenanceDue retrieves the next maintenance due for a vehicle
func (r *VehicleHistoryRepositoryImpl) GetNextMaintenanceDue(ctx context.Context, vehicleID string) (*models.VehicleHistory, error) {
	var history models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND next_service_due IS NOT NULL AND next_service_due > ?", 
			vehicleID, time.Now()).
		Order("next_service_due ASC").
		First(&history)
	
	if err := query.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No upcoming maintenance
		}
		return nil, err
	}
	
	return &history, nil
}

// UpdateMaintenanceSchedule updates the maintenance schedule for a history entry
func (r *VehicleHistoryRepositoryImpl) UpdateMaintenanceSchedule(ctx context.Context, historyID string, nextDueDate time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.VehicleHistory{}).
		Where("id = ?", historyID).
		Update("next_service_due", nextDueDate).Error
}

// GetMaintenanceSchedule retrieves the maintenance schedule for a vehicle
func (r *VehicleHistoryRepositoryImpl) GetMaintenanceSchedule(ctx context.Context, vehicleID string) ([]*models.VehicleHistory, error) {
	var histories []*models.VehicleHistory
	
	query := r.db.WithContext(ctx).
		Where("vehicle_id = ? AND next_service_due IS NOT NULL", vehicleID).
		Order("next_service_due ASC")
	
	if err := query.Find(&histories).Error; err != nil {
		return nil, err
	}
	
	return histories, nil
}
