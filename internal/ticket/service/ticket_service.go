package service

import (
	"context"
	"fmt"
	"fsm-backend/exceptions"
	"fsm-backend/helpers"
	"fsm-backend/internal/domain"
	"fsm-backend/messages"
	"math/rand"
	"time"
)

type TicketService interface {
	ReportTicket(ctx context.Context, title, desc, cat, priority, landmark string, lat, lon float64, customerUserID string) (*domain.Ticket, error)
	AutoDispatch(ctx context.Context, ticketID string) error
	StartTicket(ctx context.Context, ticketID string, technicianUserID string, beforePhoto []byte) error
	CompleteTicket(ctx context.Context, ticketID string, technicianUserID string, otp string, afterPhoto []byte) error
	SubmitFeedback(ctx context.Context, ticketID string, rating int, tags []string, comment string) error
	GetByID(ctx context.Context, id string) (*domain.Ticket, error)
	GetTimelineLogs(ctx context.Context, ticketID string) ([]*domain.TicketLog, error)
}

type ticketService struct {
	ticketRepo domain.TicketRepository
	techRepo   domain.TechnicianRepository
	custRepo   domain.CustomerRepository
}

func NewTicketService(
	ticketRepo domain.TicketRepository,
	techRepo domain.TechnicianRepository,
	custRepo domain.CustomerRepository,
) TicketService {
	return &ticketService{
		ticketRepo: ticketRepo,
		techRepo:   techRepo,
		custRepo:   custRepo,
	}
}

func (s *ticketService) ReportTicket(ctx context.Context, title, desc, cat, priority, landmark string, lat, lon float64, customerUserID string) (*domain.Ticket, error) {
	if title == "" || desc == "" || cat == "" {
		return nil, exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	// 1. Fetch customer details
	cust, err := s.custRepo.GetByUserID(ctx, customerUserID)
	if err != nil || cust == nil {
		return nil, exceptions.NewNotFoundError(messages.ErrCustomerNotFound, "CUSTOMER_NOT_FOUND")
	}

	// 2. Generate ticket credentials
	rand.Seed(time.Now().UnixNano())
	ticketNum := fmt.Sprintf("TK-%d", rand.Intn(90000)+10000)
	otp := helpers.GenerateRandomOTP()

	ticket := domain.Ticket{
		TicketNumber: ticketNum,
		CustomerID:   cust.ID,
		Title:        title,
		Description:  desc,
		Category:     cat,
		Priority:     priority,
		Status:       "REPORTED",
		Landmark:     &landmark,
		Latitude:     lat,
		Longitude:    lon,
		OTPCode:      otp,
	}

	// 3. Save ticket
	err = s.ticketRepo.Create(ctx, &ticket)
	if err != nil {
		return nil, exceptions.NewInternalServerError("Failed to create ticket")
	}

	// Log initial timeline action
	_ = s.ticketRepo.LogProgress(ctx, &domain.TicketLog{
		TicketID:    ticket.ID,
		NewStatus:   "REPORTED",
		Action:      "TICKET_REPORTED",
		Notes:       &desc,
		PerformedBy: customerUserID,
	})

	return &ticket, nil
}

func (s *ticketService) AutoDispatch(ctx context.Context, ticketID string) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil || ticket == nil {
		return exceptions.NewNotFoundError(messages.ErrTicketNotFound, "TICKET_NOT_FOUND")
	}

	// 1. Fetch nearest matching technician (radius 50km = 50000m)
	techs, _, err := s.techRepo.FindNearestMatching(ctx, ticket.Longitude, ticket.Latitude, ticket.Category, 50000.0)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	if len(techs) == 0 {
		return exceptions.NewNotFoundError(messages.ErrDispatchFailed, "NO_MATCHING_TECHNICIANS")
	}

	matchedTech := techs[0]

	// 2. Assign and update status to DISPATCHED
	err = s.ticketRepo.AssignTechnician(ctx, ticketID, matchedTech.ID)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	// Log audit trail
	notes := fmt.Sprintf("System auto-dispatched ticket to technician ID %s", matchedTech.ID)
	_ = s.ticketRepo.LogProgress(ctx, &domain.TicketLog{
		TicketID:    ticketID,
		NewStatus:   "DISPATCHED",
		Action:      "TICKET_DISPATCHED",
		Notes:       &notes,
		PerformedBy: "00000000-0000-0000-0000-000000000000", // System UUID
	})

	return nil
}

