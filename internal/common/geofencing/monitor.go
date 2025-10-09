package geofencing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// GeofenceMonitor provides real-time geofence monitoring capabilities
type GeofenceMonitor struct {
	db              *gorm.DB
	redis           *redis.Client
	geofenceManager *GeofenceManager
	monitoring      map[string]*VehicleMonitor
	mu              sync.RWMutex
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

// VehicleMonitor represents monitoring state for a vehicle
type VehicleMonitor struct {
	VehicleID       string                 `json:"vehicle_id"`
	DriverID        string                 `json:"driver_id"`
	CompanyID       string                 `json:"company_id"`
	CurrentLocation Location               `json:"current_location"`
	GeofenceStates  map[string]bool        `json:"geofence_states"` // geofence_id -> is_inside
	LastUpdate      time.Time              `json:"last_update"`
	IsActive        bool                   `json:"is_active"`
	StopChan        chan struct{}          `json:"-"`
}

// Location represents a geographical location
type Location struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	Accuracy  float64   `json:"accuracy"`
	Timestamp time.Time `json:"timestamp"`
}

// GeofenceAlert represents a real-time geofence alert
type GeofenceAlert struct {
	ID          string    `json:"id"`
	GeofenceID  string    `json:"geofence_id"`
	VehicleID   string    `json:"vehicle_id"`
	DriverID    string    `json:"driver_id"`
	CompanyID   string    `json:"company_id"`
	AlertType   string    `json:"alert_type"` // entry, exit, dwell, speed_violation, violation
	Severity    string    `json:"severity"` // low, medium, high, critical
	Message     string    `json:"message"`
	Location    Location  `json:"location"`
	Timestamp   time.Time `json:"timestamp"`
	IsRead      bool      `json:"is_read"`
	IsResolved  bool      `json:"is_resolved"`
}

// MonitoringConfig represents configuration for geofence monitoring
type MonitoringConfig struct {
	CheckInterval    time.Duration `json:"check_interval"`    // How often to check geofences
	AlertCooldown    time.Duration `json:"alert_cooldown"`    // Cooldown period for alerts
	MaxDwellTime     time.Duration `json:"max_dwell_time"`    // Maximum dwell time before alert
	SpeedThreshold   float64       `json:"speed_threshold"`   // Speed threshold for violations
	LocationAccuracy float64       `json:"location_accuracy"` // Minimum location accuracy
}

// NewGeofenceMonitor creates a new geofence monitor
func NewGeofenceMonitor(db *gorm.DB, redis *redis.Client, geofenceManager *GeofenceManager) *GeofenceMonitor {
	return &GeofenceMonitor{
		db:              db,
		redis:           redis,
		geofenceManager: geofenceManager,
		monitoring:      make(map[string]*VehicleMonitor),
		stopChan:        make(chan struct{}),
	}
}

// StartMonitoring starts the geofence monitoring service
func (gm *GeofenceMonitor) StartMonitoring(ctx context.Context, config *MonitoringConfig) error {
	if config == nil {
		config = &MonitoringConfig{
			CheckInterval:    30 * time.Second,
			AlertCooldown:    5 * time.Minute,
			MaxDwellTime:     10 * time.Minute,
			SpeedThreshold:   80.0, // km/h
			LocationAccuracy: 100.0, // meters
		}
	}

	gm.wg.Add(1)
	go gm.monitoringLoop(ctx, config)

	return nil
}

// StopMonitoring stops the geofence monitoring service
func (gm *GeofenceMonitor) StopMonitoring() {
	close(gm.stopChan)
	gm.wg.Wait()
}

// AddVehicleToMonitoring adds a vehicle to real-time monitoring
func (gm *GeofenceMonitor) AddVehicleToMonitoring(ctx context.Context, vehicleID, driverID, companyID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Check if vehicle is already being monitored
	if _, exists := gm.monitoring[vehicleID]; exists {
		return fmt.Errorf("vehicle %s is already being monitored", vehicleID)
	}

	// Create vehicle monitor
	monitor := &VehicleMonitor{
		VehicleID:      vehicleID,
		DriverID:       driverID,
		CompanyID:      companyID,
		GeofenceStates: make(map[string]bool),
		LastUpdate:     time.Now(),
		IsActive:       true,
		StopChan:       make(chan struct{}),
	}

	gm.monitoring[vehicleID] = monitor

	// Cache the monitor
	gm.cacheVehicleMonitor(ctx, monitor)

	return nil
}

