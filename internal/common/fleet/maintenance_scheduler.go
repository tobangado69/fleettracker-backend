package fleet

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// MaintenanceScheduler provides automated maintenance scheduling capabilities
type MaintenanceScheduler struct {
	db    *gorm.DB
	redis *redis.Client
}

// MaintenanceSchedule represents a maintenance schedule
type MaintenanceSchedule struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VehicleID       string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	CompanyID       string    `json:"company_id" gorm:"type:uuid;not null;index"`
	
	// Maintenance Details
	MaintenanceType string    `json:"maintenance_type" gorm:"type:varchar(50);not null"` // oil_change, tire_rotation, brake_service, etc.
	Description     string    `json:"description" gorm:"type:text"`
	Priority        string    `json:"priority" gorm:"type:varchar(20);not null"` // low, medium, high, critical
	
	// Scheduling
	ScheduledDate   time.Time `json:"scheduled_date" gorm:"not null"`
	EstimatedDuration int     `json:"estimated_duration" gorm:"not null"` // minutes
	AssignedDriver  *string   `json:"assigned_driver" gorm:"type:uuid;index"`
	AssignedMechanic *string  `json:"assigned_mechanic" gorm:"type:uuid;index"`
	
	// Triggers
	TriggerType     string    `json:"trigger_type" gorm:"type:varchar(30);not null"` // mileage, time, condition, manual
	TriggerValue    float64   `json:"trigger_value" gorm:"not null"` // km, days, or condition score
	CurrentValue    float64   `json:"current_value" gorm:"not null"` // current odometer, days since last, etc.
	
	// Status
	Status          string    `json:"status" gorm:"type:varchar(20);not null;default:'scheduled'"` // scheduled, in_progress, completed, cancelled
	CompletedAt     *time.Time `json:"completed_at"`
	CompletionNotes string    `json:"completion_notes" gorm:"type:text"`
	
	// Cost and Parts
	EstimatedCost   float64   `json:"estimated_cost" gorm:"not null"` // IDR
	ActualCost      *float64  `json:"actual_cost"`
	PartsUsed       []MaintenancePart `json:"parts_used" gorm:"type:jsonb"`
	
	// Metadata
	CreatedAt       time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"not null"`
	CreatedBy       string    `json:"created_by" gorm:"type:uuid;not null"`
}

// MaintenancePart represents a part used in maintenance
type MaintenancePart struct {
	PartID       string  `json:"part_id"`
	PartName     string  `json:"part_name"`
	PartNumber   string  `json:"part_number"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	TotalPrice   float64 `json:"total_price"`
	Supplier     string  `json:"supplier"`
	Warranty     int     `json:"warranty"` // months
}

// MaintenanceRule represents a maintenance rule
type MaintenanceRule struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID       string    `json:"company_id" gorm:"type:uuid;not null;index"`
	VehicleType     string    `json:"vehicle_type" gorm:"type:varchar(50)"` // empty means all types
	Make            string    `json:"make" gorm:"type:varchar(100)"` // empty means all makes
	Model           string    `json:"model" gorm:"type:varchar(100)"` // empty means all models
	
	// Rule Definition
	MaintenanceType string    `json:"maintenance_type" gorm:"type:varchar(50);not null"`
	Description     string    `json:"description" gorm:"type:text"`
	Priority        string    `json:"priority" gorm:"type:varchar(20);not null"`
	
	// Triggers
	TriggerType     string    `json:"trigger_type" gorm:"type:varchar(30);not null"`
	TriggerValue    float64   `json:"trigger_value" gorm:"not null"`
	BufferValue     float64   `json:"buffer_value" gorm:"not null"` // advance warning
	
	// Scheduling
	EstimatedDuration int     `json:"estimated_duration" gorm:"not null"`
	EstimatedCost   float64   `json:"estimated_cost" gorm:"not null"`
	
	// Status
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"not null"`
}

// MaintenanceAlert represents a maintenance alert
type MaintenanceAlert struct {
	ID              string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	VehicleID       string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	CompanyID       string    `json:"company_id" gorm:"type:uuid;not null;index"`
	ScheduleID      *string   `json:"schedule_id" gorm:"type:uuid;index"`
	
	// Alert Details
	AlertType       string    `json:"alert_type" gorm:"type:varchar(30);not null"` // due, overdue, critical, reminder
	MaintenanceType string    `json:"maintenance_type" gorm:"type:varchar(50);not null"`
	Message         string    `json:"message" gorm:"type:text;not null"`
	Severity        string    `json:"severity" gorm:"type:varchar(20);not null"` // low, medium, high, critical
	
	// Timing
	DueDate         time.Time `json:"due_date" gorm:"not null"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null"`
	IsRead          bool      `json:"is_read" gorm:"default:false"`
	IsResolved      bool      `json:"is_resolved" gorm:"default:false"`
}

