package geofencing

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// GeofenceManager provides advanced geofencing capabilities
type GeofenceManager struct {
	db    *gorm.DB
	redis *redis.Client
}

// Geofence represents a geofence zone
type Geofence struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompanyID   string    `json:"company_id" gorm:"type:uuid;not null;index"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Type        string    `json:"type" gorm:"type:varchar(50);not null"` // polygon, circle, rectangle, route
	
	// Geometry
	Coordinates []Coordinate `json:"coordinates" gorm:"type:jsonb"`
	CenterLat   float64     `json:"center_lat" gorm:"not null"`
	CenterLng   float64     `json:"center_lng" gorm:"not null"`
	Radius      float64     `json:"radius" gorm:"default:0"` // meters, for circular geofences
	
	// Configuration
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Priority    int       `json:"priority" gorm:"default:1"` // 1-10, higher is more important
	Color       string    `json:"color" gorm:"type:varchar(7);default:'#FF0000'"` // hex color
	
	// Alerts and Actions
	AlertOnEntry    bool   `json:"alert_on_entry" gorm:"default:true"`
	AlertOnExit     bool   `json:"alert_on_exit" gorm:"default:true"`
	AlertOnDwell    bool   `json:"alert_on_dwell" gorm:"default:false"`
	DwellTime       int    `json:"dwell_time" gorm:"default:0"` // minutes
	SpeedLimit      float64 `json:"speed_limit" gorm:"default:0"` // km/h, 0 means no limit
	AlertOnSpeed    bool   `json:"alert_on_speed" gorm:"default:false"`
	
	// Time Restrictions
	TimeRestrictions []TimeRestriction `json:"time_restrictions" gorm:"type:jsonb"`
	
	// Vehicle and Driver Restrictions
	AllowedVehicles []string `json:"allowed_vehicles" gorm:"type:jsonb"`
	AllowedDrivers  []string `json:"allowed_drivers" gorm:"type:jsonb"`
	RestrictedVehicles []string `json:"restricted_vehicles" gorm:"type:jsonb"`
	RestrictedDrivers  []string `json:"restricted_drivers" gorm:"type:jsonb"`
	
	// Metadata
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null"`
	CreatedBy   string    `json:"created_by" gorm:"type:uuid;not null"`
}

// Coordinate represents a geographical coordinate
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// TimeRestriction represents time-based restrictions for geofences
type TimeRestriction struct {
	DayOfWeek int    `json:"day_of_week"` // 0=Sunday, 1=Monday, ..., 6=Saturday
	StartTime string `json:"start_time"`  // HH:MM format
	EndTime   string `json:"end_time"`    // HH:MM format
	IsActive  bool   `json:"is_active"`
}

// GeofenceEvent represents a geofence event
type GeofenceEvent struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GeofenceID  string    `json:"geofence_id" gorm:"type:uuid;not null;index"`
	VehicleID   string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	DriverID    string    `json:"driver_id" gorm:"type:uuid;index"`
	CompanyID   string    `json:"company_id" gorm:"type:uuid;not null;index"`
	
	// Event Details
	EventType   string    `json:"event_type" gorm:"type:varchar(20);not null"` // entry, exit, dwell, speed_violation
	Latitude    float64   `json:"latitude" gorm:"not null"`
	Longitude   float64   `json:"longitude" gorm:"not null"`
	Speed       float64   `json:"speed" gorm:"default:0"` // km/h
	Heading     float64   `json:"heading" gorm:"default:0"` // degrees
	Accuracy    float64   `json:"accuracy" gorm:"default:0"` // meters
	
	// Timing
	EventTime   time.Time `json:"event_time" gorm:"not null"`
	Duration    int       `json:"duration" gorm:"default:0"` // seconds, for dwell events
	
	// Alert Information
	AlertSent   bool      `json:"alert_sent" gorm:"default:false"`
	AlertType   string    `json:"alert_type" gorm:"type:varchar(30)"` // email, sms, push, webhook
	AlertMessage string   `json:"alert_message" gorm:"type:text"`
	
	// Metadata
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
}

// GeofenceViolation represents a geofence violation
type GeofenceViolation struct {
	ID          string    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GeofenceID  string    `json:"geofence_id" gorm:"type:uuid;not null;index"`
	VehicleID   string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	DriverID    string    `json:"driver_id" gorm:"type:uuid;index"`
	CompanyID   string    `json:"company_id" gorm:"type:uuid;not null;index"`
	
	// Violation Details
	ViolationType string    `json:"violation_type" gorm:"type:varchar(30);not null"` // unauthorized_entry, unauthorized_exit, speed_violation, time_violation
	Severity      string    `json:"severity" gorm:"type:varchar(20);not null"` // low, medium, high, critical
	Description   string    `json:"description" gorm:"type:text"`
	
	// Location and Timing
	Latitude    float64   `json:"latitude" gorm:"not null"`
	Longitude   float64   `json:"longitude" gorm:"not null"`
	Speed       float64   `json:"speed" gorm:"default:0"`
	ViolationTime time.Time `json:"violation_time" gorm:"not null"`
	
	// Resolution
	IsResolved  bool      `json:"is_resolved" gorm:"default:false"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	ResolvedBy  *string   `json:"resolved_by" gorm:"type:uuid"`
	Resolution  string    `json:"resolution" gorm:"type:text"`
	
	// Metadata
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null"`
}