// RemoveVehicleFromMonitoring removes a vehicle from real-time monitoring
func (gm *GeofenceMonitor) RemoveVehicleFromMonitoring(ctx context.Context, vehicleID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	monitor, exists := gm.monitoring[vehicleID]
	if !exists {
		return fmt.Errorf("vehicle %s is not being monitored", vehicleID)
	}

	// Stop monitoring
	close(monitor.StopChan)
	monitor.IsActive = false

	// Remove from monitoring map
	delete(gm.monitoring, vehicleID)

	// Remove from cache
	gm.removeVehicleMonitorFromCache(ctx, vehicleID)

	return nil
}

// UpdateVehicleLocation updates a vehicle's location for monitoring
func (gm *GeofenceMonitor) UpdateVehicleLocation(ctx context.Context, vehicleID string, location Location) error {
	gm.mu.RLock()
	monitor, exists := gm.monitoring[vehicleID]
	gm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("vehicle %s is not being monitored", vehicleID)
	}

	// Update location
	monitor.CurrentLocation = location
	monitor.LastUpdate = time.Now()

	// Cache the updated monitor
	gm.cacheVehicleMonitor(ctx, monitor)

	// Check geofences immediately
	go gm.checkVehicleGeofences(ctx, monitor, location)

	return nil
}

// GetVehicleMonitoringStatus gets the monitoring status for a vehicle
func (gm *GeofenceMonitor) GetVehicleMonitoringStatus(ctx context.Context, vehicleID string) (*VehicleMonitor, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	monitor, exists := gm.monitoring[vehicleID]
	if !exists {
		return nil, fmt.Errorf("vehicle %s is not being monitored", vehicleID)
	}

	return monitor, nil
}

// GetActiveVehicles gets all actively monitored vehicles
func (gm *GeofenceMonitor) GetActiveVehicles(ctx context.Context, companyID string) ([]VehicleMonitor, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	var activeVehicles []VehicleMonitor
	for _, monitor := range gm.monitoring {
		if monitor.CompanyID == companyID && monitor.IsActive {
			activeVehicles = append(activeVehicles, *monitor)
		}
	}

	return activeVehicles, nil
}

// GetGeofenceAlerts gets recent geofence alerts
func (gm *GeofenceMonitor) GetGeofenceAlerts(ctx context.Context, companyID string, limit int) ([]GeofenceAlert, error) {
	// This would retrieve alerts from database or cache
	// For now, return empty slice as placeholder
	return []GeofenceAlert{}, nil
}

// monitoringLoop runs the main monitoring loop
func (gm *GeofenceMonitor) monitoringLoop(ctx context.Context, config *MonitoringConfig) {
	defer gm.wg.Done()

	ticker := time.NewTicker(config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-gm.stopChan:
			return
		case <-ticker.C:
			gm.performMonitoringCheck(ctx, config)
		}
	}
}

// performMonitoringCheck performs a monitoring check for all vehicles
func (gm *GeofenceMonitor) performMonitoringCheck(ctx context.Context, _ *MonitoringConfig) {
	gm.mu.RLock()
	monitors := make([]*VehicleMonitor, 0, len(gm.monitoring))
	for _, monitor := range gm.monitoring {
		if monitor.IsActive {
			monitors = append(monitors, monitor)
		}
	}
	gm.mu.RUnlock()

	// Check each vehicle
	for _, monitor := range monitors {
		select {
		case <-monitor.StopChan:
			continue
		default:
			gm.checkVehicleGeofences(ctx, monitor, monitor.CurrentLocation)
		}
	}
}

// checkVehicleGeofences checks geofences for a specific vehicle
func (gm *GeofenceMonitor) checkVehicleGeofences(ctx context.Context, monitor *VehicleMonitor, location Location) {
	// Create geofence check request
	req := &GeofenceCheckRequest{
		VehicleID:   monitor.VehicleID,
		DriverID:    monitor.DriverID,
		CompanyID:   monitor.CompanyID,
		Latitude:    location.Latitude,
		Longitude:   location.Longitude,
		Speed:       location.Speed,
		Heading:     location.Heading,
		Accuracy:    location.Accuracy,
		Timestamp:   location.Timestamp,
	}

	// Check geofences
	result, err := gm.geofenceManager.CheckGeofences(ctx, req)
	if err != nil {
		return
	}

	// Process events and violations
	gm.processGeofenceEvents(ctx, monitor, result.GeofenceEvents)
	gm.processGeofenceViolations(ctx, monitor, result.Violations)
	gm.processAlerts(ctx, monitor, result.AlertsGenerated)

	// Update geofence states
	gm.updateGeofenceStates(monitor, result.GeofenceEvents)
}

