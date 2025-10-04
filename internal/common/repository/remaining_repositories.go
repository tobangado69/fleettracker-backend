package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// GeofenceRepositoryImpl implements the GeofenceRepository interface
type GeofenceRepositoryImpl struct {
	*BaseRepository[models.Geofence]
}

// NewGeofenceRepository creates a new geofence repository
func NewGeofenceRepository(db *gorm.DB) GeofenceRepository {
	return &GeofenceRepositoryImpl{
		BaseRepository: NewBaseRepository[models.Geofence](db),
	}
}

// GetByCompany retrieves geofences by company ID
func (r *GeofenceRepositoryImpl) GetByCompany(ctx context.Context, companyID string) ([]*models.Geofence, error) {
	var geofences []*models.Geofence
	if err := r.db.WithContext(ctx).Where("company_id = ?", companyID).Find(&geofences).Error; err != nil {
		return nil, fmt.Errorf("failed to get geofences by company: %w", err)
	}
	return geofences, nil
}

// GetActive retrieves active geofences for a company
func (r *GeofenceRepositoryImpl) GetActive(ctx context.Context, companyID string) ([]*models.Geofence, error) {
	var geofences []*models.Geofence
	if err := r.db.WithContext(ctx).Where("company_id = ? AND is_active = true", companyID).Find(&geofences).Error; err != nil {
		return nil, fmt.Errorf("failed to get active geofences: %w", err)
	}
	return geofences, nil
}

// GetByType retrieves geofences by type within a company
func (r *GeofenceRepositoryImpl) GetByType(ctx context.Context, companyID string, geofenceType string) ([]*models.Geofence, error) {
	var geofences []*models.Geofence
	if err := r.db.WithContext(ctx).Where("company_id = ? AND type = ?", companyID, geofenceType).Find(&geofences).Error; err != nil {
		return nil, fmt.Errorf("failed to get geofences by type: %w", err)
	}
	return geofences, nil
}

// GetGeofencesNearLocation retrieves geofences near a specific location
func (r *GeofenceRepositoryImpl) GetGeofencesNearLocation(ctx context.Context, latitude, longitude, radius float64, companyID string) ([]*models.Geofence, error) {
	var geofences []*models.Geofence
	// Basic bounding box approach (would need PostGIS for proper spatial queries)
	latRange := radius / 111.0
	lngRange := radius / (111.0 * cos(latitude))
	
	if err := r.db.WithContext(ctx).Where("company_id = ? AND latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?", 
		companyID, latitude-latRange, latitude+latRange, longitude-lngRange, longitude+lngRange).Find(&geofences).Error; err != nil {
		return nil, fmt.Errorf("failed to get geofences near location: %w", err)
	}
	return geofences, nil
}

// CheckLocationInGeofences checks if a location is within any geofences
func (r *GeofenceRepositoryImpl) CheckLocationInGeofences(ctx context.Context, latitude, longitude float64, companyID string) ([]*models.Geofence, error) {
	var geofences []*models.Geofence
	// This would require PostGIS for proper spatial queries
	// For now, we'll return empty slice as this needs spatial database support
	if err := r.db.WithContext(ctx).Where("company_id = ?", companyID).Find(&geofences).Error; err != nil {
		return nil, fmt.Errorf("failed to check location in geofences: %w", err)
	}
	return geofences, nil
}

// CompanyRepositoryImpl implements the CompanyRepository interface
type CompanyRepositoryImpl struct {
	*BaseRepository[models.Company]
}

// NewCompanyRepository creates a new company repository
func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &CompanyRepositoryImpl{
		BaseRepository: NewBaseRepository[models.Company](db),
	}
}

// GetByNPWP retrieves a company by NPWP number
func (r *CompanyRepositoryImpl) GetByNPWP(ctx context.Context, npwp string) (*models.Company, error) {
	var company models.Company
	if err := r.db.WithContext(ctx).Where("npwp = ?", npwp).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("company not found with NPWP: %s", npwp)
		}
		return nil, fmt.Errorf("failed to get company by NPWP: %w", err)
	}
	return &company, nil
}