// MaintenanceAnalytics represents maintenance analytics
type MaintenanceAnalytics struct {
	Period              string    `json:"period"`
	TotalScheduled      int       `json:"total_scheduled"`
	TotalCompleted      int       `json:"total_completed"`
	TotalOverdue        int       `json:"total_overdue"`
	TotalCost           float64   `json:"total_cost"`
	AverageCost         float64   `json:"average_cost"`
	AverageDuration     float64   `json:"average_duration"`
	CompletionRate      float64   `json:"completion_rate"`
	OverdueRate         float64   `json:"overdue_rate"`
	CostTrend           []CostTrendPoint `json:"cost_trend"`
	TypeBreakdown       []MaintenanceTypeStats `json:"type_breakdown"`
	VehicleBreakdown    []VehicleMaintenanceStats `json:"vehicle_breakdown"`
}

// CostTrendPoint represents a point in cost trend
type CostTrendPoint struct {
	Date        time.Time `json:"date"`
	TotalCost   float64   `json:"total_cost"`
	Count       int       `json:"count"`
	AverageCost float64   `json:"average_cost"`
}

// MaintenanceTypeStats represents statistics by maintenance type
type MaintenanceTypeStats struct {
	Type            string  `json:"type"`
	Count           int     `json:"count"`
	TotalCost       float64 `json:"total_cost"`
	AverageCost     float64 `json:"average_cost"`
	AverageDuration float64 `json:"average_duration"`
}

// VehicleMaintenanceStats represents maintenance statistics by vehicle
type VehicleMaintenanceStats struct {
	VehicleID       string  `json:"vehicle_id"`
	LicensePlate    string  `json:"license_plate"`
	Make            string  `json:"make"`
	Model           string  `json:"model"`
	MaintenanceCount int    `json:"maintenance_count"`
	TotalCost       float64 `json:"total_cost"`
	LastMaintenance time.Time `json:"last_maintenance"`
	NextMaintenance time.Time `json:"next_maintenance"`
}

// NewMaintenanceScheduler creates a new maintenance scheduler
func NewMaintenanceScheduler(db *gorm.DB, redis *redis.Client) *MaintenanceScheduler {
	return &MaintenanceScheduler{
		db:    db,
		redis: redis,
	}
}

// CreateMaintenanceRule creates a new maintenance rule
func (ms *MaintenanceScheduler) CreateMaintenanceRule(ctx context.Context, rule *MaintenanceRule) error {
	// Validate rule
	if err := ms.validateMaintenanceRule(rule); err != nil {
		return fmt.Errorf("maintenance rule validation failed: %w", err)
	}

	// Save to database
	if err := ms.db.Create(rule).Error; err != nil {
		return fmt.Errorf("failed to create maintenance rule: %w", err)
	}

	// Apply rule to existing vehicles
	go ms.applyRuleToVehicles(ctx, rule)

	return nil
}

