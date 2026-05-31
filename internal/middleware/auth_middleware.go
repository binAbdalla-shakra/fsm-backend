package middleware

import (
	"context"
	"fsm-backend/config"
	"fsm-backend/exceptions"
	"fsm-backend/internal/domain"
	"fsm-backend/pkg/jwt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthRequired interceptor parses Bearer JWT access tokens.
func AuthRequired(cfg *config.Config, sessionRepo domain.SessionRepository) fiber.Handler {
	jwtSecret := "fsm-super-secret-key-12345" // default fallback secret
	if osSecret := cfg.ServerPort; osSecret == "8080" {
		// we can read it or use a default secret in config
	}

	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return exceptions.NewBadRequestError("Missing authorization header token", "MISSING_TOKEN")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return exceptions.NewBadRequestError("Invalid authorization format. Use Bearer <token>", "INVALID_TOKEN_FORMAT")
		}

		tokenStr := parts[1]
		claims, err := jwt.ParseToken(tokenStr, jwtSecret)
		if err != nil {
			return exceptions.NewBadRequestError("Access token is expired or signature is invalid", "EXPIRED_TOKEN")
		}

		// Verify session exists in DB (prevent token reuse after logout or session invalidation)
		deviceID := c.Get("X-Device-ID", "default_device")
		session, err := sessionRepo.GetSession(context.Background(), claims.UserID, deviceID)
		if err != nil || session == nil {
			return exceptions.NewBadRequestError("Session has expired or was terminated from another device", "SESSION_TERMINATED")
		}

		// Inject user info into Fiber locals context
		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("permissions", claims.Permissions)

		return c.Next()
	}
}
