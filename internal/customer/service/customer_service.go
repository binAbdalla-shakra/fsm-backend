package service

import (
	"context"
	"fsm-backend/exceptions"
	"fsm-backend/internal/domain"
)

type CustomerService interface {
	GetCustomerByID(ctx context.Context, id string) (*domain.Customer, error)
	GetCustomerByUserID(ctx context.Context, userID string) (*domain.Customer, error)
}

type customerService struct {
	repo domain.CustomerRepository
}

func NewCustomerService(repo domain.CustomerRepository) CustomerService {
	return &customerService{repo: repo}
}

func (s *customerService) GetCustomerByID(ctx context.Context, id string) (*domain.Customer, error) {
	if id == "" {
		return nil, exceptions.NewBadRequestError("Customer ID is required", "INVALID_CUSTOMER_ID")
	}

	cust, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalServerError(err.Error())
	}
	if cust == nil {
		return nil, exceptions.NewNotFoundError("Customer not found", "CUSTOMER_NOT_FOUND")
	}

	// Fetch and populate stats
	stats, err := s.repo.GetStats(ctx, cust.ID)
	if err == nil {
		cust.Stats = stats
	}

	return cust, nil
}

func (s *customerService) GetCustomerByUserID(ctx context.Context, userID string) (*domain.Customer, error) {
	if userID == "" {
		return nil, exceptions.NewBadRequestError("User ID is required", "INVALID_USER_ID")
	}

	cust, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, exceptions.NewInternalServerError(err.Error())
	}
	if cust == nil {
		return nil, exceptions.NewNotFoundError("Customer profile not found for this user", "CUSTOMER_PROFILE_NOT_FOUND")
	}

	// Fetch and populate stats
	stats, err := s.repo.GetStats(ctx, cust.ID)
	if err == nil {
		cust.Stats = stats
	}

	return cust, nil
}
