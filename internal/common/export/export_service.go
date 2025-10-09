package export

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// ExportService provides data export functionality with caching
type ExportService struct {
	db     *gorm.DB
	cache  *ExportCacheService
}

// ExportRequest represents a request to export data
type ExportRequest struct {
	ExportType string                 `json:"export_type" binding:"required"`
	Format     string                 `json:"format" binding:"required"`
	Filters    map[string]interface{} `json:"filters"`
	CompanyID  string                 `json:"company_id"`
	UserID     string                 `json:"user_id"`
}

// ExportResponse represents the response from an export operation
type ExportResponse struct {
	ExportID    string         `json:"export_id"`
	Data        interface{}    `json:"data"`
	Metadata    ExportMetadata `json:"metadata"`
	FromCache   bool           `json:"from_cache"`
	CacheHit    bool           `json:"cache_hit"`
	ProcessedAt time.Time      `json:"processed_at"`
}

// NewExportService creates a new export service
func NewExportService(db *gorm.DB, cache *ExportCacheService) *ExportService {
	return &ExportService{
		db:    db,
		cache: cache,
	}
}

// ExportData exports data with caching support
func (es *ExportService) ExportData(ctx context.Context, req *ExportRequest) (*ExportResponse, error) {
	// Check cache first
	cachedData, err := es.cache.GetExportCache(ctx, req.ExportType, req.Format, req.Filters, req.CompanyID, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check cache: %w", err)
	}
	
	if cachedData != nil {
		// Cache hit
		es.cache.RecordCacheHit(ctx)
		
		return &ExportResponse{
			ExportID:    fmt.Sprintf("cached_%d", time.Now().UnixNano()),
			Data:        cachedData.Data,
			Metadata:    cachedData.Metadata,
			FromCache:   true,
			CacheHit:    true,
			ProcessedAt: time.Now(),
		}, nil
	}
	
	// Cache miss - record it
	es.cache.RecordCacheMiss(ctx)
	
	// Generate new export data
	data, metadata, err := es.generateExportData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate export data: %w", err)
	}
	
	// Cache the result
	ttl := es.cache.GetTTLForExportType(req.ExportType)
	err = es.cache.SetExportCache(ctx, req.ExportType, req.Format, data, metadata, req.Filters, req.CompanyID, req.UserID, ttl)
	if err != nil {
		// Log error but don't fail the export
		fmt.Printf("Warning: failed to cache export data: %v\n", err)
	}
	
	return &ExportResponse{
		ExportID:    fmt.Sprintf("export_%d", time.Now().UnixNano()),
		Data:        data,
		Metadata:    metadata,
		FromCache:   false,
		CacheHit:    false,
		ProcessedAt: time.Now(),
	}, nil
}

// generateExportData generates export data based on the request
func (es *ExportService) generateExportData(ctx context.Context, req *ExportRequest) (interface{}, ExportMetadata, error) {
	switch req.ExportType {
	case "vehicles":
		return es.exportVehicles(ctx, req)
	case "drivers":
		return es.exportDrivers(ctx, req)
	case "trips":
		return es.exportTrips(ctx, req)
	case "gps_tracks":
		return es.exportGPSTracks(ctx, req)
	case "reports":
		return es.exportReports(ctx, req)
	default:
		return nil, ExportMetadata{}, fmt.Errorf("unsupported export type: %s", req.ExportType)
	}
}

// exportVehicles exports vehicle data
func (es *ExportService) exportVehicles(_ context.Context, req *ExportRequest) (interface{}, ExportMetadata, error) {
	startTime := time.Now()
	_ = startTime // Use the variable to avoid unused variable error
	
	// Build query
	query := es.db.Model(&models.Vehicle{}).Where("company_id = ?", req.CompanyID)
	
	// Apply filters
	if req.Filters != nil {
		if status, ok := req.Filters["status"].(string); ok && status != "" {
			query = query.Where("status = ?", status)
		}
		if make, ok := req.Filters["make"].(string); ok && make != "" {
			query = query.Where("make = ?", make)
		}
		if year, ok := req.Filters["year"].(float64); ok && year > 0 {
			query = query.Where("year = ?", int(year))
		}
	}
	
	// Get vehicles
	var vehicles []models.Vehicle
	err := query.Find(&vehicles).Error
	if err != nil {
		return nil, ExportMetadata{}, fmt.Errorf("failed to get vehicles: %w", err)
	}
	
	// Format data based on requested format
	var data interface{}
	var fileSize int64
	
	switch req.Format {
	case "json":
		data = vehicles
		jsonData, _ := json.Marshal(vehicles)
		fileSize = int64(len(jsonData))
	case "csv":
		csvData, err := es.convertToCSV(vehicles)
		if err != nil {
			return nil, ExportMetadata{}, fmt.Errorf("failed to convert to CSV: %w", err)
		}
		data = csvData
		fileSize = int64(len(csvData))
	default:
		return nil, ExportMetadata{}, fmt.Errorf("unsupported format: %s", req.Format)
	}
	
	metadata := ExportMetadata{
		RecordCount: int64(len(vehicles)),
		FileSize:    fileSize,
		ExportTime:  time.Now(),
		Format:      req.Format,
		Filters:     req.Filters,
	}
	
	return data, metadata, nil
}

