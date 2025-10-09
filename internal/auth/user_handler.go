package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tobangado69/fleettracker-pro/backend/internal/common/middleware"
)

// CreateUser handles user creation (admin-only)
// @Summary Create new user
// @Description Create a new user within the company (admin-only, role hierarchy enforced)
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "Create user request"
// @Success 201 {object} SuccessResponse{data=UserResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/users [post]
// @Security BearerAuth
func (h *Handler) CreateUser(c *gin.Context) {
	// Get creator info from context
	creatorUserID, _ := c.Get("user_id")
	creatorRole, _ := c.Get("role")
	creatorCompanyID, _ := c.Get("company_id")

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	user, appErr := h.service.CreateUser(
		c.Request.Context(),
		creatorUserID.(string),
		creatorRole.(string),
		creatorCompanyID.(string),
		&req,
	)
	if appErr != nil {
		middleware.AbortWithError(c, appErr)
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    toUserResponse(user),
		Message: "User created successfully",
	})
}

// ListUsers handles listing users (admin-only)
// @Summary List users
// @Description List all users in the company (admin-only)
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Results per page" default(10)
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/users [get]
// @Security BearerAuth
func (h *Handler) ListUsers(c *gin.Context) {
	// Get user info from context
	userRole, _ := c.Get("role")
	companyID, _ := c.Get("company_id")

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, appErr := h.service.GetUsers(
		c.Request.Context(),
		userRole.(string),
		companyID.(string),
		page,
		limit,
	)
	if appErr != nil {
		middleware.AbortWithError(c, appErr)
		return
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *toUserResponse(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    userResponses,
		"meta": gin.H{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetUserByID handles getting a single user (admin-only)
// @Summary Get user by ID
// @Description Get user details by ID (admin-only)
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} SuccessResponse{data=UserResponse}
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/{id} [get]
// @Security BearerAuth
func (h *Handler) GetUserByID(c *gin.Context) {
	userRole, _ := c.Get("role")
	companyID, _ := c.Get("company_id")
	targetUserID := c.Param("id")

	user, appErr := h.service.GetUser(
		c.Request.Context(),
		userRole.(string),
		companyID.(string),
		targetUserID,
	)
	if appErr != nil {
		middleware.AbortWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    toUserResponse(user),
	})
}

// UpdateUser handles updating a user (admin-only)
// @Summary Update user
// @Description Update user details (admin-only)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateUserRequest true "Update user request"
// @Success 200 {object} SuccessResponse{data=UserResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/{id} [put]
// @Security BearerAuth
func (h *Handler) UpdateUser(c *gin.Context) {
	userRole, _ := c.Get("role")
	companyID, _ := c.Get("company_id")
	targetUserID := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	user, appErr := h.service.UpdateUser(
		c.Request.Context(),
		userRole.(string),
		companyID.(string),
		targetUserID,
		&req,
	)
	if appErr != nil {
		middleware.AbortWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    toUserResponse(user),
		Message: "User updated successfully",
	})
}

// DeactivateUser handles user deactivation (owner/super-admin only)
// @Summary Deactivate user
// @Description Deactivate a user account (owner/super-admin only)
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/{id} [delete]
// @Security BearerAuth
func (h *Handler) DeactivateUser(c *gin.Context) {
	userRole, _ := c.Get("role")
	companyID, _ := c.Get("company_id")
	targetUserID := c.Param("id")

	appErr := h.service.DeactivateUser(
		c.Request.Context(),
		userRole.(string),
		companyID.(string),
		targetUserID,
	)
	if appErr != nil {
		middleware.AbortWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User deactivated successfully",
	})
}

// ChangeUserRole handles changing a user's role (admin-only)
// @Summary Change user role
// @Description Change a user's role (admin-only, role hierarchy enforced)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body ChangeRoleRequest true "Change role request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/{id}/role [put]
// @Security BearerAuth
func (h *Handler) ChangeUserRole(c *gin.Context) {
	changerRole, _ := c.Get("role")
	changerCompanyID, _ := c.Get("company_id")
	targetUserID := c.Param("id")

	var req ChangeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithValidation(c, err.Error())
		return
	}

	appErr := h.service.ChangeUserRole(
		c.Request.Context(),
		changerRole.(string),
		changerCompanyID.(string),
		targetUserID,
		req.NewRole,
	)
	if appErr != nil {
		middleware.AbortWithError(c, appErr)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "User role changed successfully",
	})
}

// GetAllowedRoles returns roles that the current user can assign
// @Summary Get allowed roles
// @Description Get list of roles that current user can assign
// @Tags users
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/allowed-roles [get]
// @Security BearerAuth
func (h *Handler) GetAllowedRoles(c *gin.Context) {
	userRole, _ := c.Get("role")

	allowedRoles := GetAllowedRoles(userRole.(string))

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data: gin.H{
			"allowed_roles": allowedRoles,
			"descriptions": getRoleDescriptions(allowedRoles),
		},
	})
}

// Helper functions

func getRoleDescriptions(roles []string) map[string]string {
	descriptions := make(map[string]string)
	for _, role := range roles {
		descriptions[role] = RoleDescription(role)
	}
	return descriptions
}

