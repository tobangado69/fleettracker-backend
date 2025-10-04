package repository

import (
	"context"
	"time"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// Repository defines the base repository interface for CRUD operations
type Repository[T any] interface {
	// Basic CRUD operations
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id string) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id string) error
	
	// Query operations
	List(ctx context.Context, filters FilterOptions, pagination Pagination) ([]*T, error)
	Count(ctx context.Context, filters FilterOptions) (int64, error)
	
	// Transaction support
	WithTransaction(ctx context.Context, fn func(Repository[T]) error) error
}

// FilterOptions represents filtering options for queries
type FilterOptions struct {
	// Basic filters
	Where     map[string]interface{} `json:"where"`
	WhereIn   map[string][]interface{} `json:"where_in"`
	WhereNot  map[string]interface{} `json:"where_not"`
	WhereLike map[string]string      `json:"where_like"`
	
	// Date range filters
	DateRange map[string]DateRange `json:"date_range"`
	
	// Text search
	Search     string   `json:"search"`
	SearchIn   []string `json:"search_in"`
	
	// Company isolation (for multi-tenancy)
	CompanyID string `json:"company_id"`
	
	// Additional conditions
	Conditions []Condition `json:"conditions"`
}

// Condition represents a custom query condition
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // =, !=, >, <, >=, <=, IN, NOT IN, LIKE, ILIKE
	Value    interface{} `json:"value"`
}

// DateRange represents a date range filter
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// Pagination represents pagination options
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
	Limit    int `json:"limit"`
}

// SortOptions represents sorting options
type SortOptions struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // ASC, DESC
}

// QueryOptions combines all query options
type QueryOptions struct {
	Filters    FilterOptions `json:"filters"`
	Pagination Pagination    `json:"pagination"`
	Sort       []SortOptions `json:"sort"`
}

// RepositoryResult represents the result of a repository operation
type RepositoryResult[T any] struct {
	Data       []*T                `json:"data"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
	HasMore    bool                `json:"has_more"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Transaction represents a database transaction
type Transaction interface {
	Commit() error
	Rollback() error
	GetRepository() *BaseRepository[interface{}]
}


// Entity-specific repository interfaces

// UserRepository defines user-specific repository operations
type UserRepository interface {
	Repository[models.User]
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.User, error)
	Search(ctx context.Context, query string, companyID string, pagination Pagination) ([]*models.User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	UpdateStatus(ctx context.Context, userID string, status string) error
	GetActiveUsers(ctx context.Context, companyID string) ([]*models.User, error)
	GetUsersByRole(ctx context.Context, companyID string, role string) ([]*models.User, error)
}

// VehicleRepository defines vehicle-specific repository operations
type VehicleRepository interface {
	Repository[models.Vehicle]
	GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Vehicle, error)
	GetByDriver(ctx context.Context, driverID string) (*models.Vehicle, error)
	GetByStatus(ctx context.Context, companyID string, status string) ([]*models.Vehicle, error)
	GetByType(ctx context.Context, companyID string, vehicleType string) ([]*models.Vehicle, error)
	SearchByLicensePlate(ctx context.Context, licensePlate string, companyID string) ([]*models.Vehicle, error)
	SearchByVIN(ctx context.Context, vin string, companyID string) ([]*models.Vehicle, error)
	GetAvailableVehicles(ctx context.Context, companyID string) ([]*models.Vehicle, error)
	GetVehiclesNeedingInspection(ctx context.Context, companyID string) ([]*models.Vehicle, error)
	UpdateStatus(ctx context.Context, vehicleID string, status string) error
	AssignDriver(ctx context.Context, vehicleID string, driverID *string) error
	UpdateOdometer(ctx context.Context, vehicleID string, odometer float64) error
}

