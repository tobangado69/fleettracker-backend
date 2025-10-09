package jobs

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/internal/common/export"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// EmailNotificationJob handles email notification jobs
type EmailNotificationJob struct {
	db *gorm.DB
}

// NewEmailNotificationJob creates a new email notification job handler
func NewEmailNotificationJob(db *gorm.DB) *EmailNotificationJob {
	return &EmailNotificationJob{db: db}
}

// GetJobType returns the job type
func (e *EmailNotificationJob) GetJobType() string {
	return "email_notification"
}

// Handle processes email notification jobs
func (e *EmailNotificationJob) Handle(ctx context.Context, job *Job) error {
	// Extract job data
	to, ok := job.Data["to"].(string)
	if !ok {
		return fmt.Errorf("missing 'to' field in job data")
	}

	subject, ok := job.Data["subject"].(string)
	if !ok {
		return fmt.Errorf("missing 'subject' field in job data")
	}

	body, ok := job.Data["body"].(string)
	if !ok {
		return fmt.Errorf("missing 'body' field in job data")
	}
	_ = body // Use the variable to avoid unused variable error

	// Simulate email sending (replace with actual email service)
	fmt.Printf("Sending email to %s: %s\n", to, subject)
	time.Sleep(2 * time.Second) // Simulate email sending time

	// Log email notification
	notification := models.AuditLog{
		UserID:    job.UserID,
		Action:    "email_notification_sent",
		Details:   models.JSON{"to": to, "subject": subject},
		IPAddress: "system",
	}

	if err := e.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to log email notification: %w", err)
	}

	return nil
}

// ReportGenerationJob handles report generation jobs
type ReportGenerationJob struct {
	db *gorm.DB
}

// NewReportGenerationJob creates a new report generation job handler
func NewReportGenerationJob(db *gorm.DB) *ReportGenerationJob {
	return &ReportGenerationJob{db: db}
}

// GetJobType returns the job type
func (r *ReportGenerationJob) GetJobType() string {
	return "report_generation"
}

// Handle processes report generation jobs
func (r *ReportGenerationJob) Handle(ctx context.Context, job *Job) error {
	// Extract job data
	reportType, ok := job.Data["report_type"].(string)
	if !ok {
		return fmt.Errorf("missing 'report_type' field in job data")
	}

	startDate, ok := job.Data["start_date"].(string)
	if !ok {
		return fmt.Errorf("missing 'start_date' field in job data")
	}

	endDate, ok := job.Data["end_date"].(string)
	if !ok {
		return fmt.Errorf("missing 'end_date' field in job data")
	}

	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("invalid start_date format: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("invalid end_date format: %w", err)
	}

	// Generate report based on type
	switch reportType {
	case "fleet_summary":
		return r.generateFleetSummaryReport(ctx, job, start, end)
	case "driver_performance":
		return r.generateDriverPerformanceReport(ctx, job, start, end)
	case "fuel_consumption":
		return r.generateFuelConsumptionReport(ctx, job, start, end)
	case "maintenance":
		return r.generateMaintenanceReport(ctx, job, start, end)
	default:
		return fmt.Errorf("unknown report type: %s", reportType)
	}
}

// generateFleetSummaryReport generates a fleet summary report
func (r *ReportGenerationJob) generateFleetSummaryReport(_ context.Context, job *Job, start, end time.Time) error {
	// Simulate report generation
	fmt.Printf("Generating fleet summary report for %s to %s\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
	time.Sleep(5 * time.Second) // Simulate report generation time

	// Get fleet statistics
	var totalVehicles int64
	var activeVehicles int64
	var totalTrips int64
	var totalDistance float64

	r.db.Model(&models.Vehicle{}).Where("company_id = ?", job.CompanyID).Count(&totalVehicles)
	r.db.Model(&models.Vehicle{}).Where("company_id = ? AND status = ?", job.CompanyID, "active").Count(&activeVehicles)
	
	r.db.Model(&models.Trip{}).Where("company_id = ? AND start_time >= ? AND start_time <= ?", 
		job.CompanyID, start, end).Count(&totalTrips)
	
	r.db.Model(&models.Trip{}).Where("company_id = ? AND start_time >= ? AND start_time <= ?", 
		job.CompanyID, start, end).Select("COALESCE(SUM(total_distance), 0)").Scan(&totalDistance)

	// Log report generation
	notification := models.AuditLog{
		UserID:    job.UserID,
		Action:    "report_generated",
		Details:   models.JSON{
			"type": "fleet_summary", 
			"vehicles": activeVehicles, 
			"trips": totalTrips, 
			"distance": totalDistance,
		},
		IPAddress: "system",
	}

	if err := r.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to log report generation: %w", err)
	}

	return nil
}

