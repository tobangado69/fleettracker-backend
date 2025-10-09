package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireAuth middleware ensures user is authenticated
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user_id exists in context (set by JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists || userID == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Success: false,
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware ensures user has one of the specified roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Success: false,
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		role := userRole.(string)

		// Check if user has one of the allowed roles
		hasRole := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Success: false,
				Error:   "Forbidden",
				Message: "Insufficient permissions. Required roles: " + strings.Join(allowedRoles, ", "),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdminRole middleware ensures user can manage other users
func RequireAdminRole() gin.HandlerFunc {
	return RequireRole(RoleSuperAdmin, RoleOwner, RoleAdmin)
}

// RequireOwnerRole middleware ensures user is super-admin or owner
func RequireOwnerRole() gin.HandlerFunc {
	return RequireRole(RoleSuperAdmin, RoleOwner)
}

// RequireSuperAdmin middleware ensures user is super-admin
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(RoleSuperAdmin)
}

// ValidateCompanyAccess middleware ensures user can only access their own company data
func ValidateCompanyAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user's company from context
		userCompanyID, exists := c.Get("company_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Success: false,
				Error:   "Unauthorized",
				Message: "Company information not found",
			})
			c.Abort()
			return
		}

		// Get requested company ID from URL or query params
		requestedCompanyID := c.Param("company_id")
		if requestedCompanyID == "" {
			requestedCompanyID = c.Query("company_id")
		}

		// If no company specified in request, allow (will use user's company)
		if requestedCompanyID == "" {
			c.Next()
			return
		}

		// Super-admin can access any company
		userRole, _ := c.Get("role")
		if userRole.(string) == RoleSuperAdmin {
			c.Next()
			return
		}

		// Check if user is accessing their own company
		if requestedCompanyID != userCompanyID.(string) {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Success: false,
				Error:   "Forbidden",
				Message: "You can only access data from your own company",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

