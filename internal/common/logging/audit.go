package logging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditLogger provides audit trail logging functionality
type AuditLogger struct {
	logger *Logger
	db     *gorm.DB
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *Logger, db *gorm.DB) *AuditLogger {
	return &AuditLogger{
		logger: logger,
		db:     db,
	}
}

// AuditEvent represents an audit event
type AuditEvent struct {
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id"`
	UserID     string                 `json:"user_id"`
	CompanyID  string                 `json:"company_id"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// LogCreate logs creation of a resource
func (al *AuditLogger) LogCreate(ctx context.Context, resource, resourceID, userID, companyID string, data interface{}) {
	event := AuditEvent{
		Action:     "create",
		Resource:   resource,
		ResourceID: resourceID,
		UserID:     userID,
		CompanyID:  companyID,
		Timestamp:  time.Now(),
	}

	// Add data as changes
	if data != nil {
		dataBytes, _ := json.Marshal(data)
		var changes map[string]interface{}
		json.Unmarshal(dataBytes, &changes)
		event.Changes = changes
	}

	al.logEvent(ctx, &event)
}

// LogUpdate logs update of a resource
func (al *AuditLogger) LogUpdate(ctx context.Context, resource, resourceID, userID, companyID string, oldData, newData interface{}) {
	event := AuditEvent{
		Action:     "update",
		Resource:   resource,
		ResourceID: resourceID,
		UserID:     userID,
		CompanyID:  companyID,
		Timestamp:  time.Now(),
	}

	// Calculate changes
	changes := make(map[string]interface{})
	if oldData != nil && newData != nil {
		oldBytes, _ := json.Marshal(oldData)
		newBytes, _ := json.Marshal(newData)
		
		var oldMap, newMap map[string]interface{}
		json.Unmarshal(oldBytes, &oldMap)
		json.Unmarshal(newBytes, &newMap)

		for key, newValue := range newMap {
			if oldValue, exists := oldMap[key]; !exists || oldValue != newValue {
				changes[key] = map[string]interface{}{
					"old": oldValue,
					"new": newValue,
				}
			}
		}
	}

	event.Changes = changes
	al.logEvent(ctx, &event)
}

// LogDelete logs deletion of a resource
func (al *AuditLogger) LogDelete(ctx context.Context, resource, resourceID, userID, companyID string) {
	event := AuditEvent{
		Action:     "delete",
		Resource:   resource,
		ResourceID: resourceID,
		UserID:     userID,
		CompanyID:  companyID,
		Timestamp:  time.Now(),
	}

	al.logEvent(ctx, &event)
}

// LogAccess logs access to a resource
func (al *AuditLogger) LogAccess(ctx context.Context, resource, resourceID, userID, companyID string) {
	event := AuditEvent{
		Action:     "access",
		Resource:   resource,
		ResourceID: resourceID,
		UserID:     userID,
		CompanyID:  companyID,
		Timestamp:  time.Now(),
	}

	al.logEvent(ctx, &event)
}

// LogSecurityEvent logs security-related events
func (al *AuditLogger) LogSecurityEvent(ctx context.Context, eventType, userID, ipAddress string, metadata map[string]interface{}) {
	event := AuditEvent{
		Action:     "security_event",
		Resource:   eventType,
		UserID:     userID,
		IPAddress:  ipAddress,
		Metadata:   metadata,
		Timestamp:  time.Now(),
	}

	al.logEvent(ctx, &event)
}

// LogAuthEvent logs authentication events
func (al *AuditLogger) LogAuthEvent(action, userID, email, ipAddress string, success bool) {
	metadata := map[string]interface{}{
		"success": success,
		"email":   email,
	}

	event := AuditEvent{
		Action:     action, // login, logout, register, password_change
		Resource:   "auth",
		UserID:     userID,
		IPAddress:  ipAddress,
		Metadata:   metadata,
		Timestamp:  time.Now(),
	}

	if success {
		al.logger.Info("Authentication event",
			"action", action,
			"user_id", userID,
			"email", email,
			"ip_address", ipAddress,
		)
	} else {
		al.logger.Warn("Authentication failed",
			"action", action,
			"email", email,
			"ip_address", ipAddress,
		)
	}

	al.logEvent(context.Background(), &event)
}

// LogPaymentEvent logs payment-related events
func (al *AuditLogger) LogPaymentEvent(ctx context.Context, action, paymentID, invoiceID, userID, companyID string, amount float64, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["amount"] = amount
	metadata["invoice_id"] = invoiceID

	event := AuditEvent{
		Action:     action, // payment_created, payment_completed, payment_failed
		Resource:   "payment",
		ResourceID: paymentID,
		UserID:     userID,
		CompanyID:  companyID,
		Metadata:   metadata,
		Timestamp:  time.Now(),
	}

	al.logger.Info("Payment event",
		"action", action,
		"payment_id", paymentID,
		"amount", amount,
		"company_id", companyID,
	)

	al.logEvent(ctx, &event)
}

// LogDriverEvent logs driver-related events
func (al *AuditLogger) LogDriverEvent(ctx context.Context, action, driverID, vehicleID, companyID string, metadata map[string]interface{}) {
	event := AuditEvent{
		Action:     action, // driver_assigned, driver_unassigned, violation_detected
		Resource:   "driver",
		ResourceID: driverID,
		CompanyID:  companyID,
		Metadata:   metadata,
		Timestamp:  time.Now(),
	}

	if metadata != nil {
		event.Metadata = metadata
		event.Metadata["vehicle_id"] = vehicleID
	}

	al.logEvent(ctx, &event)
}

// LogGeofenceViolation logs geofence violations
func (al *AuditLogger) LogGeofenceViolation(ctx context.Context, vehicleID, driverID, geofenceID, companyID string, violationType string, location map[string]interface{}) {
	metadata := map[string]interface{}{
		"violation_type": violationType,
		"vehicle_id":     vehicleID,
		"driver_id":      driverID,
		"geofence_id":    geofenceID,
		"location":       location,
	}

	event := AuditEvent{
		Action:     "geofence_violation",
		Resource:   "geofence",
		ResourceID: geofenceID,
		CompanyID:  companyID,
		Metadata:   metadata,
		Timestamp:  time.Now(),
	}

	al.logger.Warn("Geofence violation detected",
		"vehicle_id", vehicleID,
		"driver_id", driverID,
		"geofence_id", geofenceID,
		"violation_type", violationType,
	)

	al.logEvent(ctx, &event)
}

// logEvent persists audit event to database and logger
func (al *AuditLogger) logEvent(_ context.Context, event *AuditEvent) {
	// Log to structured logger
	fields := map[string]interface{}{
		"action":      event.Action,
		"resource":    event.Resource,
		"resource_id": event.ResourceID,
		"user_id":     event.UserID,
		"company_id":  event.CompanyID,
		"ip_address":  event.IPAddress,
		"timestamp":   event.Timestamp,
	}

	if event.Changes != nil {
		fields["changes"] = event.Changes
	}
	if event.Metadata != nil {
		fields["metadata"] = event.Metadata
	}

	al.logger.WithFields(fields).Info("Audit event recorded")

	// Persist to database (async to not block request)
	go func() {
		if al.db != nil {
			changesJSON, _ := json.Marshal(event.Changes)
			metadataJSON, _ := json.Marshal(event.Metadata)

			auditLog := map[string]interface{}{
				"user_id":     event.UserID,
				"company_id":  event.CompanyID,
				"action":      event.Action,
				"resource":    event.Resource,
				"resource_id": event.ResourceID,
				"ip_address":  event.IPAddress,
				"user_agent":  event.UserAgent,
				"details": map[string]interface{}{
					"changes":  string(changesJSON),
					"metadata": string(metadataJSON),
				},
			}

			al.db.Table("audit_logs").Create(auditLog)
		}
	}()
}

// AuditMiddleware creates audit logs for state-changing operations
func AuditMiddleware(auditLogger *AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only audit state-changing operations
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get user info
		userID, _ := c.Get("user_id")
		companyID, _ := c.Get("company_id")

		// Extract resource from path
		resource := extractResource(c.Request.URL.Path)
		resourceID := c.Param("id")

		// Process request
		c.Next()

		// Log if successful
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			action := getActionFromMethod(c.Request.Method)
			
			auditLogger.logger.LogAudit(
				action,
				resource,
				resourceID,
				userIDStr(userID),
				map[string]interface{}{
					"company_id": companyID,
					"ip_address": c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				},
			)
		}
	}
}

// Helper functions

func extractResource(path string) string {
	// Extract resource from path like /api/v1/vehicles/123 -> vehicles
	parts := splitPath(path)
	for i, part := range parts {
		if part == "api" || part == "v1" || part == "admin" {
			if i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	return "unknown"
}

func splitPath(path string) []string {
	result := []string{}
	current := ""
	for _, char := range path {
		if char == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func getActionFromMethod(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

func userIDStr(userID interface{}) string {
	if userID == nil {
		return ""
	}
	if str, ok := userID.(string); ok {
		return str
	}
	return ""
}