// exportDrivers exports driver data
func (es *ExportService) exportDrivers(_ context.Context, req *ExportRequest) (interface{}, ExportMetadata, error) {
	startTime := time.Now()
	_ = startTime // Use the variable to avoid unused variable error
	
	// Build query
	query := es.db.Model(&models.Driver{}).Where("company_id = ?", req.CompanyID)
	
	// Apply filters
	if req.Filters != nil {
		if status, ok := req.Filters["status"].(string); ok && status != "" {
			query = query.Where("status = ?", status)
		}
		if gender, ok := req.Filters["gender"].(string); ok && gender != "" {
			query = query.Where("gender = ?", gender)
		}
	}
	
	// Get drivers
	var drivers []models.Driver
	err := query.Find(&drivers).Error
	if err != nil {
		return nil, ExportMetadata{}, fmt.Errorf("failed to get drivers: %w", err)
	}
	
	// Format data based on requested format
	var data interface{}
	var fileSize int64
	
	switch req.Format {
	case "json":
		data = drivers
		jsonData, _ := json.Marshal(drivers)
		fileSize = int64(len(jsonData))
	case "csv":
		csvData, err := es.convertToCSV(drivers)
		if err != nil {
			return nil, ExportMetadata{}, fmt.Errorf("failed to convert to CSV: %w", err)
		}
		data = csvData
		fileSize = int64(len(csvData))
	default:
		return nil, ExportMetadata{}, fmt.Errorf("unsupported format: %s", req.Format)
	}
	
	metadata := ExportMetadata{
		RecordCount: int64(len(drivers)),
		FileSize:    fileSize,
		ExportTime:  time.Now(),
		Format:      req.Format,
		Filters:     req.Filters,
	}
	
	return data, metadata, nil
}

// exportTrips exports trip data
func (es *ExportService) exportTrips(_ context.Context, req *ExportRequest) (interface{}, ExportMetadata, error) {
	startTime := time.Now()
	_ = startTime // Use the variable to avoid unused variable error
	
	// Build query
	query := es.db.Model(&models.Trip{}).Where("company_id = ?", req.CompanyID)
	
	// Apply filters
	if req.Filters != nil {
		if startDate, ok := req.Filters["start_date"].(string); ok && startDate != "" {
			query = query.Where("start_time >= ?", startDate)
		}
		if endDate, ok := req.Filters["end_date"].(string); ok && endDate != "" {
			query = query.Where("start_time <= ?", endDate)
		}
		if status, ok := req.Filters["status"].(string); ok && status != "" {
			query = query.Where("status = ?", status)
		}
		if vehicleID, ok := req.Filters["vehicle_id"].(string); ok && vehicleID != "" {
			query = query.Where("vehicle_id = ?", vehicleID)
		}
		if driverID, ok := req.Filters["driver_id"].(string); ok && driverID != "" {
			query = query.Where("driver_id = ?", driverID)
		}
	}
	
	// Get trips
	var trips []models.Trip
	err := query.Find(&trips).Error
	if err != nil {
		return nil, ExportMetadata{}, fmt.Errorf("failed to get trips: %w", err)
	}
	
	// Format data based on requested format
	var data interface{}
	var fileSize int64
	
	switch req.Format {
	case "json":
		data = trips
		jsonData, _ := json.Marshal(trips)
		fileSize = int64(len(jsonData))
	case "csv":
		csvData, err := es.convertToCSV(trips)
		if err != nil {
			return nil, ExportMetadata{}, fmt.Errorf("failed to convert to CSV: %w", err)
		}
		data = csvData
		fileSize = int64(len(csvData))
	default:
		return nil, ExportMetadata{}, fmt.Errorf("unsupported format: %s", req.Format)
	}
	
	metadata := ExportMetadata{
		RecordCount: int64(len(trips)),
		FileSize:    fileSize,
		ExportTime:  time.Now(),
		Format:      req.Format,
		Filters:     req.Filters,
	}
	
	return data, metadata, nil
}

