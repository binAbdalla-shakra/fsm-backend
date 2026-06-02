package handler

import (
	"fsm-backend/exceptions"
	"fsm-backend/internal/ticket/dto"
	"fsm-backend/internal/ticket/service"
	"fsm-backend/messages"
	"fsm-backend/pkg/response"
	"io"

	"github.com/gofiber/fiber/v2"
)

type TicketHandler struct {
	service service.TicketService
}

func NewTicketHandler(service service.TicketService) *TicketHandler {
	return &TicketHandler{service: service}
}

func (h *TicketHandler) Report(c *fiber.Ctx) error {
	var req dto.CreateTicketRequest
	if err := c.BodyParser(&req); err != nil {
		return exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	customerUserID, ok := c.Locals("userID").(string)
	if !ok {
		return exceptions.NewBadRequestError("Unauthorized", "UNAUTHORIZED")
	}

	res, err := h.service.ReportTicket(
		c.UserContext(),
		req.Title,
		req.Description,
		req.SkillRequired,
		"MEDIUM",
		req.Landmark,
		req.Latitude,
		req.Longitude,
		customerUserID,
	)
	if err != nil {
		return err
	}

	err = h.service.AutoDispatch(c.UserContext(), res.ID)
	if err != nil {
		return response.SendSuccess(c, fiber.StatusCreated, messages.ErrTicketReportDispatchFail, fiber.Map{
			"ticket": res,
		})
	}

	// Fetch updated ticket containing assigned technician ID
	updatedTicket, err := h.service.GetByID(c.UserContext(), res.ID)
	if err != nil {
		updatedTicket = res
	}

	return response.SendSuccess(c, fiber.StatusCreated, messages.SuccessTicketCreated, fiber.Map{
		"ticket": updatedTicket,
	})
}

func (h *TicketHandler) AutoDispatch(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.service.AutoDispatch(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, messages.SuccessTicketDispatched, nil)
}

func (h *TicketHandler) Start(c *fiber.Ctx) error {
	id := c.Params("id")

	techUserID, ok := c.Locals("userID").(string)
	if !ok {
		return exceptions.NewBadRequestError("Unauthorized", "UNAUTHORIZED")
	}

	var beforeBytes []byte
	fileHeader, err := c.FormFile("before_photo")
	if err == nil && fileHeader != nil {
		file, err := fileHeader.Open()
		if err == nil {
			defer file.Close()
			beforeBytes, _ = io.ReadAll(file)
		}
	}

	err = h.service.StartTicket(c.UserContext(), id, techUserID, beforeBytes)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, messages.SuccessTicketStarted, nil)
}

func (h *TicketHandler) Complete(c *fiber.Ctx) error {
	id := c.Params("id")

	otp := c.FormValue("otp")
	if otp == "" {
		return exceptions.NewBadRequestError(messages.ErrOTPRequired, "MISSING_OTP")
	}

	techUserID, ok := c.Locals("userID").(string)
	if !ok {
		return exceptions.NewBadRequestError("Unauthorized", "UNAUTHORIZED")
	}

	var afterBytes []byte
	fileHeader, err := c.FormFile("after_photo")
	if err == nil && fileHeader != nil {
		file, err := fileHeader.Open()
		if err == nil {
			defer file.Close()
			afterBytes, _ = io.ReadAll(file)
		}
	}

	err = h.service.CompleteTicket(c.UserContext(), id, techUserID, otp, afterBytes)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, messages.SuccessTicketCompleted, nil)
}

func (h *TicketHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Ticket retrieved successfully", res)
}

func (h *TicketHandler) DirectAssign(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		TechnicianID string `json:"technician_id"`
	}
	if err := c.BodyParser(&req); err != nil || req.TechnicianID == "" {
		return exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	err := h.service.DirectAssign(c.UserContext(), id, req.TechnicianID)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Ticket assigned directly to technician successfully.", nil)
}

func (h *TicketHandler) GetByTechnician(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return exceptions.NewBadRequestError("Unauthorized", "UNAUTHORIZED")
	}
	userID := userIDVal.(string)
	statusFilter := c.Query("status")

	res, err := h.service.GetTicketsByTechnicianUser(c.UserContext(), userID, statusFilter)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Technician tickets retrieved successfully.", res)
}

func (h *TicketHandler) GetByCustomer(c *fiber.Ctx) error {
	userIDVal := c.Locals("userID")
	if userIDVal == nil {
		return exceptions.NewBadRequestError("Unauthorized", "UNAUTHORIZED")
	}
	userID := userIDVal.(string)

	res, err := h.service.GetTicketsByCustomerUser(c.UserContext(), userID)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Customer tickets retrieved successfully.", res)
}

func (h *TicketHandler) Review(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Rating  int      `json:"rating"`
		Tags    []string `json:"tags"`
		Comment string   `json:"comment"`
	}
	if err := c.BodyParser(&req); err != nil {
		return exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	err := h.service.SubmitFeedback(c.UserContext(), id, req.Rating, req.Tags, req.Comment)
	if err != nil {
		return err
	}

	return response.SendSuccess(c, fiber.StatusOK, "Ticket review submitted successfully. Technician rating updated.", nil)
}