// DriverRepository defines driver-specific repository operations
type DriverRepository interface {
	Repository[models.Driver]
	GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Driver, error)
	GetByVehicle(ctx context.Context, vehicleID string) (*models.Driver, error)
	GetByStatus(ctx context.Context, companyID string, status string) ([]*models.Driver, error)
	GetBySIMType(ctx context.Context, companyID string, simType string) ([]*models.Driver, error)
	SearchByNIK(ctx context.Context, nik string, companyID string) ([]*models.Driver, error)
	SearchBySIM(ctx context.Context, simNumber string, companyID string) ([]*models.Driver, error)
	GetAvailableDrivers(ctx context.Context, companyID string) ([]*models.Driver, error)
	GetDriversNeedingTraining(ctx context.Context, companyID string) ([]*models.Driver, error)
	GetDriversWithExpiredSIM(ctx context.Context, companyID string) ([]*models.Driver, error)
	GetDriversWithExpiredMedicalCheckup(ctx context.Context, companyID string) ([]*models.Driver, error)
	UpdateStatus(ctx context.Context, driverID string, status string) error
	UpdatePerformance(ctx context.Context, driverID string, performance models.PerformanceLog) error
	AssignVehicle(ctx context.Context, driverID string, vehicleID *string) error
}

// GPSTrackRepository defines GPS tracking-specific repository operations
type GPSTrackRepository interface {
	Repository[models.GPSTrack]
	GetByVehicle(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.GPSTrack, error)
	GetByVehicleAndDateRange(ctx context.Context, vehicleID string, startDate, endDate string) ([]*models.GPSTrack, error)
	GetCurrentLocation(ctx context.Context, vehicleID string) (*models.GPSTrack, error)
	GetLocationHistory(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.GPSTrack, error)
	GetByDriver(ctx context.Context, driverID string, pagination Pagination) ([]*models.GPSTrack, error)
	GetByTrip(ctx context.Context, tripID string) ([]*models.GPSTrack, error)
	GetSpeedViolations(ctx context.Context, vehicleID string, minSpeed float64) ([]*models.GPSTrack, error)
	GetRecentTracks(ctx context.Context, companyID string, limit int) ([]*models.GPSTrack, error)
	GetTracksInGeofence(ctx context.Context, geofenceID string, pagination Pagination) ([]*models.GPSTrack, error)
	AggregateByTimeRange(ctx context.Context, vehicleID string, startDate, endDate string) (map[string]interface{}, error)
}

// TripRepository defines trip-specific repository operations
type TripRepository interface {
	Repository[models.Trip]
	GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Trip, error)
	GetByVehicle(ctx context.Context, vehicleID string, pagination Pagination) ([]*models.Trip, error)
	GetByDriver(ctx context.Context, driverID string, pagination Pagination) ([]*models.Trip, error)
	GetByStatus(ctx context.Context, companyID string, status string) ([]*models.Trip, error)
	GetByDateRange(ctx context.Context, companyID string, startDate, endDate string) ([]*models.Trip, error)
	GetActiveTrips(ctx context.Context, companyID string) ([]*models.Trip, error)
	GetCompletedTrips(ctx context.Context, companyID string, pagination Pagination) ([]*models.Trip, error)
	StartTrip(ctx context.Context, trip *models.Trip) error
	EndTrip(ctx context.Context, tripID string, endData map[string]interface{}) error
	GetTripStatistics(ctx context.Context, companyID string, dateRange DateRange) (map[string]interface{}, error)
}

// GeofenceRepository defines geofence-specific repository operations
type GeofenceRepository interface {
	Repository[models.Geofence]
	GetByCompany(ctx context.Context, companyID string) ([]*models.Geofence, error)
	GetActive(ctx context.Context, companyID string) ([]*models.Geofence, error)
	GetByType(ctx context.Context, companyID string, geofenceType string) ([]*models.Geofence, error)
	GetGeofencesNearLocation(ctx context.Context, latitude, longitude float64, radius float64, companyID string) ([]*models.Geofence, error)
	CheckLocationInGeofences(ctx context.Context, latitude, longitude float64, companyID string) ([]*models.Geofence, error)
}