// GeofenceAnalytics represents geofence analytics
type GeofenceAnalytics struct {
	Period              string    `json:"period"`
	TotalGeofences      int       `json:"total_geofences"`
	ActiveGeofences     int       `json:"active_geofences"`
	TotalEvents         int       `json:"total_events"`
	TotalViolations     int       `json:"total_violations"`
	UnresolvedViolations int      `json:"unresolved_violations"`
	EventBreakdown      []EventBreakdown `json:"event_breakdown"`
	ViolationBreakdown  []ViolationBreakdown `json:"violation_breakdown"`
	TopViolatingVehicles []VehicleViolationStats `json:"top_violating_vehicles"`
	GeofenceUtilization []GeofenceUtilization `json:"geofence_utilization"`
}

// EventBreakdown represents event statistics by type
type EventBreakdown struct {
	EventType string `json:"event_type"`
	Count     int    `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ViolationBreakdown represents violation statistics by type
type ViolationBreakdown struct {
	ViolationType string `json:"violation_type"`
	Count         int    `json:"count"`
	Percentage    float64 `json:"percentage"`
	Severity      string `json:"severity"`
}

// VehicleViolationStats represents violation statistics by vehicle
type VehicleViolationStats struct {
	VehicleID       string  `json:"vehicle_id"`
	LicensePlate    string  `json:"license_plate"`
	Make            string  `json:"make"`
	Model           string  `json:"model"`
	TotalViolations int     `json:"total_violations"`
	CriticalViolations int  `json:"critical_violations"`
	LastViolation   time.Time `json:"last_violation"`
	ViolationRate   float64 `json:"violation_rate"` // violations per day
}

// GeofenceUtilization represents geofence utilization statistics
type GeofenceUtilization struct {
	GeofenceID    string  `json:"geofence_id"`
	GeofenceName  string  `json:"geofence_name"`
	TotalEvents   int     `json:"total_events"`
	TotalViolations int   `json:"total_violations"`
	UtilizationRate float64 `json:"utilization_rate"` // events per day
	AverageDwellTime float64 `json:"average_dwell_time"` // minutes
}

// GeofenceCheckRequest represents a request to check geofence status
type GeofenceCheckRequest struct {
	VehicleID   string    `json:"vehicle_id"`
	DriverID    string    `json:"driver_id"`
	CompanyID   string    `json:"company_id"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Speed       float64   `json:"speed"`
	Heading     float64   `json:"heading"`
	Accuracy    float64   `json:"accuracy"`
	Timestamp   time.Time `json:"timestamp"`
}

// GeofenceCheckResult represents the result of a geofence check
type GeofenceCheckResult struct {
	VehicleID       string            `json:"vehicle_id"`
	DriverID        string            `json:"driver_id"`
	CompanyID       string            `json:"company_id"`
	Latitude        float64           `json:"latitude"`
	Longitude       float64           `json:"longitude"`
	Speed           float64           `json:"speed"`
	Timestamp       time.Time         `json:"timestamp"`
	GeofenceEvents  []GeofenceEvent   `json:"geofence_events"`
	Violations      []GeofenceViolation `json:"violations"`
	AlertsGenerated []AlertInfo       `json:"alerts_generated"`
}

// AlertInfo represents information about generated alerts
type AlertInfo struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
	Recipients  []string `json:"recipients"`
	SentAt      time.Time `json:"sent_at"`
}

// NewGeofenceManager creates a new geofence manager
func NewGeofenceManager(db *gorm.DB, redis *redis.Client) *GeofenceManager {
	return &GeofenceManager{
		db:    db,
		redis: redis,
	}
}

// CreateGeofence creates a new geofence
func (gm *GeofenceManager) CreateGeofence(ctx context.Context, geofence *Geofence) error {
	// Validate geofence
	if err := gm.validateGeofence(geofence); err != nil {
		return fmt.Errorf("geofence validation failed: %w", err)
	}

	// Calculate center coordinates if not provided
	if geofence.CenterLat == 0 && geofence.CenterLng == 0 {
		gm.calculateCenter(geofence)
	}

	// Save to database
	if err := gm.db.Create(geofence).Error; err != nil {
		return fmt.Errorf("failed to create geofence: %w", err)
	}

	// Cache the geofence
	gm.cacheGeofence(context.Background(), geofence)

	// Invalidate company geofence cache
	gm.invalidateCompanyGeofenceCache(context.Background(), geofence.CompanyID)

	return nil
}