// generateDriverPerformanceReport generates a driver performance report
func (r *ReportGenerationJob) generateDriverPerformanceReport(_ context.Context, job *Job, start, end time.Time) error {
	// Simulate report generation
	fmt.Printf("Generating driver performance report for %s to %s\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
	time.Sleep(8 * time.Second) // Simulate report generation time

	// Get driver performance statistics
	var totalDrivers int64
	var activeDrivers int64
	var totalEvents int64

	r.db.Model(&models.Driver{}).Where("company_id = ?", job.CompanyID).Count(&totalDrivers)
	r.db.Model(&models.Driver{}).Where("company_id = ? AND status = ?", job.CompanyID, "active").Count(&activeDrivers)
	
	r.db.Model(&models.DriverEvent{}).Where("created_at >= ? AND created_at <= ?", start, end).Count(&totalEvents)

	// Log report generation
	notification := models.AuditLog{
		UserID:    job.UserID,
		Action:    "report_generated",
		Details:   models.JSON{
			"type": "driver_performance", 
			"drivers": activeDrivers, 
			"events": totalEvents,
		},
		IPAddress: "system",
	}

	if err := r.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to log report generation: %w", err)
	}

	return nil
}

// generateFuelConsumptionReport generates a fuel consumption report
func (r *ReportGenerationJob) generateFuelConsumptionReport(_ context.Context, job *Job, start, end time.Time) error {
	// Simulate report generation
	fmt.Printf("Generating fuel consumption report for %s to %s\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
	time.Sleep(6 * time.Second) // Simulate report generation time

	// Get fuel consumption statistics
	var totalFuelConsumed float64
	var totalTrips int64

	r.db.Model(&models.Trip{}).Where("company_id = ? AND start_time >= ? AND start_time <= ?", 
		job.CompanyID, start, end).Select("COALESCE(SUM(fuel_consumed), 0)").Scan(&totalFuelConsumed)
	
	r.db.Model(&models.Trip{}).Where("company_id = ? AND start_time >= ? AND start_time <= ?", 
		job.CompanyID, start, end).Count(&totalTrips)

	// Log report generation
	notification := models.AuditLog{
		UserID:    job.UserID,
		Action:    "report_generated",
		Details:   models.JSON{
			"type": "fuel_consumption", 
			"fuel": totalFuelConsumed, 
			"trips": totalTrips,
		},
		IPAddress: "system",
	}

	if err := r.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to log report generation: %w", err)
	}

	return nil
}

// generateMaintenanceReport generates a maintenance report
func (r *ReportGenerationJob) generateMaintenanceReport(_ context.Context, job *Job, start, end time.Time) error {
	// Simulate report generation
	fmt.Printf("Generating maintenance report for %s to %s\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
	time.Sleep(7 * time.Second) // Simulate report generation time

	// Get maintenance statistics
	var totalMaintenanceLogs int64
	var pendingMaintenance int64

	r.db.Model(&models.MaintenanceLog{}).Where("company_id = ? AND created_at >= ? AND created_at <= ?", 
		job.CompanyID, start, end).Count(&totalMaintenanceLogs)
	
	r.db.Model(&models.MaintenanceLog{}).Where("company_id = ? AND status = ? AND created_at >= ? AND created_at <= ?", 
		job.CompanyID, "pending", start, end).Count(&pendingMaintenance)

	// Log report generation
	notification := models.AuditLog{
		UserID:    job.UserID,
		Action:    "report_generated",
		Details:   models.JSON{
			"type": "maintenance", 
			"logs": totalMaintenanceLogs, 
			"pending": pendingMaintenance,
		},
		IPAddress: "system",
	}

	if err := r.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to log report generation: %w", err)
	}

	return nil
}

