package auth

import (
	"fmt"
)

// Role constants
const (
	RoleSuperAdmin = "super-admin"
	RoleOwner      = "owner"
	RoleAdmin      = "admin"
	RoleOperator   = "operator"
	RoleDriver     = "driver"
)

// AllRoles returns all available roles
func AllRoles() []string {
	return []string{
		RoleSuperAdmin,
		RoleOwner,
		RoleAdmin,
		RoleOperator,
		RoleDriver,
	}
}

// RoleHierarchy defines who can create which roles
var RoleHierarchy = map[string][]string{
	RoleSuperAdmin: {RoleSuperAdmin, RoleOwner, RoleAdmin, RoleOperator, RoleDriver},
	RoleOwner:      {RoleAdmin, RoleOperator, RoleDriver},
	RoleAdmin:      {RoleOperator, RoleDriver},
	RoleOperator:   {},
	RoleDriver:     {},
}

// CanCreateRole checks if a role can create another role
func CanCreateRole(creatorRole, targetRole string) bool {
	allowedRoles, exists := RoleHierarchy[creatorRole]
	if !exists {
		return false
	}

	for _, allowed := range allowedRoles {
		if allowed == targetRole {
			return true
		}
	}

	return false
}

// CanManageUsers checks if a role can manage users
func CanManageUsers(role string) bool {
	return role == RoleSuperAdmin || role == RoleOwner || role == RoleAdmin
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	for _, validRole := range AllRoles() {
		if role == validRole {
			return true
		}
	}
	return false
}

// GetRolePriority returns the priority level of a role (higher = more privileged)
func GetRolePriority(role string) int {
	priorities := map[string]int{
		RoleSuperAdmin: 5,
		RoleOwner:      4,
		RoleAdmin:      3,
		RoleOperator:   2,
		RoleDriver:     1,
	}

	if priority, exists := priorities[role]; exists {
		return priority
	}
	return 0
}

// CanAssignRole checks if a user can assign a specific role
// Prevents privilege escalation - users can only assign roles equal to or lower than their own
func CanAssignRole(assignerRole, targetRole string) bool {
	assignerPriority := GetRolePriority(assignerRole)
	targetPriority := GetRolePriority(targetRole)

	// Can only assign roles with equal or lower priority
	return assignerPriority >= targetPriority
}

// ValidateRoleCreation validates if role creation is allowed
func ValidateRoleCreation(creatorRole, targetRole string) error {
	// Check if creator role is valid
	if !IsValidRole(creatorRole) {
		return fmt.Errorf("invalid creator role: %s", creatorRole)
	}

	// Check if target role is valid
	if !IsValidRole(targetRole) {
		return fmt.Errorf("invalid target role: %s", targetRole)
	}

	// Check if creator can manage users
	if !CanManageUsers(creatorRole) {
		return fmt.Errorf("role %s cannot create users", creatorRole)
	}

	// Check if creator can create target role
	if !CanCreateRole(creatorRole, targetRole) {
		return fmt.Errorf("role %s cannot create users with role %s", creatorRole, targetRole)
	}

	return nil
}

// ValidateRoleAssignment validates if role assignment is allowed
func ValidateRoleAssignment(assignerRole, currentRole, newRole string) error {
	// Check if assigner can assign the new role
	if !CanAssignRole(assignerRole, newRole) {
		return fmt.Errorf("role %s cannot assign role %s (privilege escalation prevented)", assignerRole, newRole)
	}

	// Super-admin can change any role
	if assignerRole == RoleSuperAdmin {
		return nil
	}

	// Owner can change roles within their hierarchy
	if assignerRole == RoleOwner {
		if newRole == RoleSuperAdmin || newRole == RoleOwner {
			return fmt.Errorf("owner cannot assign super-admin or owner roles")
		}
		return nil
	}

	// Admin can only create operator/driver
	if assignerRole == RoleAdmin {
		if newRole != RoleOperator && newRole != RoleDriver {
			return fmt.Errorf("admin can only assign operator or driver roles")
		}
		return nil
	}

	return fmt.Errorf("insufficient permissions to assign roles")
}

// GetAllowedRoles returns roles that a user can create
func GetAllowedRoles(userRole string) []string {
	if allowed, exists := RoleHierarchy[userRole]; exists {
		return allowed
	}
	return []string{}
}

// RoleDescription returns a description of the role
func RoleDescription(role string) string {
	descriptions := map[string]string{
		RoleSuperAdmin: "Full system access, can manage all users and companies",
		RoleOwner:      "Company owner, can manage company users and all resources",
		RoleAdmin:      "Client administrator, can manage operators and drivers",
		RoleOperator:   "Regular user, can access company resources",
		RoleDriver:     "Mobile app user, can track trips and update location",
	}

	if desc, exists := descriptions[role]; exists {
		return desc
	}
	return "Unknown role"
}