// UpdateGeofence updates an existing geofence
func (gm *GeofenceManager) UpdateGeofence(ctx context.Context, id string, updates *Geofence) error {
	// Validate updates
	if err := gm.validateGeofence(updates); err != nil {
		return fmt.Errorf("geofence validation failed: %w", err)
	}

	// Calculate center coordinates if coordinates changed
	if len(updates.Coordinates) > 0 {
		gm.calculateCenter(updates)
	}

	// Update in database
	updates.UpdatedAt = time.Now()
	if err := gm.db.Model(&Geofence{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update geofence: %w", err)
	}

	// Get updated geofence
	var updatedGeofence Geofence
	if err := gm.db.First(&updatedGeofence, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to get updated geofence: %w", err)
	}

	// Update cache
	gm.cacheGeofence(context.Background(), &updatedGeofence)

	// Invalidate company geofence cache
	gm.invalidateCompanyGeofenceCache(context.Background(), updatedGeofence.CompanyID)

	return nil
}

// DeleteGeofence deletes a geofence
func (gm *GeofenceManager) DeleteGeofence(ctx context.Context, id string) error {
	// Get geofence to get company ID
	var geofence Geofence
	if err := gm.db.First(&geofence, "id = ?", id).Error; err != nil {
		return fmt.Errorf("geofence not found: %w", err)
	}

	// Delete from database
	if err := gm.db.Delete(&Geofence{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete geofence: %w", err)
	}

	// Remove from cache
	gm.removeGeofenceFromCache(context.Background(), id)

	// Invalidate company geofence cache
	gm.invalidateCompanyGeofenceCache(context.Background(), geofence.CompanyID)

	return nil
}

// GetGeofences retrieves geofences for a company
func (gm *GeofenceManager) GetGeofences(ctx context.Context, companyID string, activeOnly bool) ([]Geofence, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("geofences:%s:%t", companyID, activeOnly)
	cached, err := gm.getCachedGeofences(context.Background(), cacheKey)
	if err == nil && cached != nil {
		return cached, nil
	}

	var geofences []Geofence
	query := gm.db.Where("company_id = ?", companyID)
	
	if activeOnly {
		query = query.Where("is_active = true")
	}
	
	err = query.Order("priority DESC, created_at DESC").Find(&geofences).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get geofences: %w", err)
	}

	// Cache the result
	gm.cacheGeofences(context.Background(), cacheKey, geofences, 30*time.Minute)

	return geofences, nil
}

// GetGeofence retrieves a specific geofence
func (gm *GeofenceManager) GetGeofence(ctx context.Context, id string) (*Geofence, error) {
	// Check cache first
	cached, err := gm.getCachedGeofence(context.Background(), id)
	if err == nil && cached != nil {
		return cached, nil
	}

	var geofence Geofence
	err = gm.db.First(&geofence, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("geofence not found: %w", err)
	}

	// Cache the result
	gm.cacheGeofence(context.Background(), &geofence)

	return &geofence, nil
}

// CheckGeofences checks if a location is within any geofences
func (gm *GeofenceManager) CheckGeofences(ctx context.Context, req *GeofenceCheckRequest) (*GeofenceCheckResult, error) {
	// Get active geofences for the company
	geofences, err := gm.GetGeofences(ctx, req.CompanyID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get geofences: %w", err)
	}

	result := &GeofenceCheckResult{
		VehicleID:       req.VehicleID,
		DriverID:        req.DriverID,
		CompanyID:       req.CompanyID,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		Speed:           req.Speed,
		Timestamp:       req.Timestamp,
		GeofenceEvents:  []GeofenceEvent{},
		Violations:      []GeofenceViolation{},
		AlertsGenerated: []AlertInfo{},
	}

	// Check each geofence
	for _, geofence := range geofences {
		// Check if location is within geofence
		isInside, err := gm.isPointInGeofence(req.Latitude, req.Longitude, &geofence)
		if err != nil {
			continue
		}

		// Check time restrictions
		if !gm.isTimeAllowed(&geofence, req.Timestamp) {
			continue
		}

		// Check vehicle/driver restrictions
		if !gm.isVehicleAllowed(&geofence, req.VehicleID, req.DriverID) {
			continue
		}

		// Get previous state
		wasInside, err := gm.wasVehicleInGeofence(context.Background(), req.VehicleID, geofence.ID)
		if err != nil {
			wasInside = false
		}

		// Generate events based on state changes
		if isInside && !wasInside {
			// Entry event
			event := gm.createGeofenceEvent(&geofence, req, "entry")
			result.GeofenceEvents = append(result.GeofenceEvents, *event)
			
			if geofence.AlertOnEntry {
				alert := gm.generateAlert(&geofence, req, "entry")
				result.AlertsGenerated = append(result.AlertsGenerated, *alert)
			}
		} else if !isInside && wasInside {
			// Exit event
			event := gm.createGeofenceEvent(&geofence, req, "exit")
			result.GeofenceEvents = append(result.GeofenceEvents, *event)
			
			if geofence.AlertOnExit {
				alert := gm.generateAlert(&geofence, req, "exit")
				result.AlertsGenerated = append(result.AlertsGenerated, *alert)
			}
		} else if isInside {
			// Check for speed violations
			if geofence.AlertOnSpeed && geofence.SpeedLimit > 0 && req.Speed > geofence.SpeedLimit {
				event := gm.createGeofenceEvent(&geofence, req, "speed_violation")
				result.GeofenceEvents = append(result.GeofenceEvents, *event)
				
				alert := gm.generateAlert(&geofence, req, "speed_violation")
				result.AlertsGenerated = append(result.AlertsGenerated, *alert)
			}

			// Check for dwell time
			if geofence.AlertOnDwell && geofence.DwellTime > 0 {
				dwellTime, err := gm.getDwellTime(context.Background(), req.VehicleID, geofence.ID)
				if err == nil && dwellTime >= geofence.DwellTime {
					event := gm.createGeofenceEvent(&geofence, req, "dwell")
					event.Duration = dwellTime
					result.GeofenceEvents = append(result.GeofenceEvents, *event)
					
					alert := gm.generateAlert(&geofence, req, "dwell")
					result.AlertsGenerated = append(result.AlertsGenerated, *alert)
				}
			}
		}

		// Check for violations
		violations := gm.checkViolations(&geofence, req, isInside, wasInside)
		result.Violations = append(result.Violations, violations...)
	}

	// Save events and violations to database
	if len(result.GeofenceEvents) > 0 {
		gm.saveGeofenceEvents(context.Background(), result.GeofenceEvents)
	}
	if len(result.Violations) > 0 {
		gm.saveGeofenceViolations(context.Background(), result.Violations)
	}

	// Update vehicle geofence state
	gm.updateVehicleGeofenceState(context.Background(), req.VehicleID, result.GeofenceEvents)

	return result, nil
}

// GetGeofenceEvents retrieves geofence events
func (gm *GeofenceManager) GetGeofenceEvents(_ context.Context, companyID string, filters map[string]interface{}) ([]GeofenceEvent, error) {
	var events []GeofenceEvent
	
	query := gm.db.Where("company_id = ?", companyID)
	
	// Apply filters
	if geofenceID, ok := filters["geofence_id"].(string); ok {
		query = query.Where("geofence_id = ?", geofenceID)
	}
	if vehicleID, ok := filters["vehicle_id"].(string); ok {
		query = query.Where("vehicle_id = ?", vehicleID)
	}
	if driverID, ok := filters["driver_id"].(string); ok {
		query = query.Where("driver_id = ?", driverID)
	}
	if eventType, ok := filters["event_type"].(string); ok {
		query = query.Where("event_type = ?", eventType)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("event_time >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("event_time <= ?", endDate)
	}
	
	err := query.Order("event_time DESC").Limit(1000).Find(&events).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get geofence events: %w", err)
	}
	
	return events, nil
}

// GetGeofenceViolations retrieves geofence violations
func (gm *GeofenceManager) GetGeofenceViolations(_ context.Context, companyID string, filters map[string]interface{}) ([]GeofenceViolation, error) {
	var violations []GeofenceViolation
	
	query := gm.db.Where("company_id = ?", companyID)
	
	// Apply filters
	if geofenceID, ok := filters["geofence_id"].(string); ok {
		query = query.Where("geofence_id = ?", geofenceID)
	}
	if vehicleID, ok := filters["vehicle_id"].(string); ok {
		query = query.Where("vehicle_id = ?", vehicleID)
	}
	if driverID, ok := filters["driver_id"].(string); ok {
		query = query.Where("driver_id = ?", driverID)
	}
	if violationType, ok := filters["violation_type"].(string); ok {
		query = query.Where("violation_type = ?", violationType)
	}
	if severity, ok := filters["severity"].(string); ok {
		query = query.Where("severity = ?", severity)
	}
	if isResolved, ok := filters["is_resolved"].(bool); ok {
		query = query.Where("is_resolved = ?", isResolved)
	}
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query = query.Where("violation_time >= ?", startDate)
	}
	if endDate, ok := filters["end_date"].(time.Time); ok {
		query = query.Where("violation_time <= ?", endDate)
	}
	
	err := query.Order("violation_time DESC").Limit(1000).Find(&violations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get geofence violations: %w", err)
	}
	
	return violations, nil
}