func (s *ticketService) StartTicket(ctx context.Context, ticketID string, technicianUserID string, beforePhoto []byte) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil || ticket == nil {
		return exceptions.NewNotFoundError(messages.ErrTicketNotFound, "TICKET_NOT_FOUND")
	}

	// Verify technician profile matches
	tech, err := s.techRepo.GetByUserID(ctx, technicianUserID)
	if err != nil || tech == nil {
		return exceptions.NewNotFoundError(messages.ErrTechnicianNotFound, "TECHNICIAN_NOT_FOUND")
	}

	if ticket.TechnicianID == nil || *ticket.TechnicianID != tech.ID {
		return exceptions.NewBadRequestError(messages.ErrUnauthorizedTech, "UNAUTHORIZED_ACTION")
	}

	// Mock photo saving logic
	photoURL := fmt.Sprintf("http://localhost:8080/uploads/before_%s.webp", ticketID)
	if len(beforePhoto) > 0 {
		_, _ = helpers.CompressImage(beforePhoto, "webp")
	}

	err = s.ticketRepo.StartTicket(ctx, ticketID, photoURL)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	// Log progress
	notes := "Technician arrived on-site and initiated repair tasks"
	_ = s.ticketRepo.LogProgress(ctx, &domain.TicketLog{
		TicketID:    ticketID,
		NewStatus:   "IN_PROGRESS",
		Action:      "TICKET_STARTED",
		Notes:       &notes,
		PerformedBy: technicianUserID,
	})

	return nil
}

func (s *ticketService) CompleteTicket(ctx context.Context, ticketID string, technicianUserID string, otp string, afterPhoto []byte) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil || ticket == nil {
		return exceptions.NewNotFoundError(messages.ErrTicketNotFound, "TICKET_NOT_FOUND")
	}

	// Verify OTP
	if ticket.OTPCode != otp {
		return exceptions.NewBadRequestError(messages.ErrInvalidOTP, "INVALID_OTP")
	}

	// Verify technician matches
	tech, err := s.techRepo.GetByUserID(ctx, technicianUserID)
	if err != nil || tech == nil {
		return exceptions.NewNotFoundError(messages.ErrTechnicianNotFound, "TECHNICIAN_NOT_FOUND")
	}

	if ticket.TechnicianID == nil || *ticket.TechnicianID != tech.ID {
		return exceptions.NewBadRequestError(messages.ErrUnauthorizedTech, "UNAUTHORIZED_ACTION")
	}

	photoURL := fmt.Sprintf("http://localhost:8080/uploads/after_%s.webp", ticketID)
	if len(afterPhoto) > 0 {
		_, _ = helpers.CompressImage(afterPhoto, "webp")
	}

	err = s.ticketRepo.CompleteTicket(ctx, ticketID, photoURL)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	// Log progress
	notes := "Technician resolved issue on-site and submitted verification details"
	_ = s.ticketRepo.LogProgress(ctx, &domain.TicketLog{
		TicketID:    ticketID,
		NewStatus:   "COMPLETED",
		Action:      "TICKET_COMPLETED",
		Notes:       &notes,
		PerformedBy: technicianUserID,
	})

	return nil
}

func (s *ticketService) SubmitFeedback(ctx context.Context, ticketID string, rating int, tags []string, comment string) error {
	if rating < 1 || rating > 5 {
		return exceptions.NewBadRequestError("Rating must be between 1 and 5 stars", "INVALID_RATING")
	}

	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil || ticket == nil {
		return exceptions.NewNotFoundError(messages.ErrTicketNotFound, "TICKET_NOT_FOUND")
	}

	return s.ticketRepo.SubmitReview(ctx, ticketID, rating, tags, comment)
}

func (s *ticketService) GetByID(ctx context.Context, id string) (*domain.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalServerError(err.Error())
	}
	if ticket == nil {
		return nil, exceptions.NewNotFoundError(messages.ErrTicketNotFound, "TICKET_NOT_FOUND")
	}
	return ticket, nil
}

func (s *ticketService) GetTimelineLogs(ctx context.Context, ticketID string) ([]*domain.TicketLog, error) {
	return s.ticketRepo.GetProgressLogs(ctx, ticketID)
}
