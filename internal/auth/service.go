package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"github.com/tobangado69/fleettracker-pro/backend/pkg/errors"
	"github.com/tobangado69/fleettracker-pro/backend/pkg/models"
)

// Service handles authentication operations
type Service struct {
	db        *gorm.DB
	redis     *redis.Client
	jwtSecret []byte
	cache     *CacheService
}

// CacheService provides caching functionality for auth operations
type CacheService struct {
	redis *redis.Client
}

// NewCacheService creates a new cache service
func NewCacheService(redis *redis.Client) *CacheService {
	return &CacheService{redis: redis}
}

// GetUserFromCache retrieves a user from cache
func (cs *CacheService) GetUserFromCache(ctx context.Context, userID string) (*models.User, error) {
	key := fmt.Sprintf("user:%s", userID)
	
	var user models.User
	data, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get user from cache: %w", err)
	}
	
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user from cache: %w", err)
	}
	
	return &user, nil
}

// SetUserInCache stores a user in cache
func (cs *CacheService) SetUserInCache(ctx context.Context, user *models.User, expiration time.Duration) error {
	key := fmt.Sprintf("user:%s", user.ID)
	
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user for cache: %w", err)
	}
	
	if err := cs.redis.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set user in cache: %w", err)
	}
	
	return nil
}

// InvalidateUserCache removes a user from cache
func (cs *CacheService) InvalidateUserCache(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:%s", userID)
	
	if err := cs.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate user cache: %w", err)
	}
	
	return nil
}

// SetSessionInCache stores a user session in cache
func (cs *CacheService) SetSessionInCache(ctx context.Context, sessionID string, userID string, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	
	if err := cs.redis.Set(ctx, key, userID, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set session in cache: %w", err)
	}
	
	return nil
}

// GetSessionFromCache retrieves a user session from cache
func (cs *CacheService) GetSessionFromCache(ctx context.Context, sessionID string) (string, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	
	userID, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Cache miss
		}
		return "", fmt.Errorf("failed to get session from cache: %w", err)
	}
	
	return userID, nil
}

// InvalidateSessionCache removes a session from cache
func (cs *CacheService) InvalidateSessionCache(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	
	if err := cs.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate session cache: %w", err)
	}
	
	return nil
}