// ResolveViolation resolves a geofence violation
func (gm *GeofenceManager) ResolveViolation(_ context.Context, violationID string, resolvedBy string, resolution string) error {
	updates := map[string]interface{}{
		"is_resolved": true,
		"resolved_at": time.Now(),
		"resolved_by": resolvedBy,
		"resolution":  resolution,
		"updated_at":  time.Now(),
	}

	if err := gm.db.Model(&GeofenceViolation{}).Where("id = ?", violationID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to resolve violation: %w", err)
	}

	return nil
}

// GetGeofenceAnalytics retrieves geofence analytics
func (gm *GeofenceManager) GetGeofenceAnalytics(ctx context.Context, companyID string, startDate, endDate time.Time) (*GeofenceAnalytics, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("geofence_analytics:%s:%s:%s", companyID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	cached, err := gm.getCachedAnalytics(context.Background(), cacheKey)
	if err == nil && cached != nil {
		return cached, nil
	}

	// Get total geofences
	var totalGeofences, activeGeofences int64
	err = gm.db.Model(&Geofence{}).Where("company_id = ?", companyID).Count(&totalGeofences).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get total geofences: %w", err)
	}

	err = gm.db.Model(&Geofence{}).Where("company_id = ? AND is_active = true", companyID).Count(&activeGeofences).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active geofences: %w", err)
	}

	// Get total events
	var totalEvents int64
	err = gm.db.Model(&GeofenceEvent{}).Where("company_id = ? AND event_time BETWEEN ? AND ?", companyID, startDate, endDate).Count(&totalEvents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get total events: %w", err)
	}

	// Get total violations
	var totalViolations, unresolvedViolations int64
	err = gm.db.Model(&GeofenceViolation{}).Where("company_id = ? AND violation_time BETWEEN ? AND ?", companyID, startDate, endDate).Count(&totalViolations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get total violations: %w", err)
	}

	err = gm.db.Model(&GeofenceViolation{}).Where("company_id = ? AND is_resolved = false", companyID).Count(&unresolvedViolations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get unresolved violations: %w", err)
	}

	// Get event breakdown
	eventBreakdown, err := gm.getEventBreakdown(context.Background(), companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get event breakdown: %w", err)
	}

	// Get violation breakdown
	violationBreakdown, err := gm.getViolationBreakdown(context.Background(), companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get violation breakdown: %w", err)
	}

	// Get top violating vehicles
	topViolatingVehicles, err := gm.getTopViolatingVehicles(context.Background(), companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top violating vehicles: %w", err)
	}

	// Get geofence utilization
	geofenceUtilization, err := gm.getGeofenceUtilization(context.Background(), companyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get geofence utilization: %w", err)
	}

	analytics := &GeofenceAnalytics{
		Period:              fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		TotalGeofences:      int(totalGeofences),
		ActiveGeofences:     int(activeGeofences),
		TotalEvents:         int(totalEvents),
		TotalViolations:     int(totalViolations),
		UnresolvedViolations: int(unresolvedViolations),
		EventBreakdown:      eventBreakdown,
		ViolationBreakdown:  violationBreakdown,
		TopViolatingVehicles: topViolatingVehicles,
		GeofenceUtilization: geofenceUtilization,
	}

	// Cache the result
	gm.cacheAnalytics(context.Background(), cacheKey, analytics, 1*time.Hour)

	return analytics, nil
}

