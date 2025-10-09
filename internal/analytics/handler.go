package analytics

// Handler handles analytics HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new analytics handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// All handler methods are now in separate files:
// - dashboard_handler.go: GetDashboard, GetRealTimeDashboard
// - fuel_handler.go: GetFuelConsumption, GetFuelEfficiency, GetFuelTheftAlerts, GetFuelOptimization
// - driver_analytics_handler.go: GetDriverPerformance, GetDriverRanking, GetDriverBehavior, GetDriverRecommendations
// - fleet_handler.go: GetFleetUtilization, GetFleetCosts, GetMaintenanceInsights
// - reports_handler.go: GenerateReport, GetComplianceReport, ExportReport