// ScheduleMaintenance schedules a maintenance task
func (ms *MaintenanceScheduler) ScheduleMaintenance(ctx context.Context, schedule *MaintenanceSchedule) error {
	// Validate schedule
	if err := ms.validateMaintenanceSchedule(schedule); err != nil {
		return fmt.Errorf("maintenance schedule validation failed: %w", err)
	}

	// Check for conflicts
	if err := ms.checkSchedulingConflicts(ctx, schedule); err != nil {
		return fmt.Errorf("scheduling conflict detected: %w", err)
	}

	// Save to database
	if err := ms.db.Create(schedule).Error; err != nil {
		return fmt.Errorf("failed to schedule maintenance: %w", err)
	}

	// Create maintenance alert
	alert := &MaintenanceAlert{
		VehicleID:       schedule.VehicleID,
		CompanyID:       schedule.CompanyID,
		ScheduleID:      &schedule.ID,
		AlertType:       "scheduled",
		MaintenanceType: schedule.MaintenanceType,
		Message:         fmt.Sprintf("Maintenance scheduled: %s", schedule.Description),
		Severity:        schedule.Priority,
		DueDate:         schedule.ScheduledDate,
		CreatedAt:       time.Now(),
	}

	if err := ms.db.Create(alert).Error; err != nil {
		return fmt.Errorf("failed to create maintenance alert: %w", err)
	}

	// Invalidate cache
	ms.invalidateMaintenanceCache(ctx, schedule.CompanyID, schedule.VehicleID)

	return nil
}

// CompleteMaintenance marks a maintenance task as completed
func (ms *MaintenanceScheduler) CompleteMaintenance(ctx context.Context, scheduleID string, actualCost float64, completionNotes string, partsUsed []MaintenancePart) error {
	// Update schedule
	updates := map[string]interface{}{
		"status":           "completed",
		"completed_at":     time.Now(),
		"completion_notes": completionNotes,
		"actual_cost":      actualCost,
		"parts_used":       partsUsed,
		"updated_at":       time.Now(),
	}

	if err := ms.db.Model(&MaintenanceSchedule{}).Where("id = ?", scheduleID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to complete maintenance: %w", err)
	}

	// Mark related alerts as resolved
	if err := ms.db.Model(&MaintenanceAlert{}).Where("schedule_id = ?", scheduleID).Update("is_resolved", true).Error; err != nil {
		return fmt.Errorf("failed to resolve maintenance alerts: %w", err)
	}

	// Update vehicle maintenance history
	var schedule MaintenanceSchedule
	if err := ms.db.First(&schedule, "id = ?", scheduleID).Error; err != nil {
		return fmt.Errorf("failed to get maintenance schedule: %w", err)
	}

	// Create maintenance log entry
	maintenanceLog := models.MaintenanceLog{
		VehicleID:       schedule.VehicleID,
		MaintenanceType: schedule.MaintenanceType,
		Description:     schedule.Description,
		Cost:            actualCost,
		OdometerReading: schedule.CurrentValue,
	}

	if err := ms.db.Create(&maintenanceLog).Error; err != nil {
		return fmt.Errorf("failed to create maintenance log: %w", err)
	}

	// Invalidate cache
	ms.invalidateMaintenanceCache(ctx, schedule.CompanyID, schedule.VehicleID)

	return nil
}

// GetUpcomingMaintenance retrieves upcoming maintenance tasks
func (ms *MaintenanceScheduler) GetUpcomingMaintenance(ctx context.Context, companyID string, days int) ([]MaintenanceSchedule, error) {
	var schedules []MaintenanceSchedule
	
	endDate := time.Now().AddDate(0, 0, days)
	
	err := ms.db.Where("company_id = ? AND status = 'scheduled' AND scheduled_date <= ?", companyID, endDate).
		Order("scheduled_date ASC").
		Find(&schedules).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming maintenance: %w", err)
	}
	
	return schedules, nil
}