// exportGPSTracks exports GPS track data
func (es *ExportService) exportGPSTracks(_ context.Context, req *ExportRequest) (interface{}, ExportMetadata, error) {
	startTime := time.Now()
	_ = startTime // Use the variable to avoid unused variable error
	
	// Build query
	query := es.db.Model(&models.GPSTrack{})
	
	// Apply filters
	if req.Filters != nil {
		if startDate, ok := req.Filters["start_date"].(string); ok && startDate != "" {
			query = query.Where("timestamp >= ?", startDate)
		}
		if endDate, ok := req.Filters["end_date"].(string); ok && endDate != "" {
			query = query.Where("timestamp <= ?", endDate)
		}
		if vehicleID, ok := req.Filters["vehicle_id"].(string); ok && vehicleID != "" {
			query = query.Where("vehicle_id = ?", vehicleID)
		}
		if driverID, ok := req.Filters["driver_id"].(string); ok && driverID != "" {
			query = query.Where("driver_id = ?", driverID)
		}
	}
	
	// Limit GPS tracks to prevent memory issues
	limit := 10000
	if req.Filters != nil {
		if l, ok := req.Filters["limit"].(float64); ok && l > 0 {
			limit = int(l)
		}
	}
	query = query.Limit(limit)
	
	// Get GPS tracks
	var gpsTracks []models.GPSTrack
	err := query.Find(&gpsTracks).Error
	if err != nil {
		return nil, ExportMetadata{}, fmt.Errorf("failed to get GPS tracks: %w", err)
	}
	
	// Format data based on requested format
	var data interface{}
	var fileSize int64
	
	switch req.Format {
	case "json":
		data = gpsTracks
		jsonData, _ := json.Marshal(gpsTracks)
		fileSize = int64(len(jsonData))
	case "csv":
		csvData, err := es.convertToCSV(gpsTracks)
		if err != nil {
			return nil, ExportMetadata{}, fmt.Errorf("failed to convert to CSV: %w", err)
		}
		data = csvData
		fileSize = int64(len(csvData))
	default:
		return nil, ExportMetadata{}, fmt.Errorf("unsupported format: %s", req.Format)
	}
	
	metadata := ExportMetadata{
		RecordCount: int64(len(gpsTracks)),
		FileSize:    fileSize,
		ExportTime:  time.Now(),
		Format:      req.Format,
		Filters:     req.Filters,
	}
	
	return data, metadata, nil
}

// exportReports exports report data
func (es *ExportService) exportReports(_ context.Context, req *ExportRequest) (interface{}, ExportMetadata, error) {
	startTime := time.Now()
	_ = startTime // Use the variable to avoid unused variable error
	
	// This would integrate with the analytics service to generate reports
	// For now, return a placeholder
	reportData := map[string]interface{}{
		"report_type": req.Filters["report_type"],
		"generated_at": time.Now(),
		"data": "Report data would be generated here",
	}
	
	// Format data based on requested format
	var data interface{}
	var fileSize int64
	
	switch req.Format {
	case "json":
		data = reportData
		jsonData, _ := json.Marshal(reportData)
		fileSize = int64(len(jsonData))
	case "csv":
		// Convert report to CSV format
		csvData := "Report Type,Generated At,Data\n"
		csvData += fmt.Sprintf("%v,%v,%v\n", 
			reportData["report_type"], 
			reportData["generated_at"], 
			reportData["data"])
		data = csvData
		fileSize = int64(len(csvData))
	default:
		return nil, ExportMetadata{}, fmt.Errorf("unsupported format: %s", req.Format)
	}
	
	metadata := ExportMetadata{
		RecordCount: 1,
		FileSize:    fileSize,
		ExportTime:  time.Now(),
		Format:      req.Format,
		Filters:     req.Filters,
	}
	
	return data, metadata, nil
}

// convertToCSV converts a slice of structs to CSV format
func (es *ExportService) convertToCSV(data interface{}) (string, error) {
	// This is a simplified CSV conversion
	// In a real implementation, you'd want to use reflection or a proper CSV library
	
	switch v := data.(type) {
	case []models.Vehicle:
		return es.vehiclesToCSV(v)
	case []models.Driver:
		return es.driversToCSV(v)
	case []models.Trip:
		return es.tripsToCSV(v)
	case []models.GPSTrack:
		return es.gpsTracksToCSV(v)
	default:
		return "", fmt.Errorf("unsupported data type for CSV conversion")
	}
}

