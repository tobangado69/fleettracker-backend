package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Registration failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	user, tokens, err := h.service.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Login failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
		"tokens":  tokens,
	})
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	tokens, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token refresh failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"tokens":  tokens,
	})
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	// Get access token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Authorization header required",
			"message": "Access token not provided",
		})
		return
	}

	// Extract token from "Bearer <token>"
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	err := h.service.Logout(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Logout failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// GetProfile handles getting user profile
func (h *Handler) GetProfile(c *gin.Context) {
	// Get user ID from JWT claims (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	user, err := h.service.GetProfile(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Profile not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UpdateProfile handles updating user profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	// Get user ID from JWT claims (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	user, err := h.service.UpdateProfile(userID.(string), updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Profile update failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// ChangePassword handles changing user password
func (h *Handler) ChangePassword(c *gin.Context) {
	// Get user ID from JWT claims (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User ID not found in context",
		})
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	err := h.service.ChangePassword(userID.(string), req.CurrentPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Password change failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// ForgotPassword handles forgot password request
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	err := h.service.ForgotPassword(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Password reset failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset with token
func (h *Handler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	err := h.service.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Password reset failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// GetActiveSessions handles getting user's active sessions
func (h *Handler) GetActiveSessions(c *gin.Context) {
	// TODO: Implement get active sessions
	c.JSON(http.StatusOK, gin.H{
		"message": "Get active sessions not implemented yet",
	})
}

// RevokeSession handles revoking a specific session
func (h *Handler) RevokeSession(c *gin.Context) {
	// TODO: Implement revoke session
	c.JSON(http.StatusOK, gin.H{
		"message": "Revoke session not implemented yet",
	})
}