// DataExportJob handles data export jobs
type DataExportJob struct {
	db            *gorm.DB
	exportService *export.ExportService
}

// NewDataExportJob creates a new data export job handler
func NewDataExportJob(db *gorm.DB, exportService *export.ExportService) *DataExportJob {
	return &DataExportJob{
		db:            db,
		exportService: exportService,
	}
}

// GetJobType returns the job type
func (d *DataExportJob) GetJobType() string {
	return "data_export"
}

// Handle processes data export jobs
func (d *DataExportJob) Handle(ctx context.Context, job *Job) error {
	// Extract job data
	exportType, ok := job.Data["export_type"].(string)
	if !ok {
		return fmt.Errorf("missing 'export_type' field in job data")
	}

	format, ok := job.Data["format"].(string)
	if !ok {
		format = "csv" // Default format
	}

	// Extract filters
	filters := make(map[string]interface{})
	if filtersData, ok := job.Data["filters"].(map[string]interface{}); ok {
		filters = filtersData
	}

	// Create export request
	exportReq := &export.ExportRequest{
		ExportType: exportType,
		Format:     format,
		Filters:    filters,
		CompanyID:  job.CompanyID,
		UserID:     job.UserID,
	}

	// Use the export service with caching
	response, err := d.exportService.ExportData(ctx, exportReq)
	if err != nil {
		return fmt.Errorf("failed to export data: %w", err)
	}

	// Log the export
	notification := models.AuditLog{
		UserID:    job.UserID,
		Action:    "data_exported",
		Details:   models.JSON{
			"export_type": exportType,
			"format": format,
			"record_count": response.Metadata.RecordCount,
			"file_size": response.Metadata.FileSize,
			"from_cache": response.FromCache,
			"cache_hit": response.CacheHit,
		},
		IPAddress: "system",
	}

	if err := d.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to log data export: %w", err)
	}

	return nil
}


// MaintenanceReminderJob handles maintenance reminder jobs
type MaintenanceReminderJob struct {
	db *gorm.DB
}

// NewMaintenanceReminderJob creates a new maintenance reminder job handler
func NewMaintenanceReminderJob(db *gorm.DB) *MaintenanceReminderJob {
	return &MaintenanceReminderJob{db: db}
}

// GetJobType returns the job type
func (m *MaintenanceReminderJob) GetJobType() string {
	return "maintenance_reminder"
}

// Handle processes maintenance reminder jobs
func (m *MaintenanceReminderJob) Handle(ctx context.Context, job *Job) error {
	// Extract job data
	vehicleID, ok := job.Data["vehicle_id"].(string)
	if !ok {
		return fmt.Errorf("missing 'vehicle_id' field in job data")
	}

	maintenanceType, ok := job.Data["maintenance_type"].(string)
	if !ok {
		return fmt.Errorf("missing 'maintenance_type' field in job data")
	}

	// Get vehicle information
	var vehicle models.Vehicle
	if err := m.db.Where("id = ?", vehicleID).First(&vehicle).Error; err != nil {
		return fmt.Errorf("failed to get vehicle: %w", err)
	}

	// Create maintenance log
	maintenanceLog := models.MaintenanceLog{
		VehicleID:       vehicleID,
		MaintenanceType: maintenanceType,
		Description:     fmt.Sprintf("Scheduled maintenance reminder: %s", maintenanceType),
		Cost:            0, // Default cost
		OdometerReading: 0, // Default odometer reading
	}

	if err := m.db.Create(&maintenanceLog).Error; err != nil {
		return fmt.Errorf("failed to create maintenance log: %w", err)
	}

	// Log maintenance reminder
	notification := models.AuditLog{
		UserID:    job.UserID,
		Action:    "maintenance_reminder_created",
		Details:   models.JSON{
			"vehicle_id": vehicleID, 
			"license_plate": vehicle.LicensePlate, 
			"maintenance_type": maintenanceType,
		},
		IPAddress: "system",
	}

	if err := m.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to log maintenance reminder: %w", err)
	}

	return nil
}

// DataCleanupJob handles data cleanup jobs
type DataCleanupJob struct {
	db *gorm.DB
}

