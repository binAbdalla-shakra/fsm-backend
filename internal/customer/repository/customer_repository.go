package repository

import (
	"context"
	"errors"
	"fsm-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type customerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) domain.CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(ctx context.Context, cust *domain.Customer) error {
	query := `
		INSERT INTO customers (user_id, account_number, plan_type, current_speed, address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, cust.UserID, cust.AccountNumber, cust.PlanType, cust.CurrentSpeed, cust.Address).Scan(
		&cust.ID,
		&cust.CreatedAt,
		&cust.UpdatedAt,
	)
}

func (r *customerRepository) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	query := `
		SELECT id, user_id, account_number, plan_type, current_speed, address, created_at, updated_at
		FROM customers
		WHERE id = $1 AND deleted_at IS NULL
	`
	var c domain.Customer
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.UserID,
		&c.AccountNumber,
		&c.PlanType,
		&c.CurrentSpeed,
		&c.Address,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *customerRepository) GetByUserID(ctx context.Context, userID string) (*domain.Customer, error) {
	query := `
		SELECT id, user_id, account_number, plan_type, current_speed, address, created_at, updated_at
		FROM customers
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	var c domain.Customer
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&c.ID,
		&c.UserID,
		&c.AccountNumber,
		&c.PlanType,
		&c.CurrentSpeed,
		&c.Address,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *customerRepository) GetByAccountNumber(ctx context.Context, accNum string) (*domain.Customer, error) {
	query := `
		SELECT id, user_id, account_number, plan_type, current_speed, address, created_at, updated_at
		FROM customers
		WHERE account_number = $1 AND deleted_at IS NULL
	`
	var c domain.Customer
	err := r.db.QueryRow(ctx, query, accNum).Scan(
		&c.ID,
		&c.UserID,
		&c.AccountNumber,
		&c.PlanType,
		&c.CurrentSpeed,
		&c.Address,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *customerRepository) GetStats(ctx context.Context, customerID string) (*domain.CustomerStats, error) {
	query := `
		SELECT 
			COUNT(*)::integer AS total,
			COUNT(*) FILTER (WHERE status != 'COMPLETED')::integer AS active,
			COUNT(*) FILTER (WHERE status = 'COMPLETED')::integer AS completed
		FROM tickets
		WHERE customer_id = $1 AND deleted_at IS NULL
	`
	var stats domain.CustomerStats
	err := r.db.QueryRow(ctx, query, customerID).Scan(&stats.TotalTickets, &stats.ActiveTickets, &stats.CompletedTickets)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