// validateGeofence validates a geofence
func (gm *GeofenceManager) validateGeofence(geofence *Geofence) error {
	if geofence.CompanyID == "" {
		return fmt.Errorf("company ID is required")
	}
	if geofence.Name == "" {
		return fmt.Errorf("geofence name is required")
	}
	if geofence.Type == "" {
		return fmt.Errorf("geofence type is required")
	}
	if len(geofence.Coordinates) == 0 && geofence.Type != "circle" {
		return fmt.Errorf("coordinates are required for non-circular geofences")
	}
	if geofence.Type == "circle" && geofence.Radius <= 0 {
		return fmt.Errorf("radius is required for circular geofences")
	}
	if geofence.Priority < 1 || geofence.Priority > 10 {
		return fmt.Errorf("priority must be between 1 and 10")
	}
	return nil
}

// calculateCenter calculates the center coordinates of a geofence
func (gm *GeofenceManager) calculateCenter(geofence *Geofence) {
	if len(geofence.Coordinates) == 0 {
		return
	}

	var sumLat, sumLng float64
	for _, coord := range geofence.Coordinates {
		sumLat += coord.Latitude
		sumLng += coord.Longitude
	}

	geofence.CenterLat = sumLat / float64(len(geofence.Coordinates))
	geofence.CenterLng = sumLng / float64(len(geofence.Coordinates))
}

// isPointInGeofence checks if a point is inside a geofence
func (gm *GeofenceManager) isPointInGeofence(lat, lng float64, geofence *Geofence) (bool, error) {
	switch geofence.Type {
	case "circle":
		distance := gm.calculateDistance(lat, lng, geofence.CenterLat, geofence.CenterLng)
		return distance <= geofence.Radius, nil
	case "polygon":
		return gm.isPointInPolygon(lat, lng, geofence.Coordinates), nil
	case "rectangle":
		return gm.isPointInRectangle(lat, lng, geofence.Coordinates), nil
	default:
		return false, fmt.Errorf("unsupported geofence type: %s", geofence.Type)
	}
}

// isPointInPolygon checks if a point is inside a polygon using ray casting algorithm
func (gm *GeofenceManager) isPointInPolygon(lat, lng float64, coordinates []Coordinate) bool {
	if len(coordinates) < 3 {
		return false
	}

	inside := false
	j := len(coordinates) - 1

	for i := 0; i < len(coordinates); i++ {
		if ((coordinates[i].Latitude > lat) != (coordinates[j].Latitude > lat)) &&
			(lng < (coordinates[j].Longitude-coordinates[i].Longitude)*(lat-coordinates[i].Latitude)/(coordinates[j].Latitude-coordinates[i].Latitude)+coordinates[i].Longitude) {
			inside = !inside
		}
		j = i
	}

	return inside
}