// NewDataCleanupJob creates a new data cleanup job handler
func NewDataCleanupJob(db *gorm.DB) *DataCleanupJob {
	return &DataCleanupJob{db: db}
}

// GetJobType returns the job type
func (d *DataCleanupJob) GetJobType() string {
	return "data_cleanup"
}

// Handle processes data cleanup jobs
func (d *DataCleanupJob) Handle(ctx context.Context, job *Job) error {
	// Extract job data
	cleanupType, ok := job.Data["cleanup_type"].(string)
	if !ok {
		return fmt.Errorf("missing 'cleanup_type' field in job data")
	}

	olderThanDays, ok := job.Data["older_than_days"].(float64)
	if !ok {
		olderThanDays = 90 // Default to 90 days
	}

	cutoffDate := time.Now().AddDate(0, 0, -int(olderThanDays))

	// Perform cleanup based on type
	switch cleanupType {
	case "gps_tracks":
		return d.cleanupGPSTracks(ctx, job, cutoffDate)
	case "audit_logs":
		return d.cleanupAuditLogs(ctx, job, cutoffDate)
	case "driver_events":
		return d.cleanupDriverEvents(ctx, job, cutoffDate)
	default:
		return fmt.Errorf("unknown cleanup type: %s", cleanupType)
	}
}

// cleanupGPSTracks cleans up old GPS track data
func (d *DataCleanupJob) cleanupGPSTracks(_ context.Context, job *Job, cutoffDate time.Time) error {
	fmt.Printf("Cleaning up GPS tracks older than %s\n", cutoffDate.Format("2006-01-02"))
	
	// Count records to be deleted
	var count int64
	d.db.Model(&models.GPSTrack{}).Where("timestamp < ?", cutoffDate).Count(&count)

	// Delete old GPS tracks
	result := d.db.Where("timestamp < ?", cutoffDate).Delete(&models.GPSTrack{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup GPS tracks: %w", result.Error)
	}

	// Log cleanup
	notification := models.AuditLog{
		UserID:    job.UserID,
		Action:    "data_cleanup",
		Details:   models.JSON{
			"type": "gps_tracks", 
			"records_deleted": result.RowsAffected,
		},
		IPAddress: "system",
	}

	if err := d.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to log data cleanup: %w", err)
	}

	return nil
}

// cleanupAuditLogs cleans up old audit log data
func (d *DataCleanupJob) cleanupAuditLogs(_ context.Context, job *Job, cutoffDate time.Time) error {
	fmt.Printf("Cleaning up audit logs older than %s\n", cutoffDate.Format("2006-01-02"))
	
	// Count records to be deleted
	var count int64
	d.db.Model(&models.AuditLog{}).Where("created_at < ?", cutoffDate).Count(&count)

	// Delete old audit logs
	result := d.db.Where("created_at < ?", cutoffDate).Delete(&models.AuditLog{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup audit logs: %w", result.Error)
	}

	// Log cleanup
	notification := models.AuditLog{
		UserID:    job.UserID,
		Action:    "data_cleanup",
		Details:   models.JSON{
			"type": "audit_logs", 
			"records_deleted": result.RowsAffected,
		},
		IPAddress: "system",
	}

	if err := d.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to log data cleanup: %w", err)
	}

	return nil
}

// cleanupDriverEvents cleans up old driver event data
func (d *DataCleanupJob) cleanupDriverEvents(_ context.Context, job *Job, cutoffDate time.Time) error {
	fmt.Printf("Cleaning up driver events older than %s\n", cutoffDate.Format("2006-01-02"))
	
	// Count records to be deleted
	var count int64
	d.db.Model(&models.DriverEvent{}).Where("created_at < ?", cutoffDate).Count(&count)

	// Delete old driver events
	result := d.db.Where("created_at < ?", cutoffDate).Delete(&models.DriverEvent{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup driver events: %w", result.Error)
	}

	// Log cleanup
	notification := models.AuditLog{
		UserID:    job.UserID,
		Action:    "data_cleanup",
		Details:   models.JSON{
			"type": "driver_events", 
			"records_deleted": result.RowsAffected,
		},
		IPAddress: "system",
	}

	if err := d.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to log data cleanup: %w", err)
	}

	return nil
}
