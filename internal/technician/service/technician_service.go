package service

import (
	"context"
	"fsm-backend/exceptions"
	"fsm-backend/internal/domain"
	"fsm-backend/messages"
)

type TechnicianService interface {
	RegisterTechnician(ctx context.Context, u *domain.User, skills []string, zone *string, lat, lon float64) (*domain.Technician, error)
	GetByID(ctx context.Context, id string) (*domain.Technician, error)
	UpdateStatus(ctx context.Context, id string, status string, lat, lon float64) error
}

type technicianService struct {
	techRepo domain.TechnicianRepository
	userRepo domain.UserRepository
}

func NewTechnicianService(techRepo domain.TechnicianRepository, userRepo domain.UserRepository) TechnicianService {
	return &technicianService{
		techRepo: techRepo,
		userRepo: userRepo,
	}
}

func (s *technicianService) RegisterTechnician(ctx context.Context, u *domain.User, skills []string, zone *string, lat, lon float64) (*domain.Technician, error) {
	if u.Phone == "" || len(skills) == 0 {
		return nil, exceptions.NewBadRequestError(messages.ErrInvalidPayload, "INVALID_PAYLOAD")
	}

	// 1. Create Core User
	u.Status = "ACTIVE"
	u.IsVerified = true
	err := s.userRepo.Create(ctx, u)
	if err != nil {
		return nil, exceptions.NewConflictError("User with this phone number is already registered", "USER_CONFLICT")
	}

	_ = s.userRepo.AssignRole(ctx, u.ID, "TECHNICIAN")

	// 2. Create Technician Profile
	tech := domain.Technician{
		UserID:         u.ID,
		Status:         "OFFLINE",
		Skills:         skills,
		Latitude:       lat,
		Longitude:      lon,
		ZoneAssignment: zone,
		Rating:         5.0,
	}

	err = s.techRepo.Create(ctx, &tech)
	if err != nil {
		return nil, exceptions.NewInternalServerError("Failed to create technician profile record")
	}

	tech.User = u
	return &tech, nil
}

func (s *technicianService) GetByID(ctx context.Context, id string) (*domain.Technician, error) {
	tech, err := s.techRepo.GetByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalServerError(err.Error())
	}
	if tech == nil {
		return nil, exceptions.NewNotFoundError("Technician profile not found", "TECHNICIAN_NOT_FOUND")
	}

	u, err := s.userRepo.GetByID(ctx, tech.UserID)
	if err == nil {
		tech.User = u
	}
	return tech, nil
}

func (s *technicianService) UpdateStatus(ctx context.Context, id string, status string, lat, lon float64) error {
	if status != "ONLINE" && status != "OFFLINE" {
		return exceptions.NewBadRequestError(messages.ErrTechStatusInvalid, "INVALID_STATUS")
	}

	tech, err := s.techRepo.GetByID(ctx, id)
	if err != nil || tech == nil {
		return exceptions.NewNotFoundError(messages.ErrTechnicianNotFound, "TECHNICIAN_NOT_FOUND")
	}

	return s.techRepo.UpdateStatusAndLocation(ctx, id, status, lat, lon)
}
