package exceptions

import (
	"fsm-backend/pkg/response"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type AppError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *AppError) Error() string {
	return e.Message
}

func FiberErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "An unexpected server error occurred."
	errCode := "INTERNAL_SERVER_ERROR"

	if appErr, ok := err.(*AppError); ok {
		code = appErr.Status
		message = appErr.Message
		errCode = appErr.Code
	} else if fibErr, ok := err.(*fiber.Error); ok {
		code = fibErr.Code
		message = fibErr.Message
		if code == fiber.StatusNotFound {
			errCode = "ROUTE_NOT_FOUND"
		} else {
			errCode = "HTTP_ERROR"
		}
	}

	return c.Status(code).JSON(response.APIResponse{
		Success: false,
		Message: message,
		Code:    errCode,
	})
}

func NewBadRequestError(message string, code string) *AppError {
	return &AppError{
		Status:  http.StatusBadRequest,
		Message: message,
		Code:    code,
	}
}

func NewNotFoundError(message string, code string) *AppError {
	return &AppError{
		Status:  http.StatusNotFound,
		Message: message,
		Code:    code,
	}
}

func NewConflictError(message string, code string) *AppError {
	return &AppError{
		Status:  http.StatusConflict,
		Message: message,
		Code:    code,
	}
}

func NewInternalServerError(message string) *AppError {
	return &AppError{
		Status:  http.StatusInternalServerError,
		Message: message,
		Code:    "INTERNAL_SERVER_ERROR",
	}
}
