package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Alert represents a real-time alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	CompanyID   string                 `json:"company_id"`
	UserID      string                 `json:"user_id,omitempty"`
	VehicleID   string                 `json:"vehicle_id,omitempty"`
	DriverID    string                 `json:"driver_id,omitempty"`
	Severity    string                 `json:"severity"` // "low", "medium", "high", "critical"
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Read        bool                   `json:"read"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// AlertSystem manages real-time alerts
type AlertSystem struct {
	hub   *WebSocketHub
	redis *redis.Client
}

// NewAlertSystem creates a new alert system
func NewAlertSystem(hub *WebSocketHub, redis *redis.Client) *AlertSystem {
	return &AlertSystem{
		hub:   hub,
		redis: redis,
	}
}

// Alert types
const (
	AlertTypeSpeedViolation     = "speed_violation"
	AlertTypeGeofenceViolation  = "geofence_violation"
	AlertTypeMaintenance        = "maintenance"
	AlertTypeFuelTheft          = "fuel_theft"
	AlertTypeDriverEvent        = "driver_event"
	AlertTypeVehicleOffline     = "vehicle_offline"
	AlertTypeSystemError        = "system_error"
	AlertTypePaymentReceived    = "payment_received"
	AlertTypeInvoiceGenerated   = "invoice_generated"
)

// Alert severities
const (
	AlertSeverityLow      = "low"
	AlertSeverityMedium   = "medium"
	AlertSeverityHigh     = "high"
	AlertSeverityCritical = "critical"
)

// CreateAlert creates a new alert
func (as *AlertSystem) CreateAlert(ctx context.Context, alert *Alert) error {
	// Set default values
	if alert.ID == "" {
		alert.ID = fmt.Sprintf("alert_%d", time.Now().UnixNano())
	}
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}
	if alert.Severity == "" {
		alert.Severity = AlertSeverityMedium
	}
	
	// Store alert in Redis
	alertKey := fmt.Sprintf("alert:%s:%s", alert.CompanyID, alert.ID)
	alertData, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}
	
	if err := as.redis.Set(ctx, alertKey, alertData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store alert: %w", err)
	}
	
	// Add to company alerts list
	companyAlertsKey := fmt.Sprintf("company_alerts:%s", alert.CompanyID)
	if err := as.redis.LPush(ctx, companyAlertsKey, alert.ID).Err(); err != nil {
		return fmt.Errorf("failed to add alert to company list: %w", err)
	}
	
	// Set expiration for company alerts list
	as.redis.Expire(ctx, companyAlertsKey, 24*time.Hour)
	
	// Broadcast alert via WebSocket
	message := WebSocketMessage{
		Type:      "alert",
		Data:      alert,
		Timestamp: time.Now(),
		CompanyID: alert.CompanyID,
		UserID:    alert.UserID,
	}
	
	if alert.UserID != "" {
		as.hub.BroadcastToUser(alert.CompanyID, alert.UserID, message)
	} else {
		as.hub.BroadcastToCompany(alert.CompanyID, message)
	}
	
	// Publish to Redis for cross-instance communication
	return as.publishAlertToRedis(message)
}

// GetCompanyAlerts retrieves alerts for a company
func (as *AlertSystem) GetCompanyAlerts(ctx context.Context, companyID string, limit int64) ([]*Alert, error) {
	companyAlertsKey := fmt.Sprintf("company_alerts:%s", companyID)
	
	// Get alert IDs
	alertIDs, err := as.redis.LRange(ctx, companyAlertsKey, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get alert IDs: %w", err)
	}
	
	var alerts []*Alert
	for _, alertID := range alertIDs {
		alertKey := fmt.Sprintf("alert:%s:%s", companyID, alertID)
		alertData, err := as.redis.Get(ctx, alertKey).Result()
		if err != nil {
			if err == redis.Nil {
				continue // Alert expired
			}
			return nil, fmt.Errorf("failed to get alert %s: %w", alertID, err)
		}
		
		var alert Alert
		if err := json.Unmarshal([]byte(alertData), &alert); err != nil {
			return nil, fmt.Errorf("failed to unmarshal alert %s: %w", alertID, err)
		}
		
		alerts = append(alerts, &alert)
	}
	
	return alerts, nil
}

// MarkAlertAsRead marks an alert as read
func (as *AlertSystem) MarkAlertAsRead(ctx context.Context, companyID, alertID string) error {
	alertKey := fmt.Sprintf("alert:%s:%s", companyID, alertID)
	
	// Get current alert
	alertData, err := as.redis.Get(ctx, alertKey).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("alert not found")
		}
		return fmt.Errorf("failed to get alert: %w", err)
	}
	
	var alert Alert
	if err := json.Unmarshal([]byte(alertData), &alert); err != nil {
		return fmt.Errorf("failed to unmarshal alert: %w", err)
	}
	
	// Update alert
	alert.Read = true
	updatedData, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal updated alert: %w", err)
	}
	
	// Store updated alert
	if err := as.redis.Set(ctx, alertKey, updatedData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}
	
	// Broadcast update
	message := WebSocketMessage{
		Type:      "alert_read",
		Data:      map[string]string{"alert_id": alertID, "company_id": companyID},
		Timestamp: time.Now(),
		CompanyID: companyID,
	}
	
	as.hub.BroadcastToCompany(companyID, message)
	
	return nil
}

// DeleteAlert deletes an alert
func (as *AlertSystem) DeleteAlert(ctx context.Context, companyID, alertID string) error {
	alertKey := fmt.Sprintf("alert:%s:%s", companyID, alertID)
	
	// Delete alert
	if err := as.redis.Del(ctx, alertKey).Err(); err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}
	
	// Remove from company alerts list
	companyAlertsKey := fmt.Sprintf("company_alerts:%s", companyID)
	if err := as.redis.LRem(ctx, companyAlertsKey, 0, alertID).Err(); err != nil {
		return fmt.Errorf("failed to remove alert from company list: %w", err)
	}
	
	// Broadcast deletion
	message := WebSocketMessage{
		Type:      "alert_deleted",
		Data:      map[string]string{"alert_id": alertID, "company_id": companyID},
		Timestamp: time.Now(),
		CompanyID: companyID,
	}
	
	as.hub.BroadcastToCompany(companyID, message)
	
	return nil
}

// CreateSpeedViolationAlert creates a speed violation alert
func (as *AlertSystem) CreateSpeedViolationAlert(ctx context.Context, companyID, vehicleID, driverID string, speed, speedLimit float64, location string) error {
	alert := &Alert{
		Type:      AlertTypeSpeedViolation,
		CompanyID: companyID,
		VehicleID: vehicleID,
		DriverID:  driverID,
		Severity:  AlertSeverityHigh,
		Title:     "Speed Violation Detected",
		Message:   fmt.Sprintf("Vehicle exceeded speed limit by %.1f km/h at %s", speed-speedLimit, location),
		Data: map[string]interface{}{
			"speed":       speed,
			"speed_limit": speedLimit,
			"location":    location,
		},
	}
	
	return as.CreateAlert(ctx, alert)
}

// CreateGeofenceViolationAlert creates a geofence violation alert
func (as *AlertSystem) CreateGeofenceViolationAlert(ctx context.Context, companyID, vehicleID, driverID, geofenceID, geofenceName, violationType string) error {
	alert := &Alert{
		Type:      AlertTypeGeofenceViolation,
		CompanyID: companyID,
		VehicleID: vehicleID,
		DriverID:  driverID,
		Severity:  AlertSeverityMedium,
		Title:     "Geofence Violation",
		Message:   fmt.Sprintf("Vehicle %s geofence '%s'", violationType, geofenceName),
		Data: map[string]interface{}{
			"geofence_id":   geofenceID,
			"geofence_name": geofenceName,
			"violation_type": violationType,
		},
	}
	
	return as.CreateAlert(ctx, alert)
}

// CreateMaintenanceAlert creates a maintenance alert
func (as *AlertSystem) CreateMaintenanceAlert(ctx context.Context, companyID, vehicleID, alertType, description string) error {
	severity := AlertSeverityMedium
	if alertType == "critical" {
		severity = AlertSeverityCritical
	}
	
	alert := &Alert{
		Type:      AlertTypeMaintenance,
		CompanyID: companyID,
		VehicleID: vehicleID,
		Severity:  severity,
		Title:     "Maintenance Required",
		Message:   description,
		Data: map[string]interface{}{
			"alert_type": alertType,
		},
	}
	
	return as.CreateAlert(ctx, alert)
}

// CreateFuelTheftAlert creates a fuel theft alert
func (as *AlertSystem) CreateFuelTheftAlert(ctx context.Context, companyID, vehicleID, driverID string, fuelAmount float64, location string) error {
	alert := &Alert{
		Type:      AlertTypeFuelTheft,
		CompanyID: companyID,
		VehicleID: vehicleID,
		DriverID:  driverID,
		Severity:  AlertSeverityCritical,
		Title:     "Potential Fuel Theft Detected",
		Message:   fmt.Sprintf("Unusual fuel consumption detected: %.2f liters at %s", fuelAmount, location),
		Data: map[string]interface{}{
			"fuel_amount": fuelAmount,
			"location":    location,
		},
	}
	
	return as.CreateAlert(ctx, alert)
}

// CreateVehicleOfflineAlert creates a vehicle offline alert
func (as *AlertSystem) CreateVehicleOfflineAlert(ctx context.Context, companyID, vehicleID string, lastSeen time.Time) error {
	alert := &Alert{
		Type:      AlertTypeVehicleOffline,
		CompanyID: companyID,
		VehicleID: vehicleID,
		Severity:  AlertSeverityHigh,
		Title:     "Vehicle Offline",
		Message:   fmt.Sprintf("Vehicle has been offline since %s", lastSeen.Format("2006-01-02 15:04:05")),
		Data: map[string]interface{}{
			"last_seen": lastSeen,
		},
	}
	
	return as.CreateAlert(ctx, alert)
}

// CreatePaymentReceivedAlert creates a payment received alert
func (as *AlertSystem) CreatePaymentReceivedAlert(ctx context.Context, companyID, invoiceID string, amount float64) error {
	alert := &Alert{
		Type:      AlertTypePaymentReceived,
		CompanyID: companyID,
		Severity:  AlertSeverityLow,
		Title:     "Payment Received",
		Message:   fmt.Sprintf("Payment of IDR %.2f received for invoice %s", amount, invoiceID),
		Data: map[string]interface{}{
			"invoice_id": invoiceID,
			"amount":     amount,
		},
	}
	
	return as.CreateAlert(ctx, alert)
}

// publishAlertToRedis publishes an alert to Redis for cross-instance communication
func (as *AlertSystem) publishAlertToRedis(message WebSocketMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal alert message: %w", err)
	}
	
	return as.redis.Publish(context.Background(), "fleet_tracker:alerts", data).Err()
}

// StartAlertCleanup starts periodic cleanup of expired alerts
func (as *AlertSystem) StartAlertCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			as.cleanupExpiredAlerts(ctx)
		}
	}
}

// cleanupExpiredAlerts removes expired alerts
func (as *AlertSystem) cleanupExpiredAlerts(ctx context.Context) {
	// Get all company alert keys
	pattern := "company_alerts:*"
	keys, err := as.redis.Keys(ctx, pattern).Result()
	if err != nil {
		fmt.Printf("Failed to get company alert keys: %v\n", err)
		return
	}
	
	for _, key := range keys {
		// Get alert IDs for this company
		alertIDs, err := as.redis.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			continue
		}
		
		// Extract company ID from key
		companyID := key[len("company_alerts:"):]
		
		for _, alertID := range alertIDs {
			alertKey := fmt.Sprintf("alert:%s:%s", companyID, alertID)
			
			// Check if alert exists and is expired
			alertData, err := as.redis.Get(ctx, alertKey).Result()
			if err != nil {
				if err == redis.Nil {
					// Alert doesn't exist, remove from list
					as.redis.LRem(ctx, key, 0, alertID)
				}
				continue
			}
			
			var alert Alert
			if err := json.Unmarshal([]byte(alertData), &alert); err != nil {
				continue
			}
			
			// Check if alert is expired
			if alert.ExpiresAt != nil && time.Now().After(*alert.ExpiresAt) {
				as.redis.Del(ctx, alertKey)
				as.redis.LRem(ctx, key, 0, alertID)
			}
		}
	}
}