// processGeofenceEvents processes geofence events
func (gm *GeofenceMonitor) processGeofenceEvents(ctx context.Context, _ *VehicleMonitor, events []GeofenceEvent) {
	for _, event := range events {
		// Create real-time alert
		alert := &GeofenceAlert{
			ID:         fmt.Sprintf("alert_%d", time.Now().UnixNano()),
			GeofenceID: event.GeofenceID,
			VehicleID:  event.VehicleID,
			DriverID:   event.DriverID,
			CompanyID:  event.CompanyID,
			AlertType:  event.EventType,
			Severity:   gm.getAlertSeverity(event.EventType),
			Message:    gm.generateAlertMessage(event),
			Location: Location{
				Latitude:  event.Latitude,
				Longitude: event.Longitude,
				Speed:     event.Speed,
				Heading:   event.Heading,
				Accuracy:  event.Accuracy,
				Timestamp: event.EventTime,
			},
			Timestamp:  time.Now(),
			IsRead:     false,
			IsResolved: false,
		}

		// Send real-time alert
		gm.sendRealTimeAlert(ctx, alert)

		// Cache the alert
		gm.cacheGeofenceAlert(ctx, alert)
	}
}

// processGeofenceViolations processes geofence violations
func (gm *GeofenceMonitor) processGeofenceViolations(ctx context.Context, _ *VehicleMonitor, violations []GeofenceViolation) {
	for _, violation := range violations {
		// Create real-time alert for violation
		alert := &GeofenceAlert{
			ID:         fmt.Sprintf("alert_%d", time.Now().UnixNano()),
			GeofenceID: violation.GeofenceID,
			VehicleID:  violation.VehicleID,
			DriverID:   violation.DriverID,
			CompanyID:  violation.CompanyID,
			AlertType:  "violation",
			Severity:   violation.Severity,
			Message:    violation.Description,
			Location: Location{
				Latitude:  violation.Latitude,
				Longitude: violation.Longitude,
				Speed:     violation.Speed,
				Timestamp: violation.ViolationTime,
			},
			Timestamp:  time.Now(),
			IsRead:     false,
			IsResolved: false,
		}

		// Send real-time alert
		gm.sendRealTimeAlert(ctx, alert)

		// Cache the alert
		gm.cacheGeofenceAlert(ctx, alert)
	}
}

// processAlerts processes generated alerts
func (gm *GeofenceMonitor) processAlerts(ctx context.Context, monitor *VehicleMonitor, alerts []AlertInfo) {
	for _, alertInfo := range alerts {
		// Create real-time alert
		alert := &GeofenceAlert{
			ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
			VehicleID: monitor.VehicleID,
			DriverID:  monitor.DriverID,
			CompanyID: monitor.CompanyID,
			AlertType: "system_alert",
			Severity:  alertInfo.Severity,
			Message:   alertInfo.Message,
			Location:  monitor.CurrentLocation,
			Timestamp: time.Now(),
			IsRead:    false,
			IsResolved: false,
		}

		// Send real-time alert
		gm.sendRealTimeAlert(ctx, alert)

		// Cache the alert
		gm.cacheGeofenceAlert(ctx, alert)
	}
}

// updateGeofenceStates updates the geofence states for a vehicle
func (gm *GeofenceMonitor) updateGeofenceStates(monitor *VehicleMonitor, events []GeofenceEvent) {
	for _, event := range events {
		switch event.EventType {
		case "entry":
			monitor.GeofenceStates[event.GeofenceID] = true
		case "exit":
			monitor.GeofenceStates[event.GeofenceID] = false
		}
	}
}

// getAlertSeverity determines alert severity based on event type
func (gm *GeofenceMonitor) getAlertSeverity(eventType string) string {
	switch eventType {
	case "speed_violation":
		return "high"
	case "dwell":
		return "medium"
	case "entry", "exit":
		return "low"
	default:
		return "medium"
	}
}

