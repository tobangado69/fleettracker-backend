package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// UserRepositoryImpl implements the UserRepository interface
type UserRepositoryImpl struct {
	*BaseRepository[models.User]
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{
		BaseRepository: NewBaseRepository[models.User](db),
	}
}

// GetByEmail retrieves a user by email address
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found with username: %s", username)
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// GetByCompany retrieves users by company ID with pagination
func (r *UserRepositoryImpl) GetByCompany(ctx context.Context, companyID string, pagination Pagination) ([]*models.User, error) {
	var users []*models.User
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	// Apply pagination
	query = r.applyPagination(query, pagination)
	
	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by company: %w", err)
	}
	
	return users, nil
}

// Search searches users by query string within a company
func (r *UserRepositoryImpl) Search(ctx context.Context, query string, companyID string, pagination Pagination) ([]*models.User, error) {
	var users []*models.User
	dbQuery := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	// Search in multiple fields
	searchPattern := "%" + strings.ToLower(query) + "%"
	dbQuery = dbQuery.Where(
		"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(username) LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	)
	
	// Apply pagination
	dbQuery = r.applyPagination(dbQuery, pagination)
	
	if err := dbQuery.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	
	return users, nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *UserRepositoryImpl) UpdateLastLogin(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", "NOW()").Error; err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// UpdateStatus updates the status of a user
func (r *UserRepositoryImpl) UpdateStatus(ctx context.Context, userID string, status string) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	return nil
}

// GetActiveUsers retrieves all active users for a company
func (r *UserRepositoryImpl) GetActiveUsers(ctx context.Context, companyID string) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).Where("company_id = ? AND is_active = true", companyID).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	return users, nil
}

// GetUsersByRole retrieves users by role within a company
func (r *UserRepositoryImpl) GetUsersByRole(ctx context.Context, companyID string, role string) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).Where("company_id = ? AND role = ?", companyID, role).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}
	return users, nil
}

// GetUsersWithExpiredPasswords retrieves users whose passwords need to be changed
func (r *UserRepositoryImpl) GetUsersWithExpiredPasswords(ctx context.Context, companyID string, daysSinceLastChange int) ([]*models.User, error) {
	var users []*models.User
	query := `
		SELECT * FROM users 
		WHERE company_id = ? 
		AND (password_changed_at IS NULL 
			OR password_changed_at < NOW() - INTERVAL '%d days')
	`
	
	if err := r.db.WithContext(ctx).Raw(fmt.Sprintf(query, daysSinceLastChange), companyID).Scan(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users with expired passwords: %w", err)
	}
	
	return users, nil
}

// GetUsersByLastLogin retrieves users by last login date range
func (r *UserRepositoryImpl) GetUsersByLastLogin(ctx context.Context, companyID string, startDate, endDate string) ([]*models.User, error) {
	var users []*models.User
	query := r.db.WithContext(ctx).Where("company_id = ?", companyID)
	
	if startDate != "" {
		query = query.Where("last_login_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("last_login_at <= ?", endDate)
	}
	
	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by last login: %w", err)
	}
	
	return users, nil
}

// GetUsersWithFailedLoginAttempts retrieves users with recent failed login attempts
func (r *UserRepositoryImpl) GetUsersWithFailedLoginAttempts(ctx context.Context, companyID string, minAttempts int) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).Where("company_id = ? AND failed_login_attempts >= ?", companyID, minAttempts).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users with failed login attempts: %w", err)
	}
	return users, nil
}

// GetUsersNeedingEmailVerification retrieves users who haven't verified their email
func (r *UserRepositoryImpl) GetUsersNeedingEmailVerification(ctx context.Context, companyID string) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).Where("company_id = ? AND is_verified = false", companyID).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users needing email verification: %w", err)
	}
	return users, nil
}

// BulkUpdateStatus updates the status of multiple users
func (r *UserRepositoryImpl) BulkUpdateStatus(ctx context.Context, userIDs []string, status string) error {
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("id IN ?", userIDs).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to bulk update user status: %w", err)
	}
	return nil
}

