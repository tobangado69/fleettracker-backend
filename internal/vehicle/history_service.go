package vehicle

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/internal/common/repository"
	apperrors "github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
	"gorm.io/gorm"
)

// VehicleHistoryService handles vehicle history business logic
type VehicleHistoryService struct {
	db          *gorm.DB
	repoManager *repository.RepositoryManager
}

// NewVehicleHistoryService creates a new vehicle history service
func NewVehicleHistoryService(db *gorm.DB, repoManager *repository.RepositoryManager) *VehicleHistoryService {
	return &VehicleHistoryService{
		db:          db,
		repoManager: repoManager,
	}
}

// AddHistoryRequest represents a request to add vehicle history
type AddHistoryRequest struct {
	EventType       string                 `json:"event_type" validate:"required,oneof=maintenance repair status_change inspection assignment fuel insurance registration"`
	EventCategory   string                 `json:"event_category" validate:"required,oneof=scheduled emergency compliance operational preventive corrective"`
	Title           string                 `json:"title" validate:"required,min=5,max=200"`
	Description     string                 `json:"description" validate:"required,min=10,max=1000"`
	MileageAtEvent  int                    `json:"mileage_at_event" validate:"min=0"`
	Cost            float64                `json:"cost" validate:"min=0"`
	Currency        string                 `json:"currency" validate:"len=3"`
	Location        string                 `json:"location" validate:"max=200"`
	ServiceProvider string                 `json:"service_provider" validate:"max=200"`
	InvoiceNumber   string                 `json:"invoice_number" validate:"max=50"`
	Documents       []map[string]interface{} `json:"documents"`
	NextServiceDue  *time.Time             `json:"next_service_due"`
}

