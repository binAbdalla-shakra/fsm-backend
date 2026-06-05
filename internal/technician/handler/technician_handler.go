package handler

import (
	"fsm-backend/exceptions"
	"fsm-backend/internal/domain"
	"fsm-backend/internal/technician/service"
	"fsm-backend/messages"
	"fsm-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type TechnicianHandler struct {
	service service.TechnicianService
}

func NewTechnicianHandler(service service.TechnicianService) *TechnicianHandler {
	return &TechnicianHandler{service: service}
}

type registerTechRequest struct {
	Phone          string   `json:"phone"`
	Email          *string  `json:"email,omitempty"`
	Name           string   `json:"name"`
	Skills         []string `json:"skills"`
	ZoneAssignment *string  `json:"zone_assignment,omitempty"`
	Latitude       float64  `json:"latitude"`
	Longitude      float64  `json:"longitude"`
}

func (h *TechnicianHandler) Register(c *fiber.Ctx) error {
	var req registerTechRequest
	if err := c.BodyParser(&req); err != nil {
		return exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	user := domain.User{
		Phone: req.Phone,
		Email: req.Email,
		Name:  req.Name,
	}

	res, err := h.service.RegisterTechnician(c.UserContext(), &user, req.Skills, req.ZoneAssignment, req.Latitude, req.Longitude)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusCreated, messages.SuccessTechCreated, res)
}

func (h *TechnicianHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Technician profile retrieved.", res)
}

func (h *TechnicianHandler) GetMe(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return response.SendError(c, fiber.StatusUnauthorized, "Unauthorized session context", "UNAUTHORIZED")
	}

	userID := userIDVal.(string)
	res, err := h.service.GetByUserID(c.UserContext(), userID)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Technician profile retrieved.", res)
}

type updateStatusRequest struct {
	Status    string  `json:"status"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (h *TechnicianHandler) UpdateStatus(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return response.SendError(c, fiber.StatusUnauthorized, "Unauthorized session context", "UNAUTHORIZED")
	}
	userID := userIDVal.(string)

	var req updateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	// Resolve tech ID from user ID
	tech, err := h.service.GetByUserID(c.UserContext(), userID)
	if err != nil {
		return err
	}

	err = h.service.UpdateStatus(c.UserContext(), tech.ID, req.Status, req.Latitude, req.Longitude)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, messages.SuccessTechStatusUpdated, nil)
}
