package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VehicleHistory represents a historical event for a vehicle
type VehicleHistory struct {
	ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	VehicleID       string    `json:"vehicle_id" gorm:"type:uuid;not null;index"`
	CompanyID       string    `json:"company_id" gorm:"type:uuid;not null;index"`
	EventType       string    `json:"event_type" gorm:"size:50;not null;index"` // maintenance, repair, status_change, inspection, assignment
	EventCategory   string    `json:"event_category" gorm:"size:50;not null"`   // scheduled, emergency, compliance, operational
	Title           string    `json:"title" gorm:"size:200;not null"`
	Description     string    `json:"description" gorm:"type:text;not null"`
	MileageAtEvent  int       `json:"mileage_at_event" gorm:"default:0"`
	Cost            float64   `json:"cost" gorm:"type:decimal(12,2);default:0"`
	Currency        string    `json:"currency" gorm:"size:3;default:'IDR'"`
	Location        string    `json:"location" gorm:"size:200"`
	ServiceProvider string    `json:"service_provider" gorm:"size:200"`
	InvoiceNumber   string    `json:"invoice_number" gorm:"size:50"`
	Documents       json.RawMessage `json:"documents" gorm:"type:jsonb"` // Receipts, invoices, certificates
	NextServiceDue  *time.Time `json:"next_service_due"`
	CreatedBy       string    `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Vehicle Vehicle `json:"vehicle" gorm:"foreignKey:VehicleID"`
	Company Company `json:"company" gorm:"foreignKey:CompanyID"`
	Creator User    `json:"creator" gorm:"foreignKey:CreatedBy"`
}

// BeforeCreate hook to generate UUID if not provided
func (vh *VehicleHistory) BeforeCreate(tx *gorm.DB) error {
	if vh.ID == "" {
		vh.ID = uuid.New().String()
	}
	return nil
}

// EventType constants
const (
	EventTypeMaintenance   = "maintenance"
	EventTypeRepair        = "repair"
	EventTypeStatusChange  = "status_change"
	EventTypeInspection    = "inspection"
	EventTypeAssignment    = "assignment"
	EventTypeFuel          = "fuel"
	EventTypeInsurance     = "insurance"
	EventTypeRegistration  = "registration"
)

// EventCategory constants
const (
	EventCategoryScheduled   = "scheduled"
	EventCategoryEmergency   = "emergency"
	EventCategoryCompliance  = "compliance"
	EventCategoryOperational = "operational"
	EventCategoryPreventive  = "preventive"
	EventCategoryCorrective  = "corrective"
)

// GetEventTypeDisplay returns a human-readable event type
func (vh *VehicleHistory) GetEventTypeDisplay() string {
	eventTypeMap := map[string]string{
		EventTypeMaintenance:   "Maintenance",
		EventTypeRepair:        "Repair",
		EventTypeStatusChange:  "Status Change",
		EventTypeInspection:    "Inspection",
		EventTypeAssignment:    "Driver Assignment",
		EventTypeFuel:          "Fuel",
		EventTypeInsurance:     "Insurance",
		EventTypeRegistration:  "Registration",
	}
	
	if display, exists := eventTypeMap[vh.EventType]; exists {
		return display
	}
	return vh.EventType
}

// GetEventCategoryDisplay returns a human-readable event category
func (vh *VehicleHistory) GetEventCategoryDisplay() string {
	categoryMap := map[string]string{
		EventCategoryScheduled:   "Scheduled",
		EventCategoryEmergency:   "Emergency",
		EventCategoryCompliance:  "Compliance",
		EventCategoryOperational: "Operational",
		EventCategoryPreventive:  "Preventive",
		EventCategoryCorrective:  "Corrective",
	}
	
	if display, exists := categoryMap[vh.EventCategory]; exists {
		return display
	}
	return vh.EventCategory
}

// IsMaintenanceEvent checks if the event is maintenance-related
func (vh *VehicleHistory) IsMaintenanceEvent() bool {
	return vh.EventType == EventTypeMaintenance || vh.EventType == EventTypeRepair
}

// IsComplianceEvent checks if the event is compliance-related
func (vh *VehicleHistory) IsComplianceEvent() bool {
	return vh.EventCategory == EventCategoryCompliance || vh.EventType == EventTypeInspection
}

// GetFormattedCost returns the cost formatted with currency
func (vh *VehicleHistory) GetFormattedCost() string {
	switch vh.Currency {
	case "IDR":
		return "Rp " + formatIndonesianNumber(vh.Cost)
	case "USD":
		return "$" + formatNumberWithDecimals(vh.Cost, 2)
	default:
		return vh.Currency + " " + formatNumberWithDecimals(vh.Cost, 2)
	}
}

// HasDocuments checks if the event has attached documents
func (vh *VehicleHistory) HasDocuments() bool {
	return vh.Documents != nil && len(vh.Documents) > 0
}

// GetDocumentCount returns the number of documents attached
func (vh *VehicleHistory) GetDocumentCount() int {
	if !vh.HasDocuments() {
		return 0
	}
	
	var docs []map[string]interface{}
	if err := json.Unmarshal(vh.Documents, &docs); err != nil {
		return 0
	}
	
	return len(docs)
}

// IsUpcomingService checks if this is a scheduled service that's due soon
func (vh *VehicleHistory) IsUpcomingService() bool {
	if vh.NextServiceDue == nil {
		return false
	}
	
	// Check if service is due within the next 30 days
	thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)
	return vh.NextServiceDue.Before(thirtyDaysFromNow) && vh.NextServiceDue.After(time.Now())
}

// IsOverdueService checks if this is a scheduled service that's overdue
func (vh *VehicleHistory) IsOverdueService() bool {
	if vh.NextServiceDue == nil {
		return false
	}
	
	return vh.NextServiceDue.Before(time.Now())
}

// GetServiceStatus returns the status of the service
func (vh *VehicleHistory) GetServiceStatus() string {
	if vh.NextServiceDue == nil {
		return "completed"
	}
	
	if vh.IsOverdueService() {
		return "overdue"
	}
	
	if vh.IsUpcomingService() {
		return "upcoming"
	}
	
	return "scheduled"
}

// Helper functions for number formatting
func formatIndonesianNumber(num float64) string {
	// Simple Indonesian number formatting
	// In production, you might want to use a proper localization library
	if num >= 1000000000 {
		return formatNumberWithDecimals(num/1000000000, 1) + "M"
	} else if num >= 1000000 {
		return formatNumberWithDecimals(num/1000000, 1) + "Jt"
	} else if num >= 1000 {
		return formatNumberWithDecimals(num/1000, 1) + "K"
	}
	return formatNumberWithDecimals(num, 0)
}

func formatNumberWithDecimals(num float64, decimals int) string {
	// Simple number formatting
	// In production, you might want to use a proper number formatting library
	if decimals == 0 {
		return fmt.Sprintf("%.0f", num)
	}
	return fmt.Sprintf("%."+strconv.Itoa(decimals)+"f", num)
}

// ValidateEventType validates if the event type is valid
func ValidateEventType(eventType string) bool {
	validTypes := []string{
		EventTypeMaintenance,
		EventTypeRepair,
		EventTypeStatusChange,
		EventTypeInspection,
		EventTypeAssignment,
		EventTypeFuel,
		EventTypeInsurance,
		EventTypeRegistration,
	}
	
	for _, validType := range validTypes {
		if eventType == validType {
			return true
		}
	}
	return false
}

// ValidateEventCategory validates if the event category is valid
func ValidateEventCategory(eventCategory string) bool {
	validCategories := []string{
		EventCategoryScheduled,
		EventCategoryEmergency,
		EventCategoryCompliance,
		EventCategoryOperational,
		EventCategoryPreventive,
		EventCategoryCorrective,
	}
	
	for _, validCategory := range validCategories {
		if eventCategory == validCategory {
			return true
		}
	}
	return false
}