// Claims represents JWT claims
type Claims struct {
	UserID    string `json:"user_id"`
	CompanyID string `json:"company_id"`
	Role      string `json:"role"`
	Username  string `json:"username"`
	jwt.RegisteredClaims
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	CompanyID string `json:"company_id" binding:"required"`
	Role      string `json:"role"` // Optional, defaults to operator
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse represents JWT token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// UserResponse represents user response data
type UserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Phone       string    `json:"phone"`
	Role        string    `json:"role"`
	CompanyID   string    `json:"company_id"`
	IsActive    bool      `json:"is_active"`
	IsVerified  bool      `json:"is_verified"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewService creates a new authentication service
func NewService(db *gorm.DB, redis *redis.Client, jwtSecret string) *Service {
	return &Service{
		db:        db,
		redis:     redis,
		jwtSecret: []byte(jwtSecret),
		cache:     NewCacheService(redis),
	}
}

// Register creates a new user account
func (s *Service) Register(req RegisterRequest) (*UserResponse, error) {
	// Validate email uniqueness
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.NewConflictError("Email already registered")
	}

	// Validate username uniqueness
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.NewConflictError("Username already taken")
	}

	// Validate company exists
	var company models.Company
	if err := s.db.Where("id = ? AND is_active = true", req.CompanyID).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("Company")
		}
		return nil, errors.NewInternalError("Failed to validate company").WithInternal(err)
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "operator"
	}

	// Validate role
	validRoles := []string{"admin", "manager", "operator"}
	if !contains(validRoles, role) {
		return nil, errors.NewValidationError(fmt.Sprintf("Invalid role: %s", role))
	}

	// Generate email verification token
	verificationToken, err := s.generateSecureToken()
	if err != nil {
		return nil, errors.NewInternalError("Failed to generate verification token").WithInternal(err)
	}

	// Create user
	user := models.User{
		Email:                 req.Email,
		Username:              req.Username,
		Password:              req.Password, // Will be hashed in BeforeCreate hook
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Phone:                 req.Phone,
		CompanyID:             req.CompanyID,
		Role:                  role,
		Status:                "active",
		IsActive:              true,
		IsVerified:            false,
		EmailVerificationToken: verificationToken,
		Language:              "id",
		Timezone:              "Asia/Jakarta",
	}

	// Save user to database
	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.NewInternalError("Failed to create user").WithInternal(err)
	}

	// TODO: Send verification email
	// s.sendVerificationEmail(user.Email, verificationToken)

	return s.userToResponse(&user), nil
}

// Login authenticates a user and returns JWT tokens
func (s *Service) Login(req LoginRequest) (*UserResponse, *TokenResponse, error) {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ? AND is_active = true", req.Email).First(&user).Error; err != nil {
		// Don't reveal whether email exists or not
		return nil, nil, errors.NewUnauthorizedError("Invalid email or password")
	}

	// Check if account is locked
	if user.IsAccountLocked() {
		return nil, nil, errors.NewForbiddenError("Account is locked due to too many failed login attempts")
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		// Increment failed attempts
		user.IncrementFailedAttempts()
		s.db.Save(&user)
		return nil, nil, errors.NewUnauthorizedError("Invalid email or password")
	}

	// Reset failed attempts on successful login
	user.ResetFailedAttempts()
	user.UpdateLastLogin()
	s.db.Save(&user)

	// Generate JWT tokens
	tokenResponse, err := s.generateTokens(&user)
	if err != nil {
		return nil, nil, errors.NewInternalError("Failed to generate tokens").WithInternal(err)
	}

	// Create session
	if err := s.createSession(&user, tokenResponse.AccessToken, tokenResponse.RefreshToken); err != nil {
		return nil, nil, errors.NewInternalError("Failed to create session").WithInternal(err)
	}

	return s.userToResponse(&user), tokenResponse, nil
}

// RefreshToken generates new access token using refresh token
func (s *Service) RefreshToken(refreshToken string) (*TokenResponse, error) {
	// Find session by refresh token
	var session models.Session
	if err := s.db.Where("refresh_token = ? AND is_active = true AND expires_at > ?", refreshToken, time.Now()).First(&session).Error; err != nil {
		return nil, errors.NewUnauthorizedError("Invalid or expired refresh token")
	}

	// Get user
	var user models.User
	if err := s.db.Where("id = ?", session.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("User")
		}
		return nil, errors.NewInternalError("Failed to fetch user").WithInternal(err)
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, errors.NewForbiddenError("User account is inactive")
	}

	// Generate new tokens
	tokenResponse, err := s.generateTokens(&user)
	if err != nil {
		return nil, errors.NewInternalError("Failed to generate tokens").WithInternal(err)
	}

	// Update session with new tokens
	session.Token = tokenResponse.AccessToken
	session.RefreshToken = tokenResponse.RefreshToken
	session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days
	s.db.Save(&session)

	return tokenResponse, nil
}

// Logout invalidates user session
func (s *Service) Logout(accessToken string) error {
	// Find and deactivate session
	if err := s.db.Model(&models.Session{}).Where("token = ?", accessToken).Update("is_active", false).Error; err != nil {
		return errors.NewInternalError("Failed to logout").WithInternal(err)
	}
	return nil
}

// ValidateToken validates JWT token and returns claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, errors.NewUnauthorizedError("Invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.NewUnauthorizedError("Invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.NewUnauthorizedError("Token expired")
	}

	// Verify user still exists and is active
	var user models.User
	if err := s.db.Where("id = ? AND is_active = true", claims.UserID).First(&user).Error; err != nil {
		return nil, errors.NewUnauthorizedError("User not found or inactive")
	}

	return claims, nil
}

// GetProfile returns user profile information
func (s *Service) GetProfile(userID string) (*UserResponse, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("User")
		}
		return nil, errors.NewInternalError("Failed to fetch user").WithInternal(err)
	}

	return s.userToResponse(&user), nil
}

// UpdateProfile updates user profile information
func (s *Service) UpdateProfile(userID string, updates map[string]interface{}) (*UserResponse, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("User")
		}
		return nil, errors.NewInternalError("Failed to fetch user").WithInternal(err)
	}

	// Update allowed fields
	allowedFields := []string{"first_name", "last_name", "phone", "language", "timezone", "preferences"}
	for _, field := range allowedFields {
		if value, exists := updates[field]; exists {
			s.db.Model(&user).Update(field, value)
		}
	}

	// Reload user
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.NewInternalError("Failed to reload user").WithInternal(err)
	}

	return s.userToResponse(&user), nil
}

// ChangePassword changes user password
func (s *Service) ChangePassword(userID, currentPassword, newPassword string) error {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("User")
		}
		return errors.NewInternalError("Failed to fetch user").WithInternal(err)
	}

	// Verify current password
	if !user.CheckPassword(currentPassword) {
		return errors.NewUnauthorizedError("Current password is incorrect")
	}

	// Validate new password strength
	if err := s.validatePasswordStrength(newPassword); err != nil {
		return err
	}

	// Update password
	user.Password = newPassword // Will be hashed in BeforeUpdate hook
	user.PasswordChangedAt = time.Now()
	
	if err := s.db.Save(&user).Error; err != nil {
		return errors.NewInternalError("Failed to update password").WithInternal(err)
	}

	// Invalidate all sessions for security
	if err := s.db.Model(&models.Session{}).Where("user_id = ?", userID).Update("is_active", false).Error; err != nil {
		return errors.NewInternalError("Failed to invalidate sessions").WithInternal(err)
	}

	return nil
}

// ForgotPassword initiates password reset process
func (s *Service) ForgotPassword(email string) error {
	var user models.User
	if err := s.db.Where("email = ? AND is_active = true", email).First(&user).Error; err != nil {
		// Don't reveal if email exists or not (security best practice)
		return nil
	}

	// Generate reset token
	resetToken, err := s.generateSecureToken()
	if err != nil {
		return errors.NewInternalError("Failed to generate reset token").WithInternal(err)
	}

	// Create password reset token record
	resetTokenRecord := models.PasswordResetToken{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiry
	}

	if err := s.db.Create(&resetTokenRecord).Error; err != nil {
		return errors.NewInternalError("Failed to create reset token").WithInternal(err)
	}

	// TODO: Send password reset email
	// s.sendPasswordResetEmail(user.Email, resetToken)

	return nil
}

// ResetPassword resets user password using reset token
func (s *Service) ResetPassword(token, newPassword string) error {
	// Find valid reset token
	var resetToken models.PasswordResetToken
	if err := s.db.Where("token = ? AND expires_at > ? AND used_at IS NULL", token, time.Now()).First(&resetToken).Error; err != nil {
		return errors.NewUnauthorizedError("Invalid or expired reset token")
	}

	// Get user
	var user models.User
	if err := s.db.Where("id = ?", resetToken.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewNotFoundError("User")
		}
		return errors.NewInternalError("Failed to fetch user").WithInternal(err)
	}

	// Validate new password strength
	if err := s.validatePasswordStrength(newPassword); err != nil {
		return err
	}

	// Update password
	user.Password = newPassword // Will be hashed in BeforeUpdate hook
	user.PasswordChangedAt = time.Now()
	
	if err := s.db.Save(&user).Error; err != nil {
		return errors.NewInternalError("Failed to update password").WithInternal(err)
	}

	// Mark reset token as used
	now := time.Now()
	resetToken.UsedAt = &now
	s.db.Save(&resetToken)

	// Invalidate all sessions for security
	if err := s.db.Model(&models.Session{}).Where("user_id = ?", user.ID).Update("is_active", false).Error; err != nil {
		return errors.NewInternalError("Failed to invalidate sessions").WithInternal(err)
	}

	return nil
}

// generateTokens creates JWT access and refresh tokens
func (s *Service) generateTokens(user *models.User) (*TokenResponse, error) {
	// Access token (15 minutes)
	accessClaims := &Claims{
		UserID:    user.ID,
		CompanyID: user.CompanyID,
		Role:      user.Role,
		Username:  user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh token (7 days)
	refreshToken, err := s.generateSecureToken()
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
		TokenType:    "Bearer",
	}, nil
}

// createSession creates a new user session
func (s *Service) createSession(user *models.User, accessToken, refreshToken string) error {
	session := models.Session{
		UserID:       user.ID,
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
		IsActive:     true,
	}

	return s.db.Create(&session).Error
}

// generateSecureToken generates a cryptographically secure random token
func (s *Service) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// validatePasswordStrength validates password strength
func (s *Service) validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.NewValidationError("Password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.NewValidationError("Password must be less than 128 characters")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.NewValidationError("Password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.NewValidationError("Password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.NewValidationError("Password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.NewValidationError("Password must contain at least one special character")
	}

	return nil
}

// userToResponse converts User model to UserResponse
func (s *Service) userToResponse(user *models.User) *UserResponse {
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

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
