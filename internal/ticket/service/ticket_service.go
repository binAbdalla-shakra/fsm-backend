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
	DirectAssign(ctx context.Context, ticketRefOrID string, technicianID string) error
	GetTicketsByTechnician(ctx context.Context, technicianID string, statusFilter string) ([]*domain.Ticket, error)
	GetTicketsByCustomer(ctx context.Context, customerID string) ([]*domain.Ticket, error)
	GetTicketsByTechnicianUser(ctx context.Context, techUserID string, statusFilter string) ([]*domain.Ticket, error)
	GetTicketsByCustomerUser(ctx context.Context, custUserID string) ([]*domain.Ticket, error)
	AcceptTicket(ctx context.Context, ticketID string, technicianUserID string) error
	RejectTicket(ctx context.Context, ticketID string, technicianUserID string) error
	TransitTicket(ctx context.Context, ticketID string, technicianUserID string) error
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

	// 1. Fetch nearest matching technician with workload = 0 within 3km
	techs, _, err := s.techRepo.FindNearestMatching(ctx, ticket.Longitude, ticket.Latitude, ticket.Category, 3000.0, ticket.RejectedTechnicianIDs, false)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	var matchedTech *domain.Technician
	if len(techs) > 0 {
		matchedTech = techs[0]
	} else {
		// 2. Fallback: Fetch nearest matching technician with workload > 0 (lowest workload first) within 3km
		techs, _, err = s.techRepo.FindNearestMatching(ctx, ticket.Longitude, ticket.Latitude, ticket.Category, 3000.0, ticket.RejectedTechnicianIDs, true)
		if err != nil {
			return exceptions.NewInternalServerError(err.Error())
		}
		if len(techs) > 0 {
			matchedTech = techs[0]
		}
	}

	if matchedTech == nil {
		// No online technician found within 3km. Reset/leave status as REPORTED (Queued).
		_ = s.ticketRepo.UpdateStatus(ctx, ticketID, "REPORTED", "No available technicians within 3km. Queued.", "00000000-0000-0000-0000-000000000000")
		return exceptions.NewNotFoundError(messages.ErrDispatchFailed, "NO_MATCHING_TECHNICIANS")
	}

	// 3. Assign technician (status becomes AUTO_DISPATCHING)
	err = s.ticketRepo.AssignTechnician(ctx, ticketID, matchedTech.ID)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	// Log audit trail
	notes := fmt.Sprintf("System auto-dispatched ticket to technician ID %s (Pending Acceptance)", matchedTech.ID)
	_ = s.ticketRepo.LogProgress(ctx, &domain.TicketLog{
		TicketID:    ticketID,
		OldStatus:   &ticket.Status,
		NewStatus:   "AUTO_DISPATCHING",
		Action:      "TICKET_AUTO_ASSIGNED",
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

func (s *ticketService) DirectAssign(ctx context.Context, ticketRefOrID string, technicianID string) error {
	// 1. Fetch technician
	tech, err := s.techRepo.GetByID(ctx, technicianID)
	if err != nil || tech == nil {
		return exceptions.NewNotFoundError(messages.ErrTechnicianNotFound, "TECHNICIAN_NOT_FOUND")
	}

	// 2. Fetch ticket (try reference first, then ID)
	var ticket *domain.Ticket
	ticket, err = s.ticketRepo.GetByNumber(ctx, ticketRefOrID)
	if err != nil || ticket == nil {
		ticket, err = s.ticketRepo.GetByID(ctx, ticketRefOrID)
		if err != nil || ticket == nil {
			return exceptions.NewNotFoundError(messages.ErrTicketNotFound, "TICKET_NOT_FOUND")
		}
	}

	// 3. Assign
	err = s.ticketRepo.AssignTechnician(ctx, ticket.ID, technicianID)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	// Log progress timeline
	notes := fmt.Sprintf("Dispatcher assigned ticket directly to technician %s (Reference check: %s)", tech.ID, ticket.TicketNumber)
	_ = s.ticketRepo.LogProgress(ctx, &domain.TicketLog{
		TicketID:    ticket.ID,
		NewStatus:   "DISPATCHED",
		Action:      "TICKET_DIRECT_ASSIGN",
		Notes:       &notes,
		PerformedBy: "00000000-0000-0000-0000-000000000000",
	})

	return nil
}

func (s *ticketService) GetTicketsByTechnician(ctx context.Context, technicianID string, statusFilter string) ([]*domain.Ticket, error) {
	tech, err := s.techRepo.GetByID(ctx, technicianID)
	if err != nil || tech == nil {
		return nil, exceptions.NewNotFoundError(messages.ErrTechnicianNotFound, "TECHNICIAN_NOT_FOUND")
	}

	return s.ticketRepo.GetByTechnicianID(ctx, technicianID, statusFilter)
}

func (s *ticketService) GetTicketsByCustomer(ctx context.Context, customerID string) ([]*domain.Ticket, error) {
	cust, err := s.custRepo.GetByID(ctx, customerID)
	if err != nil || cust == nil {
		return nil, exceptions.NewNotFoundError(messages.ErrCustomerNotFound, "CUSTOMER_NOT_FOUND")
	}

	return s.ticketRepo.GetByCustomerID(ctx, customerID, "")
}

func (s *ticketService) GetTicketsByTechnicianUser(ctx context.Context, techUserID string, statusFilter string) ([]*domain.Ticket, error) {
	tech, err := s.techRepo.GetByUserID(ctx, techUserID)
	if err != nil || tech == nil {
		return nil, exceptions.NewNotFoundError(messages.ErrTechnicianNotFound, "TECHNICIAN_NOT_FOUND")
	}
	return s.ticketRepo.GetByTechnicianID(ctx, tech.ID, statusFilter)
}

func (s *ticketService) GetTicketsByCustomerUser(ctx context.Context, custUserID string) ([]*domain.Ticket, error) {
	cust, err := s.custRepo.GetByUserID(ctx, custUserID)
	if err != nil || cust == nil {
		return nil, exceptions.NewNotFoundError(messages.ErrCustomerNotFound, "CUSTOMER_NOT_FOUND")
	}
	return s.ticketRepo.GetByCustomerID(ctx, cust.ID, "")
}

func (s *ticketService) AcceptTicket(ctx context.Context, ticketID string, technicianUserID string) error {
	tech, err := s.techRepo.GetByUserID(ctx, technicianUserID)
	if err != nil || tech == nil {
		return exceptions.NewNotFoundError(messages.ErrTechnicianNotFound, "TECHNICIAN_NOT_FOUND")
	}

	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil || ticket == nil {
		return exceptions.NewNotFoundError(messages.ErrTicketNotFound, "TICKET_NOT_FOUND")
	}

	if ticket.TechnicianID == nil || *ticket.TechnicianID != tech.ID {
		return exceptions.NewBadRequestError(messages.ErrUnauthorizedTech, "UNAUTHORIZED_ACTION")
	}

	err = s.ticketRepo.AcceptTicket(ctx, ticketID)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	notes := fmt.Sprintf("Technician %s accepted the ticket", tech.ID)
	_ = s.ticketRepo.LogProgress(ctx, &domain.TicketLog{
		TicketID:    ticketID,
		OldStatus:   &ticket.Status,
		NewStatus:   "DISPATCHED",
		Action:      "TICKET_ACCEPTED",
		Notes:       &notes,
		PerformedBy: technicianUserID,
	})

	return nil
}

func (s *ticketService) RejectTicket(ctx context.Context, ticketID string, technicianUserID string) error {
	tech, err := s.techRepo.GetByUserID(ctx, technicianUserID)
	if err != nil || tech == nil {
		return exceptions.NewNotFoundError(messages.ErrTechnicianNotFound, "TECHNICIAN_NOT_FOUND")
	}

	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil || ticket == nil {
		return exceptions.NewNotFoundError(messages.ErrTicketNotFound, "TICKET_NOT_FOUND")
	}

	if ticket.TechnicianID == nil || *ticket.TechnicianID != tech.ID {
		return exceptions.NewBadRequestError(messages.ErrUnauthorizedTech, "UNAUTHORIZED_ACTION")
	}

	err = s.ticketRepo.RejectTicket(ctx, ticketID, tech.ID)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	notes := fmt.Sprintf("Technician %s rejected the ticket", tech.ID)
	_ = s.ticketRepo.LogProgress(ctx, &domain.TicketLog{
		TicketID:    ticketID,
		OldStatus:   &ticket.Status,
		NewStatus:   "REPORTED",
		Action:      "TICKET_REJECTED",
		Notes:       &notes,
		PerformedBy: technicianUserID,
	})

	// Re-run AutoDispatch to find the next nearest technician excluding the rejected ones
	_ = s.AutoDispatch(ctx, ticketID)

	return nil
}

func (s *ticketService) TransitTicket(ctx context.Context, ticketID string, technicianUserID string) error {
	tech, err := s.techRepo.GetByUserID(ctx, technicianUserID)
	if err != nil || tech == nil {
		return exceptions.NewNotFoundError(messages.ErrTechnicianNotFound, "TECHNICIAN_NOT_FOUND")
	}

	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil || ticket == nil {
		return exceptions.NewNotFoundError(messages.ErrTicketNotFound, "TICKET_NOT_FOUND")
	}

	if ticket.TechnicianID == nil || *ticket.TechnicianID != tech.ID {
		return exceptions.NewBadRequestError(messages.ErrUnauthorizedTech, "UNAUTHORIZED_ACTION")
	}

	err = s.ticketRepo.TransitTicket(ctx, ticketID)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	notes := "Technician marked themselves en route to customer location"
	_ = s.ticketRepo.LogProgress(ctx, &domain.TicketLog{
		TicketID:    ticketID,
		OldStatus:   &ticket.Status,
		NewStatus:   "ON_THE_WAY",
		Action:      "TICKET_TRANSIT",
		Notes:       &notes,
		PerformedBy: technicianUserID,
	})

	return nil
}