// CompanyRepository defines company-specific repository operations
type CompanyRepository interface {
	Repository[models.Company]
	GetByNPWP(ctx context.Context, npwp string) (*models.Company, error)
	GetByEmail(ctx context.Context, email string) (*models.Company, error)
	GetActiveCompanies(ctx context.Context) ([]*models.Company, error)
	UpdateStatus(ctx context.Context, companyID string, status string) error
	GetCompanyStatistics(ctx context.Context, companyID string) (map[string]interface{}, error)
}

// AuditLogRepository defines audit log-specific repository operations
type AuditLogRepository interface {
	Repository[models.AuditLog]
	GetByUser(ctx context.Context, userID string, pagination Pagination) ([]*models.AuditLog, error)
	GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.AuditLog, error)
	GetByAction(ctx context.Context, action string, pagination Pagination) ([]*models.AuditLog, error)
	GetByResource(ctx context.Context, resource string, resourceID string) ([]*models.AuditLog, error)
	GetByDateRange(ctx context.Context, startDate, endDate string, pagination Pagination) ([]*models.AuditLog, error)
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
}

// SessionRepository defines session-specific repository operations
type SessionRepository interface {
	Repository[models.Session]
	GetByUser(ctx context.Context, userID string) ([]*models.Session, error)
	GetByToken(ctx context.Context, token string) (*models.Session, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)
	GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error)
	DeactivateSession(ctx context.Context, sessionID string) error
	DeactivateUserSessions(ctx context.Context, userID string) error
	CleanupExpiredSessions(ctx context.Context) error
}

// PasswordResetTokenRepository defines password reset token-specific repository operations
type PasswordResetTokenRepository interface {
	Repository[models.PasswordResetToken]
	GetByToken(ctx context.Context, token string) (*models.PasswordResetToken, error)
	GetByUser(ctx context.Context, userID string) ([]*models.PasswordResetToken, error)
	GetValidToken(ctx context.Context, token string) (*models.PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, tokenID string) error
	CleanupExpiredTokens(ctx context.Context) error
}

// InvoiceRepository defines invoice-specific repository operations
type InvoiceRepository interface {
	Repository[models.Invoice]
	GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Invoice, error)
	GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error)
	GetByStatus(ctx context.Context, status string, pagination Pagination) ([]*models.Invoice, error)
	GetOverdueInvoices(ctx context.Context) ([]*models.Invoice, error)
	GetByDueDateRange(ctx context.Context, startDate, endDate time.Time, pagination Pagination) ([]*models.Invoice, error)
	UpdatePaymentStatus(ctx context.Context, invoiceID string, status string, paidAmount float64) error
}

// PaymentRepository defines payment-specific repository operations
type PaymentRepository interface {
	Repository[models.Payment]
	GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Payment, error)
	GetBySubscription(ctx context.Context, subscriptionID string, pagination Pagination) ([]*models.Payment, error)
	GetByStatus(ctx context.Context, status string, pagination Pagination) ([]*models.Payment, error)
	GetByPaymentMethod(ctx context.Context, paymentMethod string, pagination Pagination) ([]*models.Payment, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time, pagination Pagination) ([]*models.Payment, error)
	GetByReferenceNumber(ctx context.Context, referenceNumber string) (*models.Payment, error)
}

// SubscriptionRepository defines subscription-specific repository operations
type SubscriptionRepository interface {
	Repository[models.Subscription]
	GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.Subscription, error)
	GetActiveSubscriptions(ctx context.Context) ([]*models.Subscription, error)
	GetExpiringSubscriptions(ctx context.Context, days int) ([]*models.Subscription, error)
	GetByStatus(ctx context.Context, status string, pagination Pagination) ([]*models.Subscription, error)
	GetByPlanType(ctx context.Context, planType string, pagination Pagination) ([]*models.Subscription, error)
	UpdateStatus(ctx context.Context, subscriptionID string, status string) error
	GetCompanyActiveSubscription(ctx context.Context, companyID string) (*models.Subscription, error)
}
