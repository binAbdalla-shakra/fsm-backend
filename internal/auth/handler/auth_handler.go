package handler

import (
	"fsm-backend/exceptions"
	"fsm-backend/internal/auth/dto"
	"fsm-backend/internal/auth/service"
	"fsm-backend/messages"
	"fsm-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) SignUp(c *fiber.Ctx) error {
	var payload dto.SignUpPayload
	if err := c.BodyParser(&payload); err != nil {
		return exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	// Capture device info headers if not set in body
	if payload.DeviceID == "" {
		payload.DeviceID = c.Get("X-Device-ID", "default_device")
	}
	if payload.DeviceName == "" {
		payload.DeviceName = c.Get("X-Device-Name", "Unknown Device")
	}

	err := h.service.SignUpCustomer(c.UserContext(), &payload)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusCreated, "Verification OTP code sent to your phone number.", nil)
}

func (h *AuthHandler) RequestOTP(c *fiber.Ctx) error {
	var payload dto.SendOTPRequest
	if err := c.BodyParser(&payload); err != nil {
		return exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	err := h.service.SendOTP(c.UserContext(), &payload)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "OTP sent successfully.", nil)
}

func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var payload dto.VerifyOTPPayload
	if err := c.BodyParser(&payload); err != nil {
		return exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	if payload.DeviceID == "" {
		payload.DeviceID = c.Get("X-Device-ID", "default_device")
	}
	if payload.DeviceName == "" {
		payload.DeviceName = c.Get("X-Device-Name", "Unknown Device")
	}
	ipAddress := c.IP()

	tokens, err := h.service.VerifyOTPAndLogin(c.UserContext(), &payload, ipAddress)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Login successful.", tokens)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var payload dto.RefreshTokenPayload
	if err := c.BodyParser(&payload); err != nil {
		return exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	if payload.DeviceID == "" {
		payload.DeviceID = c.Get("X-Device-ID", "default_device")
	}
	ipAddress := c.IP()

	tokens, err := h.service.RefreshSession(c.UserContext(), &payload, ipAddress)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Session tokens refreshed.", tokens)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return exceptions.NewBadRequestError("Unauthorized session", "UNAUTHORIZED")
	}

	userID := userIDVal.(string)
	deviceID := c.Get("X-Device-ID", "default_device")

	err := h.service.Logout(c.UserContext(), userID, deviceID)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	return response.SendSuccess(c, fiber.StatusOK, "Logout successful. Active session invalidated.", nil)
}