// GetUserStatistics retrieves user statistics for a company
func (r *UserRepositoryImpl) GetUserStatistics(ctx context.Context, companyID string) (map[string]interface{}, error) {
	var stats struct {
		TotalUsers        int64 `json:"total_users"`
		ActiveUsers       int64 `json:"active_users"`
		InactiveUsers     int64 `json:"inactive_users"`
		VerifiedUsers     int64 `json:"verified_users"`
		UnverifiedUsers   int64 `json:"unverified_users"`
		AdminUsers        int64 `json:"admin_users"`
		ManagerUsers      int64 `json:"manager_users"`
		OperatorUsers     int64 `json:"operator_users"`
	}

	// Get total users
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ?", companyID).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// Get active users
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ? AND is_active = true", companyID).Count(&stats.ActiveUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	// Get inactive users
	stats.InactiveUsers = stats.TotalUsers - stats.ActiveUsers

	// Get verified users
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ? AND is_verified = true", companyID).Count(&stats.VerifiedUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count verified users: %w", err)
	}

	// Get unverified users
	stats.UnverifiedUsers = stats.TotalUsers - stats.VerifiedUsers

	// Get users by role
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ? AND role = ?", companyID, "admin").Count(&stats.AdminUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count admin users: %w", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ? AND role = ?", companyID, "manager").Count(&stats.ManagerUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count manager users: %w", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ? AND role = ?", companyID, "operator").Count(&stats.OperatorUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count operator users: %w", err)
	}

	return map[string]interface{}{
		"total_users":        stats.TotalUsers,
		"active_users":       stats.ActiveUsers,
		"inactive_users":     stats.InactiveUsers,
		"verified_users":     stats.VerifiedUsers,
		"unverified_users":   stats.UnverifiedUsers,
		"admin_users":        stats.AdminUsers,
		"manager_users":      stats.ManagerUsers,
		"operator_users":     stats.OperatorUsers,
	}, nil
}

// GetUserActivitySummary retrieves user activity summary for a company
func (r *UserRepositoryImpl) GetUserActivitySummary(ctx context.Context, companyID string, days int) (map[string]interface{}, error) {
	var summary struct {
		UsersLoggedInToday    int64 `json:"users_logged_in_today"`
		UsersLoggedInThisWeek int64 `json:"users_logged_in_this_week"`
		UsersLoggedInThisMonth int64 `json:"users_logged_in_this_month"`
		NeverLoggedIn         int64 `json:"never_logged_in"`
	}

	// Users logged in today
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ? AND DATE(last_login_at) = CURRENT_DATE", companyID).Count(&summary.UsersLoggedInToday).Error; err != nil {
		return nil, fmt.Errorf("failed to count users logged in today: %w", err)
	}

	// Users logged in this week
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ? AND last_login_at >= NOW() - INTERVAL '7 days'", companyID).Count(&summary.UsersLoggedInThisWeek).Error; err != nil {
		return nil, fmt.Errorf("failed to count users logged in this week: %w", err)
	}

	// Users logged in this month
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ? AND last_login_at >= NOW() - INTERVAL '30 days'", companyID).Count(&summary.UsersLoggedInThisMonth).Error; err != nil {
		return nil, fmt.Errorf("failed to count users logged in this month: %w", err)
	}

	// Users who never logged in
	if err := r.db.WithContext(ctx).Model(&models.User{}).Where("company_id = ? AND last_login_at IS NULL", companyID).Count(&summary.NeverLoggedIn).Error; err != nil {
		return nil, fmt.Errorf("failed to count users who never logged in: %w", err)
	}

	return map[string]interface{}{
		"users_logged_in_today":     summary.UsersLoggedInToday,
		"users_logged_in_this_week": summary.UsersLoggedInThisWeek,
		"users_logged_in_this_month": summary.UsersLoggedInThisMonth,
		"never_logged_in":           summary.NeverLoggedIn,
	}, nil
}
