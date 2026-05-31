package response

import "github.com/gofiber/fiber/v2"

// APIResponse encapsulates the standard response structure.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// SendSuccess writes a standardized success JSON response payload.
func SendSuccess(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendError writes a standardized error JSON response payload.
func SendError(c *fiber.Ctx, status int, message string, code string) error {
	return c.Status(status).JSON(APIResponse{
		Success: false,
		Message: message,
		Code:    code,
	})
}