// GetMaintenanceAlerts retrieves maintenance alerts
func (ms *MaintenanceScheduler) GetMaintenanceAlerts(ctx context.Context, companyID string, severity string) ([]MaintenanceAlert, error) {
	var alerts []MaintenanceAlert
	
	query := ms.db.Where("company_id = ? AND is_resolved = false", companyID)
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	
	err := query.Order("due_date ASC").Find(&alerts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance alerts: %w", err)
	}
	
	return alerts, nil
}

// GetMaintenanceAnalytics retrieves maintenance analytics
func (ms *MaintenanceScheduler) GetMaintenanceAnalytics(ctx context.Context, companyID string, startDate, endDate time.Time) (*MaintenanceAnalytics, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("maintenance_analytics:%s:%s:%s", companyID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	cached, err := ms.getCachedAnalytics(ctx, cacheKey)
	if err == nil && cached != nil {
		return cached, nil
	}

	// Get total counts
	var totalScheduled, totalCompleted, totalOverdue int
	var totalCost float64

	err = ms.db.Model(&MaintenanceSchedule{}).
		Where("company_id = ? AND created_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Select("COUNT(*) as total_scheduled").
		Row().Scan(&totalScheduled)
	if err != nil {
		return nil, fmt.Errorf("failed to get total scheduled: %w", err)
	}

	err = ms.db.Model(&MaintenanceSchedule{}).
		Where("company_id = ? AND status = 'completed' AND completed_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Select("COUNT(*) as total_completed, COALESCE(SUM(actual_cost), 0) as total_cost").
		Row().Scan(&totalCompleted, &totalCost)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed maintenance: %w", err)
	}

	err = ms.db.Model(&MaintenanceSchedule{}).
		Where("company_id = ? AND status = 'scheduled' AND scheduled_date < ?", companyID, time.Now()).
		Select("COUNT(*) as total_overdue").
		Row().Scan(&totalOverdue)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue maintenance: %w", err)
	}

	// Calculate rates
	var completionRate, overdueRate float64
	if totalScheduled > 0 {
		completionRate = float64(totalCompleted) / float64(totalScheduled) * 100
		overdueRate = float64(totalOverdue) / float64(totalScheduled) * 100
	}

	// Calculate averages
	var averageCost, averageDuration float64
	if totalCompleted > 0 {
		averageCost = totalCost / float64(totalCompleted)
		
		err = ms.db.Model(&MaintenanceSchedule{}).
			Where("company_id = ? AND status = 'completed' AND completed_at BETWEEN ? AND ?", companyID, startDate, endDate).
			Select("AVG(estimated_duration)").
			Row().Scan(&averageDuration)
		if err != nil {
			averageDuration = 0
		}
	}

	// Get cost trend
	costTrend, err := ms.getCostTrend(ctx, companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost trend: %w", err)
	}

	// Get type breakdown
	typeBreakdown, err := ms.getTypeBreakdown(ctx, companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get type breakdown: %w", err)
	}

	// Get vehicle breakdown
	vehicleBreakdown, err := ms.getVehicleBreakdown(ctx, companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle breakdown: %w", err)
	}

	analytics := &MaintenanceAnalytics{
		Period:           fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		TotalScheduled:   totalScheduled,
		TotalCompleted:   totalCompleted,
		TotalOverdue:     totalOverdue,
		TotalCost:        totalCost,
		AverageCost:      averageCost,
		AverageDuration:  averageDuration,
		CompletionRate:   completionRate,
		OverdueRate:      overdueRate,
		CostTrend:        costTrend,
		TypeBreakdown:    typeBreakdown,
		VehicleBreakdown: vehicleBreakdown,
	}

	// Cache the result
	ms.cacheAnalytics(ctx, cacheKey, analytics, 1*time.Hour)

	return analytics, nil
}

// CheckMaintenanceTriggers checks all maintenance triggers and creates schedules
func (ms *MaintenanceScheduler) CheckMaintenanceTriggers(ctx context.Context, companyID string) error {
	// Get all active maintenance rules
	var rules []MaintenanceRule
	err := ms.db.Where("company_id = ? AND is_active = true", companyID).Find(&rules).Error
	if err != nil {
		return fmt.Errorf("failed to get maintenance rules: %w", err)
	}

	// Get all vehicles for the company
	var vehicles []models.Vehicle
	err = ms.db.Where("company_id = ?", companyID).Find(&vehicles).Error
	if err != nil {
		return fmt.Errorf("failed to get vehicles: %w", err)
	}

	// Check each rule against each vehicle
	for _, rule := range rules {
		for _, vehicle := range vehicles {
			// Check if rule applies to this vehicle
			if !ms.ruleAppliesToVehicle(rule, vehicle) {
				continue
			}

			// Check if maintenance is needed
			needed, currentValue, err := ms.isMaintenanceNeeded(ctx, rule, vehicle)
			if err != nil {
				continue
			}

			if needed {
				// Create maintenance schedule
				schedule := &MaintenanceSchedule{
					VehicleID:       vehicle.ID,
					CompanyID:       companyID,
					MaintenanceType: rule.MaintenanceType,
					Description:     rule.Description,
					Priority:        rule.Priority,
					TriggerType:     rule.TriggerType,
					TriggerValue:    rule.TriggerValue,
					CurrentValue:    currentValue,
					EstimatedDuration: rule.EstimatedDuration,
					EstimatedCost:   rule.EstimatedCost,
					Status:          "scheduled",
					ScheduledDate:   ms.calculateScheduledDate(rule, currentValue),
					CreatedAt:       time.Now(),
					CreatedBy:       "system",
				}

				// Check if schedule already exists
				var existingCount int64
				ms.db.Model(&MaintenanceSchedule{}).
					Where("vehicle_id = ? AND maintenance_type = ? AND status = 'scheduled'", vehicle.ID, rule.MaintenanceType).
					Count(&existingCount)

				if existingCount == 0 {
					ms.ScheduleMaintenance(ctx, schedule)
				}
			}
		}
	}

	return nil
}

// validateMaintenanceRule validates a maintenance rule
func (ms *MaintenanceScheduler) validateMaintenanceRule(rule *MaintenanceRule) error {
	if rule.CompanyID == "" {
		return fmt.Errorf("company ID is required")
	}
	if rule.MaintenanceType == "" {
		return fmt.Errorf("maintenance type is required")
	}
	if rule.TriggerType == "" {
		return fmt.Errorf("trigger type is required")
	}
	if rule.TriggerValue <= 0 {
		return fmt.Errorf("trigger value must be positive")
	}
	if rule.EstimatedDuration <= 0 {
		return fmt.Errorf("estimated duration must be positive")
	}
	if rule.EstimatedCost < 0 {
		return fmt.Errorf("estimated cost cannot be negative")
	}
	return nil
}

// validateMaintenanceSchedule validates a maintenance schedule
func (ms *MaintenanceScheduler) validateMaintenanceSchedule(schedule *MaintenanceSchedule) error {
	if schedule.VehicleID == "" {
		return fmt.Errorf("vehicle ID is required")
	}
	if schedule.CompanyID == "" {
		return fmt.Errorf("company ID is required")
	}
	if schedule.MaintenanceType == "" {
		return fmt.Errorf("maintenance type is required")
	}
	if schedule.ScheduledDate.IsZero() {
		return fmt.Errorf("scheduled date is required")
	}
	if schedule.EstimatedDuration <= 0 {
		return fmt.Errorf("estimated duration must be positive")
	}
	if schedule.EstimatedCost < 0 {
		return fmt.Errorf("estimated cost cannot be negative")
	}
	return nil
}

// checkSchedulingConflicts checks for scheduling conflicts
func (ms *MaintenanceScheduler) checkSchedulingConflicts(_ context.Context, schedule *MaintenanceSchedule) error {
	// Check for overlapping maintenance on the same vehicle
	var conflictCount int64
	err := ms.db.Model(&MaintenanceSchedule{}).
		Where("vehicle_id = ? AND status = 'scheduled' AND scheduled_date BETWEEN ? AND ?",
			schedule.VehicleID,
			schedule.ScheduledDate.Add(-time.Duration(schedule.EstimatedDuration)*time.Minute),
			schedule.ScheduledDate.Add(time.Duration(schedule.EstimatedDuration)*time.Minute)).
		Count(&conflictCount).Error

	if err != nil {
		return fmt.Errorf("failed to check scheduling conflicts: %w", err)
	}

	if conflictCount > 0 {
		return fmt.Errorf("scheduling conflict detected: overlapping maintenance on the same vehicle")
	}

	return nil
}

// applyRuleToVehicles applies a maintenance rule to existing vehicles
func (ms *MaintenanceScheduler) applyRuleToVehicles(_ context.Context, _ *MaintenanceRule) {
	// This would apply the rule to all existing vehicles that match the criteria
	// Implementation would depend on specific requirements
}

// ruleAppliesToVehicle checks if a rule applies to a specific vehicle
func (ms *MaintenanceScheduler) ruleAppliesToVehicle(rule MaintenanceRule, vehicle models.Vehicle) bool {
	if rule.VehicleType != "" && rule.VehicleType != vehicle.Make {
		return false
	}
	if rule.Make != "" && rule.Make != vehicle.Make {
		return false
	}
	if rule.Model != "" && rule.Model != vehicle.Model {
		return false
	}
	return true
}

// isMaintenanceNeeded checks if maintenance is needed for a vehicle based on a rule
func (ms *MaintenanceScheduler) isMaintenanceNeeded(_ context.Context, rule MaintenanceRule, vehicle models.Vehicle) (bool, float64, error) {
	switch rule.TriggerType {
	case "mileage":
		// Get current odometer reading
		var currentOdometer float64
		err := ms.db.Model(&models.MaintenanceLog{}).
			Where("vehicle_id = ?", vehicle.ID).
			Select("COALESCE(MAX(odometer_reading), 0)").
			Row().Scan(&currentOdometer)
		
		if err != nil {
			return false, 0, fmt.Errorf("failed to get current odometer: %w", err)
		}

		return currentOdometer >= rule.TriggerValue, currentOdometer, nil

	case "time":
		// Get last maintenance date
		var lastMaintenance time.Time
		err := ms.db.Model(&models.MaintenanceLog{}).
			Where("vehicle_id = ? AND maintenance_type = ?", vehicle.ID, rule.MaintenanceType).
			Select("COALESCE(MAX(created_at), '1900-01-01')").
			Row().Scan(&lastMaintenance)
		
		if err != nil {
			return false, 0, fmt.Errorf("failed to get last maintenance: %w", err)
		}

		daysSinceLast := time.Since(lastMaintenance).Hours() / 24
		return daysSinceLast >= rule.TriggerValue, daysSinceLast, nil

	default:
		return false, 0, fmt.Errorf("unsupported trigger type: %s", rule.TriggerType)
	}
}

// calculateScheduledDate calculates the scheduled date for maintenance
func (ms *MaintenanceScheduler) calculateScheduledDate(rule MaintenanceRule, currentValue float64) time.Time {
	// Schedule maintenance based on buffer value
	bufferValue := rule.BufferValue
	if bufferValue == 0 {
		bufferValue = rule.TriggerValue * 0.1 // 10% buffer by default
	}

	switch rule.TriggerType {
	case "mileage":
		// Schedule when current value + buffer reaches trigger value
		if currentValue+bufferValue >= rule.TriggerValue {
			return time.Now().AddDate(0, 0, 7) // Schedule 1 week from now
		}
	case "time":
		// Schedule when current value + buffer reaches trigger value
		if currentValue+bufferValue >= rule.TriggerValue {
			return time.Now().AddDate(0, 0, 3) // Schedule 3 days from now
		}
	}

	return time.Now().AddDate(0, 0, 1) // Default: schedule tomorrow
}

// Helper methods for analytics
func (ms *MaintenanceScheduler) getCostTrend(_ context.Context, companyID string, startDate, endDate time.Time) ([]CostTrendPoint, error) {
	var trend []CostTrendPoint
	
	rows, err := ms.db.Model(&MaintenanceSchedule{}).
		Select("DATE(completed_at) as date, SUM(actual_cost) as total_cost, COUNT(*) as count, AVG(actual_cost) as average_cost").
		Where("company_id = ? AND status = 'completed' AND completed_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("DATE(completed_at)").
		Order("date ASC").
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get cost trend: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var point CostTrendPoint
		var dateStr string
		
		err := rows.Scan(&dateStr, &point.TotalCost, &point.Count, &point.AverageCost)
		if err != nil {
			continue
		}
		
		point.Date, _ = time.Parse("2006-01-02", dateStr)
		trend = append(trend, point)
	}
	
	return trend, nil
}

func (ms *MaintenanceScheduler) getTypeBreakdown(_ context.Context, companyID string, startDate, endDate time.Time) ([]MaintenanceTypeStats, error) {
	var breakdown []MaintenanceTypeStats
	
	rows, err := ms.db.Model(&MaintenanceSchedule{}).
		Select("maintenance_type, COUNT(*) as count, SUM(actual_cost) as total_cost, AVG(actual_cost) as average_cost, AVG(estimated_duration) as average_duration").
		Where("company_id = ? AND status = 'completed' AND completed_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("maintenance_type").
		Order("count DESC").
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get type breakdown: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var stats MaintenanceTypeStats
		err := rows.Scan(&stats.Type, &stats.Count, &stats.TotalCost, &stats.AverageCost, &stats.AverageDuration)
		if err != nil {
			continue
		}
		
		breakdown = append(breakdown, stats)
	}
	
	return breakdown, nil
}

func (ms *MaintenanceScheduler) getVehicleBreakdown(_ context.Context, companyID string, startDate, endDate time.Time) ([]VehicleMaintenanceStats, error) {
	var breakdown []VehicleMaintenanceStats
	
	rows, err := ms.db.Table("maintenance_schedules ms").
		Select("ms.vehicle_id, v.license_plate, v.make, v.model, COUNT(*) as maintenance_count, SUM(ms.actual_cost) as total_cost, MAX(ms.completed_at) as last_maintenance, MIN(CASE WHEN ms.status = 'scheduled' THEN ms.scheduled_date END) as next_maintenance").
		Joins("JOIN vehicles v ON ms.vehicle_id = v.id").
		Where("ms.company_id = ? AND ms.completed_at BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("ms.vehicle_id, v.license_plate, v.make, v.model").
		Order("maintenance_count DESC").
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle breakdown: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var stats VehicleMaintenanceStats
		var lastMaintenance, nextMaintenance *time.Time
		
		err := rows.Scan(&stats.VehicleID, &stats.LicensePlate, &stats.Make, &stats.Model, 
			&stats.MaintenanceCount, &stats.TotalCost, &lastMaintenance, &nextMaintenance)
		if err != nil {
			continue
		}
		
		if lastMaintenance != nil {
			stats.LastMaintenance = *lastMaintenance
		}
		if nextMaintenance != nil {
			stats.NextMaintenance = *nextMaintenance
		}
		
		breakdown = append(breakdown, stats)
	}
	
	return breakdown, nil
}

// Cache methods
func (ms *MaintenanceScheduler) getCachedAnalytics(_ context.Context, _ string) (*MaintenanceAnalytics, error) {
	// Implementation would use Redis to get cached analytics
	return nil, fmt.Errorf("cache miss")
}

func (ms *MaintenanceScheduler) cacheAnalytics(_ context.Context, _ string, _ *MaintenanceAnalytics, _ time.Duration) error {
	// Implementation would use Redis to cache analytics
	return nil
}

func (ms *MaintenanceScheduler) invalidateMaintenanceCache(_ context.Context, _, _ string) error {
	// Implementation would invalidate relevant cache entries
	return nil
}