// GetByEmail retrieves a company by email
func (r *CompanyRepositoryImpl) GetByEmail(ctx context.Context, email string) (*models.Company, error) {
	var company models.Company
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("company not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to get company by email: %w", err)
	}
	return &company, nil
}

// GetActiveCompanies retrieves all active companies
func (r *CompanyRepositoryImpl) GetActiveCompanies(ctx context.Context) ([]*models.Company, error) {
	var companies []*models.Company
	if err := r.db.WithContext(ctx).Where("is_active = true").Find(&companies).Error; err != nil {
		return nil, fmt.Errorf("failed to get active companies: %w", err)
	}
	return companies, nil
}

// UpdateStatus updates the status of a company
func (r *CompanyRepositoryImpl) UpdateStatus(ctx context.Context, companyID string, status string) error {
	if err := r.db.WithContext(ctx).Model(&models.Company{}).Where("id = ?", companyID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update company status: %w", err)
	}
	return nil
}

// GetCompanyStatistics retrieves company statistics
func (r *CompanyRepositoryImpl) GetCompanyStatistics(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats struct {
		TotalUsers    int64 `json:"total_users"`
		TotalVehicles int64 `json:"total_vehicles"`
		TotalDrivers  int64 `json:"total_drivers"`
		ActiveTrips   int64 `json:"active_trips"`
	}

	// Get total users
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ?", companyID).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// Get total vehicles
	if err := r.db.WithContext(ctx).Model(&models.Vehicle{}).Where("company_id = ?", companyID).Count(&stats.TotalVehicles).Error; err != nil {
		return nil, fmt.Errorf("failed to count total vehicles: %w", err)
	}

	// Get total drivers
	if err := r.db.WithContext(ctx).Model(&models.Driver{}).Where("company_id = ?", companyID).Count(&stats.TotalDrivers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total drivers: %w", err)
	}

	// Get active trips
	if err := r.db.WithContext(ctx).Model(&models.Trip{}).Where("company_id = ? AND status = ?", companyID, "active").Count(&stats.ActiveTrips).Error; err != nil {
		return nil, fmt.Errorf("failed to count active trips: %w", err)
	}

	return map[string]interface{}{
		"total_users":    stats.TotalUsers,
		"total_vehicles": stats.TotalVehicles,
		"total_drivers":  stats.TotalDrivers,
		"active_trips":   stats.ActiveTrips,
	}, nil
}

// AuditLogRepositoryImpl implements the AuditLogRepository interface
type AuditLogRepositoryImpl struct {
	*BaseRepository[models.AuditLog]
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &AuditLogRepositoryImpl{
		BaseRepository: NewBaseRepository[models.AuditLog](db),
	}
}

// GetByUser retrieves audit logs by user ID with pagination
func (r *AuditLogRepositoryImpl) GetByUser(ctx context.Context, userID string, pagination Pagination) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs by user: %w", err)
	}
	
	return logs, nil
}

// GetByCompany retrieves audit logs by company ID with pagination
func (r *AuditLogRepositoryImpl) GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs by company: %w", err)
	}
	
	return logs, nil
}

// GetByAction retrieves audit logs by action with pagination
func (r *AuditLogRepositoryImpl) GetByAction(ctx context.Context, action string, pagination Pagination) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := r.db.WithContext(ctx).Where("action = ?", action).Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs by action: %w", err)
	}
	
	return logs, nil
}

// GetByResource retrieves audit logs by resource and resource ID
func (r *AuditLogRepositoryImpl) GetByResource(ctx context.Context, resource string, resourceID string) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	if err := r.db.WithContext(ctx).Where("resource = ? AND resource_id = ?", resource, resourceID).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs by resource: %w", err)
	}
	return logs, nil
}

// GetByDateRange retrieves audit logs by date range with pagination
func (r *AuditLogRepositoryImpl) GetByDateRange(ctx context.Context, startDate, endDate string, pagination Pagination) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := r.db.WithContext(ctx)
	
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}
	
	query = query.Order("created_at DESC")
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs by date range: %w", err)
	}
	
	return logs, nil
}