// generateAlertMessage generates an alert message for an event
func (gm *GeofenceMonitor) generateAlertMessage(event GeofenceEvent) string {
	switch event.EventType {
	case "entry":
		return fmt.Sprintf("Vehicle entered geofence at %.6f, %.6f", event.Latitude, event.Longitude)
	case "exit":
		return fmt.Sprintf("Vehicle exited geofence at %.6f, %.6f", event.Latitude, event.Longitude)
	case "speed_violation":
		return fmt.Sprintf("Speed violation: %.1f km/h at %.6f, %.6f", event.Speed, event.Latitude, event.Longitude)
	case "dwell":
		return fmt.Sprintf("Vehicle dwelling for %d seconds at %.6f, %.6f", event.Duration, event.Latitude, event.Longitude)
	default:
		return fmt.Sprintf("Geofence event: %s at %.6f, %.6f", event.EventType, event.Latitude, event.Longitude)
	}
}

// sendRealTimeAlert sends a real-time alert
func (gm *GeofenceMonitor) sendRealTimeAlert(_ context.Context, alert *GeofenceAlert) {
	// Publish to Redis for real-time notifications
	_ = map[string]interface{}{
		"type":      "geofence_alert",
		"alert":     alert,
		"timestamp": time.Now(),
	}

	// This would publish to Redis channel for real-time updates
	// For now, just log the alert
	fmt.Printf("Geofence Alert: %s - %s\n", alert.AlertType, alert.Message)
}

// Cache methods
func (gm *GeofenceMonitor) cacheVehicleMonitor(_ context.Context, _ *VehicleMonitor) error {
	// Implementation would use Redis to cache vehicle monitor
	return nil
}

func (gm *GeofenceMonitor) removeVehicleMonitorFromCache(_ context.Context, _ string) error {
	// Implementation would use Redis to remove vehicle monitor from cache
	return nil
}

func (gm *GeofenceMonitor) cacheGeofenceAlert(_ context.Context, _ *GeofenceAlert) error {
	// Implementation would use Redis to cache geofence alert
	return nil
}

// GetMonitoringStats gets monitoring statistics
func (gm *GeofenceMonitor) GetMonitoringStats(ctx context.Context, companyID string) (map[string]interface{}, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	stats := map[string]interface{}{
		"total_monitored_vehicles": 0,
		"active_monitored_vehicles": 0,
		"company_monitored_vehicles": 0,
		"last_update": time.Now(),
	}

	totalMonitored := 0
	activeMonitored := 0
	companyMonitored := 0

	for _, monitor := range gm.monitoring {
		totalMonitored++
		if monitor.IsActive {
			activeMonitored++
		}
		if monitor.CompanyID == companyID {
			companyMonitored++
		}
	}

	stats["total_monitored_vehicles"] = totalMonitored
	stats["active_monitored_vehicles"] = activeMonitored
	stats["company_monitored_vehicles"] = companyMonitored

	return stats, nil
}

// GetVehicleGeofenceStatus gets the current geofence status for a vehicle
func (gm *GeofenceMonitor) GetVehicleGeofenceStatus(ctx context.Context, vehicleID string) (map[string]interface{}, error) {
	gm.mu.RLock()
	monitor, exists := gm.monitoring[vehicleID]
	gm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("vehicle %s is not being monitored", vehicleID)
	}

	status := map[string]interface{}{
		"vehicle_id":      monitor.VehicleID,
		"driver_id":       monitor.DriverID,
		"company_id":      monitor.CompanyID,
		"current_location": monitor.CurrentLocation,
		"geofence_states": monitor.GeofenceStates,
		"last_update":     monitor.LastUpdate,
		"is_active":       monitor.IsActive,
	}

	return status, nil
}

// SetMonitoringConfig updates monitoring configuration for a vehicle
func (gm *GeofenceMonitor) SetMonitoringConfig(_ context.Context, vehicleID string, config *MonitoringConfig) error {
	gm.mu.RLock()
	_, exists := gm.monitoring[vehicleID]
	gm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("vehicle %s is not being monitored", vehicleID)
	}

	// Update monitoring configuration
	// This would store the configuration for the specific vehicle
	// For now, just return nil as a placeholder
	return nil
}

// GetGeofenceViolationTrends gets violation trends for analytics
func (gm *GeofenceMonitor) GetGeofenceViolationTrends(_ context.Context, companyID string, days int) (map[string]interface{}, error) {
	// This would analyze violation trends over time
	// For now, return empty trends as placeholder
	trends := map[string]interface{}{
		"period": fmt.Sprintf("last_%d_days", days),
		"trends": []map[string]interface{}{},
	}

	return trends, nil
}