// isPointInRectangle checks if a point is inside a rectangle
func (gm *GeofenceManager) isPointInRectangle(lat, lng float64, coordinates []Coordinate) bool {
	if len(coordinates) != 2 {
		return false
	}

	minLat := math.Min(coordinates[0].Latitude, coordinates[1].Latitude)
	maxLat := math.Max(coordinates[0].Latitude, coordinates[1].Latitude)
	minLng := math.Min(coordinates[0].Longitude, coordinates[1].Longitude)
	maxLng := math.Max(coordinates[0].Longitude, coordinates[1].Longitude)

	return lat >= minLat && lat <= maxLat && lng >= minLng && lng <= maxLng
}

// isTimeAllowed checks if the current time is allowed for the geofence
func (gm *GeofenceManager) isTimeAllowed(geofence *Geofence, timestamp time.Time) bool {
	if len(geofence.TimeRestrictions) == 0 {
		return true
	}

	dayOfWeek := int(timestamp.Weekday())
	timeStr := timestamp.Format("15:04")

	for _, restriction := range geofence.TimeRestrictions {
		if restriction.IsActive && restriction.DayOfWeek == dayOfWeek {
			if timeStr >= restriction.StartTime && timeStr <= restriction.EndTime {
				return true
			}
		}
	}

	return false
}

