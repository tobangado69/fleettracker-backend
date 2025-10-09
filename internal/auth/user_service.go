package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	apperrors "github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// toUserResponse converts models.User to UserResponse
func toUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Phone:       user.Phone,
		Role:        user.Role,
		CompanyID:   user.CompanyID,
		IsActive:    user.IsActive,
		IsVerified:  user.IsVerified,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
	}
}

// CreateUserRequest represents a request to create a new user (admin-only)
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	Role      string `json:"role" binding:"required"`
	CompanyID string `json:"company_id"` // For super-admin creating users in other companies
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
}

// ChangeRoleRequest represents a request to change user role
type ChangeRoleRequest struct {
	NewRole string `json:"new_role" binding:"required"`
}

// CreateUser creates a new user (admin-only)
func (s *Service) CreateUser(ctx context.Context, creatorUserID, creatorRole, creatorCompanyID string, req *CreateUserRequest) (*models.User, *apperrors.AppError) {
	// Validate role is valid
	if !IsValidRole(req.Role) {
		return nil, apperrors.NewValidationError(fmt.Sprintf("invalid role: %s", req.Role))
	}

	// Check if creator can create users
	if !CanManageUsers(creatorRole) {
		return nil, apperrors.NewForbiddenError("You do not have permission to create users")
	}

	// Validate role creation permission
	if err := ValidateRoleCreation(creatorRole, req.Role); err != nil {
		return nil, apperrors.NewForbiddenError(err.Error())
	}

	// Determine company ID
	companyID := creatorCompanyID
	if req.CompanyID != "" {
		// Only super-admin can create users in other companies
		if creatorRole != RoleSuperAdmin {
			return nil, apperrors.NewForbiddenError("Only super-admin can create users in other companies")
		}
		companyID = req.CompanyID
	}

	// Verify company exists
	var company models.Company
	if err := s.db.First(&company, "id = ?", companyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError(fmt.Sprintf("company %s not found", companyID))
		}
		return nil, apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	// Check if email already exists
	var existingUser models.User
	err := s.db.Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		return nil, apperrors.NewConflictError(fmt.Sprintf("email %s already exists", req.Email))
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, apperrors.NewInternalError(fmt.Sprintf("failed to hash password: %v", err))
	}

	// Create user
	user := &models.User{
		CompanyID: companyID,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      req.Role,
		IsActive:  true,
		Status:    "active",
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	// Clear password before returning
	user.Password = ""

	return user, nil
}

// GetUsers lists all users in the company (admin-only)
func (s *Service) GetUsers(ctx context.Context, userRole, companyID string, page, limit int) ([]models.User, int64, *apperrors.AppError) {
	// Check permission
	if !CanManageUsers(userRole) {
		return nil, 0, apperrors.NewForbiddenError("You do not have permission to list users")
	}

	var users []models.User
	var total int64

	query := s.db.Model(&models.User{})

	// Company isolation (except super-admin)
	if userRole != RoleSuperAdmin {
		query = query.Where("company_id = ?", companyID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	// Clear passwords
	for i := range users {
		users[i].Password = ""
	}

	return users, total, nil
}

// GetUser gets a single user by ID (admin-only)
func (s *Service) GetUser(ctx context.Context, userRole, companyID, targetUserID string) (*models.User, *apperrors.AppError) {
	// Check permission
	if !CanManageUsers(userRole) {
		return nil, apperrors.NewForbiddenError("You do not have permission to view user details")
	}

	var user models.User
	query := s.db.Where("id = ?", targetUserID)

	// Company isolation (except super-admin)
	if userRole != RoleSuperAdmin {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError(fmt.Sprintf("user %s not found", targetUserID))
		}
		return nil, apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	user.Password = ""
	return &user, nil
}

// UpdateUser updates a user (admin-only)
func (s *Service) UpdateUser(ctx context.Context, userRole, companyID, targetUserID string, req *UpdateUserRequest) (*models.User, *apperrors.AppError) {
	// Check permission
	if !CanManageUsers(userRole) {
		return nil, apperrors.NewForbiddenError("You do not have permission to update users")
	}

	// Get existing user
	var user models.User
	query := s.db.Where("id = ?", targetUserID)

	// Company isolation
	if userRole != RoleSuperAdmin {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError(fmt.Sprintf("user %s not found", targetUserID))
		}
		return nil, apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Email != nil {
		// Check if email already exists
		var existingUser models.User
		err := s.db.Where("email = ? AND id != ?", *req.Email, targetUserID).First(&existingUser).Error
		if err == nil {
			return nil, apperrors.NewConflictError(fmt.Sprintf("email %s already exists", *req.Email))
		}
		updates["email"] = *req.Email
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	user.Password = ""
	return &user, nil
}

// DeactivateUser deactivates a user (owner/super-admin only)
func (s *Service) DeactivateUser(ctx context.Context, userRole, companyID, targetUserID string) *apperrors.AppError {
	// Only owner and super-admin can deactivate users
	if userRole != RoleSuperAdmin && userRole != RoleOwner {
		return apperrors.NewForbiddenError("Only super-admin or owner can deactivate users")
	}

	// Get user
	var user models.User
	query := s.db.Where("id = ?", targetUserID)

	// Company isolation
	if userRole != RoleSuperAdmin {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.NewNotFoundError(fmt.Sprintf("user %s not found", targetUserID))
		}
		return apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	// Prevent deactivating yourself
	if user.ID == targetUserID {
		return apperrors.NewBadRequestError("Cannot deactivate your own account")
	}

	// Deactivate user
	updates := map[string]interface{}{
		"is_active": false,
		"status":    "inactive",
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	// Invalidate all user sessions
	pattern := fmt.Sprintf("session:user:%s:*", targetUserID)
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err == nil {
		if len(keys) > 0 {
			s.redis.Del(ctx, keys...)
		}
	}

	return nil
}

// ChangeUserRole changes a user's role (admin-only)
func (s *Service) ChangeUserRole(ctx context.Context, changerRole, changerCompanyID, targetUserID, newRole string) *apperrors.AppError {
	// Validate new role
	if !IsValidRole(newRole) {
		return apperrors.NewValidationError(fmt.Sprintf("invalid role: %s", newRole))
	}

	// Check permission
	if !CanManageUsers(changerRole) {
		return apperrors.NewForbiddenError("You do not have permission to change user roles")
	}

	// Get target user
	var user models.User
	query := s.db.Where("id = ?", targetUserID)

	// Company isolation
	if changerRole != RoleSuperAdmin {
		query = query.Where("company_id = ?", changerCompanyID)
	}

	if err := query.First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return apperrors.NewNotFoundError(fmt.Sprintf("user %s not found", targetUserID))
		}
		return apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	// Validate role assignment
	if err := ValidateRoleAssignment(changerRole, user.Role, newRole); err != nil {
		return apperrors.NewForbiddenError(err.Error())
	}

	// Update role
	if err := s.db.Model(&user).Update("role", newRole).Error; err != nil {
		return apperrors.NewInternalError(err.Error()).WithInternal(err)
	}

	return nil
}

// IsFirstUser checks if this is the first user in the system
func (s *Service) IsFirstUser() (bool, error) {
	var count int64
	if err := s.db.Model(&models.User{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}

// IsFirstUserInCompany checks if this is the first user in a company
func (s *Service) IsFirstUserInCompany(companyID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.User{}).Where("company_id = ?", companyID).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}

