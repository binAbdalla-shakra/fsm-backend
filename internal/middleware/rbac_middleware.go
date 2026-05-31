package middleware

import (
	"fsm-backend/exceptions"

	"github.com/gofiber/fiber/v2"
)

// HasPermission checks if the authenticated user has the required permission code in their JWT claims.
func HasPermission(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permissionsVal := c.Locals("permissions")
		if permissionsVal == nil {
			return exceptions.NewBadRequestError("Forbidden: Insufficient permissions for this action", "FORBIDDEN")
		}

		permissions, ok := permissionsVal.([]string)
		if !ok {
			return exceptions.NewBadRequestError("Forbidden: Insufficient permissions for this action", "FORBIDDEN")
		}

		hasPerm := false
		for _, perm := range permissions {
			if perm == requiredPermission {
				hasPerm = true
				break
			}
		}

		if !hasPerm {
			return exceptions.NewBadRequestError("Forbidden: Access denied for requested action", "ACCESS_DENIED")
		}

		return c.Next()
	}
}

// RequireRole checks if the user's role equals the required string.
func RequireRole(roleName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRoleVal := c.Locals("role")
		if userRoleVal == nil {
			return exceptions.NewBadRequestError("Forbidden: Role requirements not met", "ROLE_REQUIRED")
		}

		userRole, ok := userRoleVal.(string)
		if !ok || userRole != roleName {
			return exceptions.NewBadRequestError("Forbidden: Restricted resource action", "ROLE_ACCESS_DENIED")
		}

		return c.Next()
	}
}