// isVehicleAllowed checks if a vehicle/driver is allowed in the geofence
func (gm *GeofenceManager) isVehicleAllowed(geofence *Geofence, vehicleID, driverID string) bool {
	// Check if vehicle is restricted
	for _, restrictedVehicle := range geofence.RestrictedVehicles {
		if restrictedVehicle == vehicleID {
			return false
		}
	}

	// Check if driver is restricted
	for _, restrictedDriver := range geofence.RestrictedDrivers {
		if restrictedDriver == driverID {
			return false
		}
	}

	// If there are allowed vehicles/drivers, check if current vehicle/driver is in the list
	if len(geofence.AllowedVehicles) > 0 {
		allowed := false
		for _, allowedVehicle := range geofence.AllowedVehicles {
			if allowedVehicle == vehicleID {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	if len(geofence.AllowedDrivers) > 0 {
		allowed := false
		for _, allowedDriver := range geofence.AllowedDrivers {
			if allowedDriver == driverID {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	return true
}

// wasVehicleInGeofence checks if a vehicle was previously in a geofence
func (gm *GeofenceManager) wasVehicleInGeofence(_ context.Context, _, _ string) (bool, error) {
	// This would check the vehicle's previous geofence state
	// For now, return false as a placeholder
	return false, nil
}

// getDwellTime gets the dwell time for a vehicle in a geofence
func (gm *GeofenceManager) getDwellTime(_ context.Context, _, _ string) (int, error) {
	// This would calculate the dwell time based on entry/exit events
	// For now, return 0 as a placeholder
	return 0, nil
}

// createGeofenceEvent creates a geofence event
func (gm *GeofenceManager) createGeofenceEvent(geofence *Geofence, req *GeofenceCheckRequest, eventType string) *GeofenceEvent {
	return &GeofenceEvent{
		GeofenceID:  geofence.ID,
		VehicleID:   req.VehicleID,
		DriverID:    req.DriverID,
		CompanyID:   req.CompanyID,
		EventType:   eventType,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Speed:       req.Speed,
		Heading:     req.Heading,
		Accuracy:    req.Accuracy,
		EventTime:   req.Timestamp,
		AlertSent:   false,
		CreatedAt:   time.Now(),
	}
}

// generateAlert generates an alert for a geofence event
func (gm *GeofenceManager) generateAlert(geofence *Geofence, req *GeofenceCheckRequest, eventType string) *AlertInfo {
	message := fmt.Sprintf("Geofence %s: %s event for vehicle %s", geofence.Name, eventType, req.VehicleID)
	
	severity := "medium"
	if geofence.Priority >= 8 {
		severity = "high"
	} else if geofence.Priority <= 3 {
		severity = "low"
	}

	return &AlertInfo{
		Type:     "geofence_alert",
		Message:  message,
		Severity: severity,
		SentAt:   time.Now(),
	}
}

// checkViolations checks for geofence violations
func (gm *GeofenceManager) checkViolations(geofence *Geofence, req *GeofenceCheckRequest, isInside, wasInside bool) []GeofenceViolation {
	var violations []GeofenceViolation

	// Check for unauthorized entry/exit
	if !gm.isVehicleAllowed(geofence, req.VehicleID, req.DriverID) {
		if isInside && !wasInside {
			violation := GeofenceViolation{
				GeofenceID:    geofence.ID,
				VehicleID:     req.VehicleID,
				DriverID:      req.DriverID,
				CompanyID:     req.CompanyID,
				ViolationType: "unauthorized_entry",
				Severity:      gm.getViolationSeverity(geofence.Priority),
				Description:   fmt.Sprintf("Unauthorized entry into geofence %s", geofence.Name),
				Latitude:      req.Latitude,
				Longitude:     req.Longitude,
				Speed:         req.Speed,
				ViolationTime: req.Timestamp,
				CreatedAt:     time.Now(),
			}
			violations = append(violations, violation)
		} else if !isInside && wasInside {
			violation := GeofenceViolation{
				GeofenceID:    geofence.ID,
				VehicleID:     req.VehicleID,
				DriverID:      req.DriverID,
				CompanyID:     req.CompanyID,
				ViolationType: "unauthorized_exit",
				Severity:      gm.getViolationSeverity(geofence.Priority),
				Description:   fmt.Sprintf("Unauthorized exit from geofence %s", geofence.Name),
				Latitude:      req.Latitude,
				Longitude:     req.Longitude,
				Speed:         req.Speed,
				ViolationTime: req.Timestamp,
				CreatedAt:     time.Now(),
			}
			violations = append(violations, violation)
		}
	}

	// Check for speed violations
	if geofence.AlertOnSpeed && geofence.SpeedLimit > 0 && req.Speed > geofence.SpeedLimit {
		violation := GeofenceViolation{
			GeofenceID:    geofence.ID,
			VehicleID:     req.VehicleID,
			DriverID:      req.DriverID,
			CompanyID:     req.CompanyID,
			ViolationType: "speed_violation",
			Severity:      "high",
			Description:   fmt.Sprintf("Speed violation in geofence %s: %.1f km/h (limit: %.1f km/h)", geofence.Name, req.Speed, geofence.SpeedLimit),
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
			Speed:         req.Speed,
			ViolationTime: req.Timestamp,
			CreatedAt:     time.Now(),
		}
		violations = append(violations, violation)
	}

	// Check for time violations
	if !gm.isTimeAllowed(geofence, req.Timestamp) && isInside {
		violation := GeofenceViolation{
			GeofenceID:    geofence.ID,
			VehicleID:     req.VehicleID,
			DriverID:      req.DriverID,
			CompanyID:     req.CompanyID,
			ViolationType: "time_violation",
			Severity:      "medium",
			Description:   fmt.Sprintf("Time violation in geofence %s", geofence.Name),
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
			Speed:         req.Speed,
			ViolationTime: req.Timestamp,
			CreatedAt:     time.Now(),
		}
		violations = append(violations, violation)
	}

	return violations
}

// getViolationSeverity determines violation severity based on geofence priority
func (gm *GeofenceManager) getViolationSeverity(priority int) string {
	if priority >= 8 {
		return "critical"
	} else if priority >= 6 {
		return "high"
	} else if priority >= 4 {
		return "medium"
	}
	return "low"
}

// saveGeofenceEvents saves geofence events to database
func (gm *GeofenceManager) saveGeofenceEvents(_ context.Context, events []GeofenceEvent) error {
	if len(events) == 0 {
		return nil
	}

	return gm.db.Create(&events).Error
}

// saveGeofenceViolations saves geofence violations to database
func (gm *GeofenceManager) saveGeofenceViolations(_ context.Context, violations []GeofenceViolation) error {
	if len(violations) == 0 {
		return nil
	}

	return gm.db.Create(&violations).Error
}

// updateVehicleGeofenceState updates the vehicle's geofence state
func (gm *GeofenceManager) updateVehicleGeofenceState(_ context.Context, _ string, _ []GeofenceEvent) error {
	// This would update the vehicle's current geofence state
	// For now, just return nil as a placeholder
	return nil
}

// Helper methods for analytics
func (gm *GeofenceManager) getEventBreakdown(_ context.Context, companyID string, startDate, endDate time.Time) ([]EventBreakdown, error) {
	var breakdown []EventBreakdown
	
	rows, err := gm.db.Model(&GeofenceEvent{}).
		Select("event_type, COUNT(*) as count").
		Where("company_id = ? AND event_time BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("event_type").
		Order("count DESC").
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get event breakdown: %w", err)
	}
	defer rows.Close()
	
	var totalEvents int64
	gm.db.Model(&GeofenceEvent{}).Where("company_id = ? AND event_time BETWEEN ? AND ?", companyID, startDate, endDate).Count(&totalEvents)
	
	for rows.Next() {
		var item EventBreakdown
		err := rows.Scan(&item.EventType, &item.Count)
		if err != nil {
			continue
		}
		
		if totalEvents > 0 {
			item.Percentage = float64(item.Count) / float64(totalEvents) * 100
		}
		
		breakdown = append(breakdown, item)
	}
	
	return breakdown, nil
}

func (gm *GeofenceManager) getViolationBreakdown(_ context.Context, companyID string, startDate, endDate time.Time) ([]ViolationBreakdown, error) {
	var breakdown []ViolationBreakdown
	
	rows, err := gm.db.Model(&GeofenceViolation{}).
		Select("violation_type, severity, COUNT(*) as count").
		Where("company_id = ? AND violation_time BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("violation_type, severity").
		Order("count DESC").
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get violation breakdown: %w", err)
	}
	defer rows.Close()
	
	var totalViolations int64
	gm.db.Model(&GeofenceViolation{}).Where("company_id = ? AND violation_time BETWEEN ? AND ?", companyID, startDate, endDate).Count(&totalViolations)
	
	for rows.Next() {
		var item ViolationBreakdown
		err := rows.Scan(&item.ViolationType, &item.Severity, &item.Count)
		if err != nil {
			continue
		}
		
		if totalViolations > 0 {
			item.Percentage = float64(item.Count) / float64(totalViolations) * 100
		}
		
		breakdown = append(breakdown, item)
	}
	
	return breakdown, nil
}

func (gm *GeofenceManager) getTopViolatingVehicles(_ context.Context, companyID string, startDate, endDate time.Time) ([]VehicleViolationStats, error) {
	var stats []VehicleViolationStats
	
	rows, err := gm.db.Table("geofence_violations gv").
		Select("gv.vehicle_id, v.license_plate, v.make, v.model, COUNT(*) as total_violations, COUNT(CASE WHEN gv.severity = 'critical' THEN 1 END) as critical_violations, MAX(gv.violation_time) as last_violation").
		Joins("JOIN vehicles v ON gv.vehicle_id = v.id").
		Where("gv.company_id = ? AND gv.violation_time BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("gv.vehicle_id, v.license_plate, v.make, v.model").
		Order("total_violations DESC").
		Limit(10).
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get top violating vehicles: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var stat VehicleViolationStats
		err := rows.Scan(&stat.VehicleID, &stat.LicensePlate, &stat.Make, &stat.Model, 
			&stat.TotalViolations, &stat.CriticalViolations, &stat.LastViolation)
		if err != nil {
			continue
		}
		
		// Calculate violation rate (violations per day)
		days := endDate.Sub(startDate).Hours() / 24
		if days > 0 {
			stat.ViolationRate = float64(stat.TotalViolations) / days
		}
		
		stats = append(stats, stat)
	}
	
	return stats, nil
}

func (gm *GeofenceManager) getGeofenceUtilization(_ context.Context, companyID string, startDate, endDate time.Time) ([]GeofenceUtilization, error) {
	var utilization []GeofenceUtilization
	
	rows, err := gm.db.Table("geofence_events ge").
		Select("ge.geofence_id, g.name, COUNT(*) as total_events, COUNT(CASE WHEN gv.id IS NOT NULL THEN 1 END) as total_violations, AVG(ge.duration) as average_dwell_time").
		Joins("JOIN geofences g ON ge.geofence_id = g.id").
		Joins("LEFT JOIN geofence_violations gv ON ge.geofence_id = gv.geofence_id AND ge.vehicle_id = gv.vehicle_id AND ge.event_time = gv.violation_time").
		Where("ge.company_id = ? AND ge.event_time BETWEEN ? AND ?", companyID, startDate, endDate).
		Group("ge.geofence_id, g.name").
		Order("total_events DESC").
		Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to get geofence utilization: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var util GeofenceUtilization
		err := rows.Scan(&util.GeofenceID, &util.GeofenceName, &util.TotalEvents, &util.TotalViolations, &util.AverageDwellTime)
		if err != nil {
			continue
		}
		
		// Calculate utilization rate (events per day)
		days := endDate.Sub(startDate).Hours() / 24
		if days > 0 {
			util.UtilizationRate = float64(util.TotalEvents) / days
		}
		
		utilization = append(utilization, util)
	}
	
	return utilization, nil
}

// Utility methods
func (gm *GeofenceManager) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth's radius in meters

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// Cache methods
func (gm *GeofenceManager) cacheGeofence(_ context.Context, _ *Geofence) error {
	// Implementation would use Redis to cache geofence
	return nil
}

func (gm *GeofenceManager) getCachedGeofence(_ context.Context, _ string) (*Geofence, error) {
	// Implementation would use Redis to get cached geofence
	return nil, fmt.Errorf("cache miss")
}

func (gm *GeofenceManager) removeGeofenceFromCache(_ context.Context, _ string) error {
	// Implementation would use Redis to remove geofence from cache
	return nil
}

func (gm *GeofenceManager) cacheGeofences(_ context.Context, _ string, _ []Geofence, _ time.Duration) error {
	// Implementation would use Redis to cache geofences
	return nil
}

func (gm *GeofenceManager) getCachedGeofences(_ context.Context, _ string) ([]Geofence, error) {
	// Implementation would use Redis to get cached geofences
	return nil, fmt.Errorf("cache miss")
}

func (gm *GeofenceManager) invalidateCompanyGeofenceCache(_ context.Context, _ string) error {
	// Implementation would invalidate company geofence cache
	return nil
}

func (gm *GeofenceManager) cacheAnalytics(_ context.Context, _ string, _ *GeofenceAnalytics, _ time.Duration) error {
	// Implementation would use Redis to cache analytics
	return nil
}

func (gm *GeofenceManager) getCachedAnalytics(_ context.Context, _ string) (*GeofenceAnalytics, error) {
	// Implementation would use Redis to get cached analytics
	return nil, fmt.Errorf("cache miss")
}
