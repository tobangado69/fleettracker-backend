package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/validators"
)

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty" example:"Operation successful"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Error   string      `json:"error" example:"Bad request"`
	Message string      `json:"message,omitempty" example:"Invalid input"`
	Data    interface{} `json:"data,omitempty"` // Additional error context
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Success bool                   `json:"success" example:"false"`
	Error   string                 `json:"error" example:"Validation failed"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldPassword123"`
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"newPassword123"`
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// ResetPasswordRequest represents password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required" example:"reset_token_123"`
	NewPassword string `json:"new_password" binding:"required,min=8" example:"newPassword123"`
}

// Handler handles authentication HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new authentication handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Register handles user registration (RESTRICTED: First user/company owner only)
// @Summary Register new user
// @Description Register a new company owner account (restricted to first user only). For additional users, contact your company administrator.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data"
// @Success 201 {object} SuccessResponse{data=UserResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse "Registration closed - contact your admin"
// @Failure 422 {object} ValidationErrorResponse
// @Router /api/v1/auth/register [post]
// Register endpoint is DEPRECATED and removed for security
// @Summary [DEPRECATED] User registration
// @Description This endpoint is deprecated. FleetTracker Pro uses an invite-only system. Contact your administrator.
// @Tags auth
// @Produce json
// @Success 410 {object} ErrorResponse "Endpoint deprecated - use invite-only system"
// @Router /api/v1/auth/register [post]
// @Deprecated
func (h *Handler) Register(c *gin.Context) {
	c.JSON(http.StatusGone, ErrorResponse{
		Success: false,
		Error:   "endpoint_deprecated",
		Message: "Public registration is no longer supported. FleetTracker Pro uses an invite-only system. Please contact your company administrator to create an account, or contact support@fleettracker.id for help.",
		Data: gin.H{
			"reason":            "invite_only_system",
			"how_to_get_access": "Contact your company administrator or support@fleettracker.id",
			"documentation":     "See POST /api/v1/users endpoint for user creation by administrators",
		},
	})
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with email and password, returns JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} SuccessResponse{data=UserResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate email format
	if err := validators.ValidateEmail(req.Email); err != nil {
		middleware.AbortWithBadRequest(c, "Invalid email: "+err.Error())
		return
	}

	// Sanitize email (trim, lowercase)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, tokens, err := h.service.Login(req)
	if err != nil {
		middleware.AbortWithUnauthorized(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
		"tokens":  tokens,
	})
}

// RefreshToken handles token refresh
// @Summary Refresh JWT token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token data"
// @Success 200 {object} SuccessResponse{data=TokenResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	tokens, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		middleware.AbortWithUnauthorized(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"tokens":  tokens,
	})
}

// Logout handles user logout
// @Summary User logout
// @Description Logout user and invalidate JWT token
// @Tags auth
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/logout [post]
// @Security BearerAuth
func (h *Handler) Logout(c *gin.Context) {
	// Get access token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		middleware.AbortWithBadRequest(c, "Access token not provided")
		return
	}

	// Extract token from "Bearer <token>"
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	err := h.service.Logout(tokenString)
	if err != nil {
		middleware.AbortWithInternal(c, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// GetProfile handles getting user profile
// @Summary Get user profile
// @Description Get current user profile information
// @Tags auth
// @Produce json
// @Success 200 {object} SuccessResponse{data=UserResponse}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/auth/profile [get]
// @Security BearerAuth
func (h *Handler) GetProfile(c *gin.Context) {
	// Get user ID from JWT claims (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "User ID not found in context")
		return
	}

	user, err := h.service.GetProfile(userID.(string))
	if err != nil {
		middleware.AbortWithNotFound(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateProfile handles updating user profile
// @Summary Update user profile
// @Description Update current user profile information
// @Tags auth
// @Accept json
// @Produce json
// @Param updates body map[string]interface{} true "Profile updates"
// @Success 200 {object} SuccessResponse{data=UserResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/profile [put]
// @Security BearerAuth
func (h *Handler) UpdateProfile(c *gin.Context) {
	// Get user ID from JWT claims (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "User ID not found in context")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	user, err := h.service.UpdateProfile(userID.(string), updates)
	if err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// ChangePassword handles changing user password
// @Summary Change user password
// @Description Change current user password with validation
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Password change data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/change-password [put]
// @Security BearerAuth
func (h *Handler) ChangePassword(c *gin.Context) {
	// Get user ID from JWT claims (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "User ID not found in context")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate new password
	if err := validators.ValidatePassword(req.NewPassword); err != nil {
		middleware.AbortWithBadRequest(c, "Invalid new password: "+err.Error())
		return
	}

	err := h.service.ChangePassword(userID.(string), req.CurrentPassword, req.NewPassword)
	if err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// ForgotPassword handles forgot password request
// @Summary Request password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Email address"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate and sanitize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := validators.ValidateEmail(req.Email); err != nil {
		middleware.AbortWithBadRequest(c, "Invalid email: "+err.Error())
		return
	}

	err := h.service.ForgotPassword(req.Email)
	if err != nil {
		middleware.AbortWithInternal(c, err.Error(), err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset with token
// @Summary Reset password with token
// @Description Reset user password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Password reset data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	// Validate new password
	if err := validators.ValidatePassword(req.NewPassword); err != nil {
		middleware.AbortWithBadRequest(c, "Invalid password: "+err.Error())
		return
	}

	err := h.service.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		middleware.AbortWithBadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// GetActiveSessions handles getting user's active sessions
// @Summary Get active sessions
// @Description Get all active sessions for the current user
// @Tags auth
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/sessions [get]
// @Security BearerAuth
func (h *Handler) GetActiveSessions(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "User ID not found")
		return
	}

	// Get current token from header
	token := c.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	sessions, appErr := h.service.GetActiveSessions(c.Request.Context(), userID.(string), token)
	if appErr != nil {
		middleware.AbortWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    sessions,
	})
}

// RevokeSession handles revoking a specific session
// @Summary Revoke session
// @Description Revoke a specific session by ID
// @Tags auth
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/sessions/{id} [delete]
// @Security BearerAuth
func (h *Handler) RevokeSession(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.AbortWithUnauthorized(c, "User ID not found")
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		middleware.AbortWithBadRequest(c, "Session ID is required")
		return
	}

	appErr := h.service.RevokeSession(c.Request.Context(), userID.(string), sessionID)
	if appErr != nil {
		middleware.AbortWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Session revoked successfully",
	})
}