// CreateAuditLog creates a new audit log entry
func (r *AuditLogRepositoryImpl) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// SessionRepositoryImpl implements the SessionRepository interface
type SessionRepositoryImpl struct {
	*BaseRepository[models.Session]
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &SessionRepositoryImpl{
		BaseRepository: NewBaseRepository[models.Session](db),
	}
}

// GetByUser retrieves sessions by user ID
func (r *SessionRepositoryImpl) GetByUser(ctx context.Context, userID string) ([]*models.Session, error) {
	var sessions []*models.Session
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get sessions by user: %w", err)
	}
	return sessions, nil
}

// GetByToken retrieves a session by token
func (r *SessionRepositoryImpl) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found with token")
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}
	return &session, nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *SessionRepositoryImpl) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	var session models.Session
	if err := r.db.WithContext(ctx).Where("refresh_token = ?", refreshToken).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found with refresh token")
		}
		return nil, fmt.Errorf("failed to get session by refresh token: %w", err)
	}
	return &session, nil
}

// GetActiveSessions retrieves active sessions for a user
func (r *SessionRepositoryImpl) GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	var sessions []*models.Session
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = true AND expires_at > NOW()", userID).Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	return sessions, nil
}

// DeactivateSession deactivates a session
func (r *SessionRepositoryImpl) DeactivateSession(ctx context.Context, sessionID string) error {
	if err := r.db.WithContext(ctx).Model(&models.Session{}).Where("id = ?", sessionID).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate session: %w", err)
	}
	return nil
}

// DeactivateUserSessions deactivates all sessions for a user
func (r *SessionRepositoryImpl) DeactivateUserSessions(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Model(&models.Session{}).Where("user_id = ?", userID).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate user sessions: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *SessionRepositoryImpl) CleanupExpiredSessions(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("expires_at < NOW()").Delete(&models.Session{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

// PasswordResetTokenRepositoryImpl implements the PasswordResetTokenRepository interface
type PasswordResetTokenRepositoryImpl struct {
	*BaseRepository[models.PasswordResetToken]
}

// NewPasswordResetTokenRepository creates a new password reset token repository
func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &PasswordResetTokenRepositoryImpl{
		BaseRepository: NewBaseRepository[models.PasswordResetToken](db),
	}
}

// GetByToken retrieves a password reset token by token value
func (r *PasswordResetTokenRepositoryImpl) GetByToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&resetToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("password reset token not found")
		}
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}
	return &resetToken, nil
}

// GetByUser retrieves password reset tokens by user ID
func (r *PasswordResetTokenRepositoryImpl) GetByUser(ctx context.Context, userID string) ([]*models.PasswordResetToken, error) {
	var tokens []*models.PasswordResetToken
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to get password reset tokens by user: %w", err)
	}
	return tokens, nil
}

// GetValidToken retrieves a valid password reset token
func (r *PasswordResetTokenRepositoryImpl) GetValidToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	if err := r.db.WithContext(ctx).Where("token = ? AND expires_at > NOW() AND used_at IS NULL", token).First(&resetToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("valid password reset token not found")
		}
		return nil, fmt.Errorf("failed to get valid password reset token: %w", err)
	}
	return &resetToken, nil
}

// MarkAsUsed marks a password reset token as used
func (r *PasswordResetTokenRepositoryImpl) MarkAsUsed(ctx context.Context, tokenID string) error {
	if err := r.db.WithContext(ctx).Model(&models.PasswordResetToken{}).Where("id = ?", tokenID).Update("used_at", "NOW()").Error; err != nil {
		return fmt.Errorf("failed to mark password reset token as used: %w", err)
	}
	return nil
}

// CleanupExpiredTokens removes expired password reset tokens
func (r *PasswordResetTokenRepositoryImpl) CleanupExpiredTokens(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("expires_at < NOW()").Delete(&models.PasswordResetToken{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired password reset tokens: %w", err)
	}
	return nil
}

// Helper function for cosine calculation (used in geofence queries)
func cos(x float64) float64 {
	// Simple cosine approximation for small angles
	return 1.0 - (x*x)/2.0 + (x*x*x*x)/24.0
}