// vehiclesToCSV converts vehicles to CSV format
func (es *ExportService) vehiclesToCSV(vehicles []models.Vehicle) (string, error) {
	var csvData string
	
	// Header
	csvData += "ID,License Plate,Make,Model,Year,Color,Status,Company ID,Created At\n"
	
	// Data rows
	for _, vehicle := range vehicles {
		csvData += fmt.Sprintf("%s,%s,%s,%s,%d,%s,%s,%s,%s\n",
			vehicle.ID,
			vehicle.LicensePlate,
			vehicle.Make,
			vehicle.Model,
			vehicle.Year,
			vehicle.Color,
			vehicle.Status,
			vehicle.CompanyID,
			vehicle.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}
	
	return csvData, nil
}

// driversToCSV converts drivers to CSV format
func (es *ExportService) driversToCSV(drivers []models.Driver) (string, error) {
	var csvData string
	
	// Header
	csvData += "ID,First Name,Last Name,Email,Phone,NIK,Status,Company ID,Created At\n"
	
	// Data rows
	for _, driver := range drivers {
		csvData += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			driver.ID,
			driver.FirstName,
			driver.LastName,
			driver.Email,
			driver.Phone,
			driver.NIK,
			driver.Status,
			driver.CompanyID,
			driver.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}
	
	return csvData, nil
}

// tripsToCSV converts trips to CSV format
func (es *ExportService) tripsToCSV(trips []models.Trip) (string, error) {
	var csvData string
	
	// Header
	csvData += "ID,Vehicle ID,Driver ID,Start Time,End Time,Status,Total Distance,Total Duration,Company ID,Created At\n"
	
	// Data rows
	for _, trip := range trips {
		csvData += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%.2f,%d,%s,%s\n",
			trip.ID,
			trip.VehicleID,
			*trip.DriverID,
			trip.StartTime.Format("2006-01-02 15:04:05"),
			trip.EndTime.Format("2006-01-02 15:04:05"),
			trip.Status,
			trip.TotalDistance,
			trip.TotalDuration,
			trip.CompanyID,
			trip.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}
	
	return csvData, nil
}

// gpsTracksToCSV converts GPS tracks to CSV format
func (es *ExportService) gpsTracksToCSV(gpsTracks []models.GPSTrack) (string, error) {
	var csvData string
	
	// Header
	csvData += "ID,Vehicle ID,Driver ID,Latitude,Longitude,Speed,Heading,Timestamp\n"
	
	// Data rows
	for _, track := range gpsTracks {
		csvData += fmt.Sprintf("%s,%s,%s,%.6f,%.6f,%.2f,%.2f,%s\n",
			track.ID,
			track.VehicleID,
			*track.DriverID,
			track.Latitude,
			track.Longitude,
			track.Speed,
			track.Heading,
			track.Timestamp.Format("2006-01-02 15:04:05"),
		)
	}
	
	return csvData, nil
}

// InvalidateCache invalidates cache for specific export parameters
func (es *ExportService) InvalidateCache(ctx context.Context, exportType, format string, filters map[string]interface{}, companyID, userID string) error {
	return es.cache.InvalidateExportCache(ctx, exportType, format, filters, companyID, userID)
}

// InvalidateCompanyCache invalidates all cache for a company
func (es *ExportService) InvalidateCompanyCache(ctx context.Context, companyID string) error {
	return es.cache.InvalidateCompanyExports(ctx, companyID)
}

// InvalidateUserCache invalidates all cache for a user
func (es *ExportService) InvalidateUserCache(ctx context.Context, userID string) error {
	return es.cache.InvalidateUserExports(ctx, userID)
}

// GetCacheStats returns cache statistics
func (es *ExportService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	return es.cache.GetCacheStats(ctx)
}

// GetCacheHitRate returns cache hit rate statistics
func (es *ExportService) GetCacheHitRate(ctx context.Context) (map[string]interface{}, error) {
	return es.cache.GetCacheHitRate(ctx)
}

// CleanupExpiredCache removes expired cache entries
func (es *ExportService) CleanupExpiredCache(ctx context.Context) error {
	return es.cache.CleanupExpiredCache(ctx)
}