// VehicleHistoryResponse represents a vehicle history response
type VehicleHistoryResponse struct {
	ID              string                 `json:"id"`
	VehicleID       string                 `json:"vehicle_id"`
	EventType       string                 `json:"event_type"`
	EventCategory   string                 `json:"event_category"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	MileageAtEvent  int                    `json:"mileage_at_event"`
	Cost            float64                `json:"cost"`
	Currency        string                 `json:"currency"`
	FormattedCost   string                 `json:"formatted_cost"`
	Location        string                 `json:"location"`
	ServiceProvider string                 `json:"service_provider"`
	InvoiceNumber   string                 `json:"invoice_number"`
	Documents       []map[string]interface{} `json:"documents"`
	DocumentCount   int                    `json:"document_count"`
	NextServiceDue  *time.Time             `json:"next_service_due"`
	ServiceStatus   string                 `json:"service_status"`
	CreatedBy       string                 `json:"created_by"`
	CreatorName     string                 `json:"creator_name"`
	CreatedAt       time.Time              `json:"created_at"`
}

// HistoryFilters represents filters for vehicle history queries
type HistoryFilters struct {
	EventType       *string    `json:"event_type" form:"event_type"`
	EventCategory   *string    `json:"event_category" form:"event_category"`
	StartDate       *time.Time `json:"start_date" form:"start_date"`
	EndDate         *time.Time `json:"end_date" form:"end_date"`
	MinCost         *float64   `json:"min_cost" form:"min_cost"`
	MaxCost         *float64   `json:"max_cost" form:"max_cost"`
	ServiceProvider *string    `json:"service_provider" form:"service_provider"`
	Search          *string    `json:"search" form:"search"`
	
	// Pagination
	Page            int        `json:"page" form:"page" validate:"min=1"`
	Limit           int        `json:"limit" form:"limit" validate:"min=1,max=100"`
	SortBy          string     `json:"sort_by" form:"sort_by" validate:"oneof=created_at cost mileage_at_event"`
	SortOrder       string     `json:"sort_order" form:"sort_order" validate:"oneof=asc desc"`
}

// MaintenanceScheduleRequest represents a maintenance schedule update request
type MaintenanceScheduleRequest struct {
	NextServiceDue  *time.Time `json:"next_service_due" validate:"required"`
	MaintenanceType string     `json:"maintenance_type" validate:"required,oneof=regular major inspection"`
	MileageInterval int        `json:"mileage_interval" validate:"min=1000"`
	TimeInterval    int        `json:"time_interval"` // days
}

// AddVehicleHistory adds a new history entry for a vehicle
func (s *VehicleHistoryService) AddVehicleHistory(ctx context.Context, companyID, vehicleID, userID string, req AddHistoryRequest) (*VehicleHistoryResponse, error) {
	// Validate vehicle belongs to company
	vehicleRepo := s.repoManager.GetVehicles()
	vehicle, err := vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("Vehicle").WithInternal(err)
	}
	
	if vehicle.CompanyID != companyID {
		return nil, apperrors.NewForbiddenError("Vehicle does not belong to this company")
	}
	
	// Validate event type and category
	if !models.ValidateEventType(req.EventType) {
		return nil, apperrors.NewValidationError("Invalid event type")
	}
	
	if !models.ValidateEventCategory(req.EventCategory) {
		return nil, apperrors.NewValidationError("Invalid event category")
	}
	
	// Set default currency to IDR if not provided
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	
	// Convert documents to JSON
	var documentsJSON json.RawMessage
	if req.Documents != nil && len(req.Documents) > 0 {
		documentsJSON, err = json.Marshal(req.Documents)
		if err != nil {
			return nil, apperrors.NewInternalError("Failed to marshal documents").WithInternal(err)
		}
	}
	
	// Create vehicle history entry
	history := &models.VehicleHistory{
		VehicleID:       vehicleID,
		CompanyID:       companyID,
		EventType:       req.EventType,
		EventCategory:   req.EventCategory,
		Title:           req.Title,
		Description:     req.Description,
		MileageAtEvent:  req.MileageAtEvent,
		Cost:            req.Cost,
		Currency:        req.Currency,
		Location:        req.Location,
		ServiceProvider: req.ServiceProvider,
		InvoiceNumber:   req.InvoiceNumber,
		Documents:       documentsJSON,
		NextServiceDue:  req.NextServiceDue,
		CreatedBy:       userID,
	}
	
	// Save to database
	historyRepo := s.repoManager.GetVehicleHistories()
	if err := historyRepo.Create(ctx, history); err != nil {
		return nil, apperrors.NewInternalError("Failed to create vehicle history").WithInternal(err)
	}
	
	// Update vehicle odometer if mileage is provided
	if req.MileageAtEvent > 0 {
		vehicle.OdometerReading = float64(req.MileageAtEvent)
		if err := vehicleRepo.Update(ctx, vehicle); err != nil {
			// Log error but don't fail the history creation
			fmt.Printf("Warning: failed to update vehicle odometer: %v\n", err)
		}
	}
	
	return s.historyToResponse(history), nil
}

// GetVehicleHistory retrieves vehicle history with filters
func (s *VehicleHistoryService) GetVehicleHistory(ctx context.Context, companyID, vehicleID string, filters HistoryFilters) ([]*VehicleHistoryResponse, error) {
	// Validate vehicle belongs to company
	vehicleRepo := s.repoManager.GetVehicles()
	vehicle, err := vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("Vehicle").WithInternal(err)
	}
	
	if vehicle.CompanyID != companyID {
		return nil, apperrors.NewForbiddenError("Vehicle does not belong to this company")
	}
	
	// Set default pagination
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	
	// Set default sorting
	if filters.SortBy == "" {
		filters.SortBy = "created_at"
	}
	if filters.SortOrder == "" {
		filters.SortOrder = "desc"
	}
	
	pagination := repository.Pagination{
		Page:     filters.Page,
		PageSize: filters.Limit,
		Offset:   (filters.Page - 1) * filters.Limit,
		Limit:    filters.Limit,
	}
	
	historyRepo := s.repoManager.GetVehicleHistories()
	var histories []*models.VehicleHistory
	
	// Apply filters
	if filters.EventType != nil {
		histories, err = historyRepo.GetByVehicleAndType(ctx, vehicleID, *filters.EventType, pagination)
	} else if filters.StartDate != nil && filters.EndDate != nil {
		histories, err = historyRepo.GetByDateRange(ctx, vehicleID, *filters.StartDate, *filters.EndDate, pagination)
	} else if filters.Search != nil {
		histories, err = historyRepo.SearchByTitle(ctx, vehicleID, *filters.Search, pagination)
	} else {
		histories, err = historyRepo.GetByVehicle(ctx, vehicleID, pagination)
	}
	
	if err != nil {
		return nil, apperrors.NewInternalError("Failed to retrieve vehicle history").WithInternal(err)
	}
	
	// Convert to response format
	var responses []*VehicleHistoryResponse
	for _, history := range histories {
		responses = append(responses, s.historyToResponse(history))
	}
	
	return responses, nil
}

// GetMaintenanceHistory retrieves maintenance history for a vehicle
func (s *VehicleHistoryService) GetMaintenanceHistory(ctx context.Context, companyID, vehicleID string, pagination repository.Pagination) ([]*VehicleHistoryResponse, error) {
	// Validate vehicle belongs to company
	vehicleRepo := s.repoManager.GetVehicles()
	vehicle, err := vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("Vehicle").WithInternal(err)
	}
	
	if vehicle.CompanyID != companyID {
		return nil, apperrors.NewForbiddenError("Vehicle does not belong to this company")
	}
	
	historyRepo := s.repoManager.GetVehicleHistories()
	histories, err := historyRepo.GetMaintenanceHistory(ctx, vehicleID, pagination)
	if err != nil {
		return nil, apperrors.NewInternalError("Failed to retrieve maintenance history").WithInternal(err)
	}
	
	var responses []*VehicleHistoryResponse
	for _, history := range histories {
		responses = append(responses, s.historyToResponse(history))
	}
	
	return responses, nil
}

// GetUpcomingMaintenance retrieves upcoming maintenance for a company
func (s *VehicleHistoryService) GetUpcomingMaintenance(ctx context.Context, companyID string, days int) ([]*VehicleHistoryResponse, error) {
	if days <= 0 {
		days = 30 // Default to 30 days
	}
	
	historyRepo := s.repoManager.GetVehicleHistories()
	histories, err := historyRepo.GetUpcomingMaintenance(ctx, companyID, days)
	if err != nil {
		return nil, apperrors.NewInternalError("Failed to retrieve upcoming maintenance").WithInternal(err)
	}
	
	var responses []*VehicleHistoryResponse
	for _, history := range histories {
		responses = append(responses, s.historyToResponse(history))
	}
	
	return responses, nil
}

// GetOverdueMaintenance retrieves overdue maintenance for a company
func (s *VehicleHistoryService) GetOverdueMaintenance(ctx context.Context, companyID string) ([]*VehicleHistoryResponse, error) {
	historyRepo := s.repoManager.GetVehicleHistories()
	histories, err := historyRepo.GetOverdueMaintenance(ctx, companyID)
	if err != nil {
		return nil, apperrors.NewInternalError("Failed to retrieve overdue maintenance").WithInternal(err)
	}
	
	var responses []*VehicleHistoryResponse
	for _, history := range histories {
		responses = append(responses, s.historyToResponse(history))
	}
	
	return responses, nil
}

// UpdateMaintenanceSchedule updates the maintenance schedule for a history entry
func (s *VehicleHistoryService) UpdateMaintenanceSchedule(ctx context.Context, companyID, historyID string, req MaintenanceScheduleRequest) error {
	// Validate history entry belongs to company
	historyRepo := s.repoManager.GetVehicleHistories()
	history, err := historyRepo.GetByID(ctx, historyID)
	if err != nil {
		return apperrors.NewNotFoundError("History entry").WithInternal(err)
	}
	
	if history.CompanyID != companyID {
		return apperrors.NewForbiddenError("History entry does not belong to this company")
	}
	
	// Update maintenance schedule
	if err := historyRepo.UpdateMaintenanceSchedule(ctx, historyID, *req.NextServiceDue); err != nil {
		return apperrors.NewInternalError("Failed to update maintenance schedule").WithInternal(err)
	}
	
	return nil
}

// GetCostSummary calculates cost summary for a vehicle
func (s *VehicleHistoryService) GetCostSummary(ctx context.Context, companyID, vehicleID string, startDate, endDate time.Time) (*repository.CostSummary, error) {
	// Validate vehicle belongs to company
	vehicleRepo := s.repoManager.GetVehicles()
	vehicle, err := vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("Vehicle").WithInternal(err)
	}
	
	if vehicle.CompanyID != companyID {
		return nil, apperrors.NewForbiddenError("Vehicle does not belong to this company")
	}
	
	historyRepo := s.repoManager.GetVehicleHistories()
	return historyRepo.GetCostSummary(ctx, vehicleID, startDate, endDate)
}

// GetMaintenanceTrends retrieves maintenance trends for a vehicle
func (s *VehicleHistoryService) GetMaintenanceTrends(ctx context.Context, companyID, vehicleID string, months int) ([]*repository.MaintenanceTrend, error) {
	// Validate vehicle belongs to company
	vehicleRepo := s.repoManager.GetVehicles()
	vehicle, err := vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("Vehicle").WithInternal(err)
	}
	
	if vehicle.CompanyID != companyID {
		return nil, apperrors.NewForbiddenError("Vehicle does not belong to this company")
	}
	
	if months <= 0 {
		months = 12 // Default to 12 months
	}
	
	historyRepo := s.repoManager.GetVehicleHistories()
	return historyRepo.GetMaintenanceTrends(ctx, vehicleID, months)
}

// historyToResponse converts a VehicleHistory model to response format
func (s *VehicleHistoryService) historyToResponse(history *models.VehicleHistory) *VehicleHistoryResponse {
	response := &VehicleHistoryResponse{
		ID:              history.ID,
		VehicleID:       history.VehicleID,
		EventType:       history.EventType,
		EventCategory:   history.EventCategory,
		Title:           history.Title,
		Description:     history.Description,
		MileageAtEvent:  history.MileageAtEvent,
		Cost:            history.Cost,
		Currency:        history.Currency,
		FormattedCost:   history.GetFormattedCost(),
		Location:        history.Location,
		ServiceProvider: history.ServiceProvider,
		InvoiceNumber:   history.InvoiceNumber,
		NextServiceDue:  history.NextServiceDue,
		ServiceStatus:   history.GetServiceStatus(),
		CreatedBy:       history.CreatedBy,
		CreatedAt:       history.CreatedAt,
	}
	
	// Parse documents if they exist
	if history.HasDocuments() {
		var docs []map[string]interface{}
		if err := json.Unmarshal(history.Documents, &docs); err == nil {
			response.Documents = docs
		}
		response.DocumentCount = history.GetDocumentCount()
	}
	
	// Get creator name if available
	if history.Creator.ID != "" {
		response.CreatorName = history.Creator.Username
	}
	
	return response
}

// GetVehicleHistoryByID retrieves a specific vehicle history entry
func (s *VehicleHistoryService) GetVehicleHistoryByID(ctx context.Context, companyID, historyID string) (*VehicleHistoryResponse, error) {
	historyRepo := s.repoManager.GetVehicleHistories()
	history, err := historyRepo.GetByID(ctx, historyID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("History entry").WithInternal(err)
	}
	
	if history.CompanyID != companyID {
		return nil, apperrors.NewForbiddenError("History entry does not belong to this company")
	}
	
	return s.historyToResponse(history), nil
}

// UpdateVehicleHistory updates a vehicle history entry
func (s *VehicleHistoryService) UpdateVehicleHistory(ctx context.Context, companyID, historyID, userID string, req AddHistoryRequest) (*VehicleHistoryResponse, error) {
	// Validate history entry belongs to company
	historyRepo := s.repoManager.GetVehicleHistories()
	history, err := historyRepo.GetByID(ctx, historyID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("History entry").WithInternal(err)
	}
	
	if history.CompanyID != companyID {
		return nil, apperrors.NewForbiddenError("History entry does not belong to this company")
	}
	
	// Validate event type and category
	if !models.ValidateEventType(req.EventType) {
		return nil, apperrors.NewValidationError("Invalid event type")
	}
	
	if !models.ValidateEventCategory(req.EventCategory) {
		return nil, apperrors.NewValidationError("Invalid event category")
	}
	
	// Set default currency to IDR if not provided
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	
	// Convert documents to JSON
	var documentsJSON json.RawMessage
	if req.Documents != nil && len(req.Documents) > 0 {
		documentsJSON, err = json.Marshal(req.Documents)
		if err != nil {
			return nil, apperrors.NewInternalError("Failed to marshal documents").WithInternal(err)
		}
	}
	
	// Update history entry
	history.EventType = req.EventType
	history.EventCategory = req.EventCategory
	history.Title = req.Title
	history.Description = req.Description
	history.MileageAtEvent = req.MileageAtEvent
	history.Cost = req.Cost
	history.Currency = req.Currency
	history.Location = req.Location
	history.ServiceProvider = req.ServiceProvider
	history.InvoiceNumber = req.InvoiceNumber
	history.Documents = documentsJSON
	history.NextServiceDue = req.NextServiceDue
	
	if err := historyRepo.Update(ctx, history); err != nil {
		return nil, apperrors.NewInternalError("Failed to update vehicle history").WithInternal(err)
	}
	
	return s.historyToResponse(history), nil
}

// DeleteVehicleHistory deletes a vehicle history entry
func (s *VehicleHistoryService) DeleteVehicleHistory(ctx context.Context, companyID, historyID string) error {
	// Validate history entry belongs to company
	historyRepo := s.repoManager.GetVehicleHistories()
	history, err := historyRepo.GetByID(ctx, historyID)
	if err != nil {
		return apperrors.NewNotFoundError("History entry").WithInternal(err)
	}
	
	if history.CompanyID != companyID {
		return apperrors.NewForbiddenError("History entry does not belong to this company")
	}
	
	if err := historyRepo.Delete(ctx, historyID); err != nil {
		return apperrors.NewInternalError("Failed to delete vehicle history").WithInternal(err)
	}
	
	return nil
}
