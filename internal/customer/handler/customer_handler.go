package handler

import (
	"fsm-backend/internal/customer/service"
	"fsm-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type CustomerHandler struct {
	service service.CustomerService
}

func NewCustomerHandler(service service.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) GetMe(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return response.SendError(c, fiber.StatusUnauthorized, "Unauthorized session context", "UNAUTHORIZED")
	}

	userID := userIDVal.(string)
	res, err := h.service.GetCustomerByUserID(c.UserContext(), userID)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Customer profile retrieved.", res)
}

func (h *CustomerHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.service.GetCustomerByID(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Customer profile retrieved.", res)
}
