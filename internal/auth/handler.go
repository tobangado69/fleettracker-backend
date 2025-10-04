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
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Name     string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	user, err := h.service.Register(req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
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
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	user, token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Login failed",
			"message": "Invalid credentials",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	// TODO: Implement token refresh logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Token refresh not implemented yet",
	})
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	// TODO: Implement logout logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// GetProfile handles getting user profile
func (h *Handler) GetProfile(c *gin.Context) {
	// TODO: Implement get profile logic
	c.JSON(http.StatusOK, gin.H{
		"user": map[string]interface{}{
			"id":    "temp-user-id",
			"email": "demo@fleettracker.id",
			"name":  "Demo User",
			"role":  "fleet_manager",
		},
	})
}

// UpdateProfile handles updating user profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	// TODO: Implement update profile logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile update not implemented yet",
	})
}
