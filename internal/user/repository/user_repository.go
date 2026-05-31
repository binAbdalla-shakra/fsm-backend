package repository

import (
	"context"
	"errors"
	"fsm-backend/internal/domain"

	"github.com/jackc/pgx/v5"
)

type userRepository struct {
	db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (phone, email, password_hash, status, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, u.Phone, u.Email, u.PasswordHash, u.Status, u.IsVerified).Scan(
		&u.ID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, phone, email, status, is_verified, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	var u domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Phone,
		&u.Email,
		&u.Status,
		&u.IsVerified,
		&u.LastLoginAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, phone, email, status, is_verified, last_login_at, created_at, updated_at
		FROM users
		WHERE phone = $1 AND deleted_at IS NULL
	`
	var u domain.User
	err := r.db.QueryRow(ctx, query, phone).Scan(
		&u.ID,
		&u.Phone,
		&u.Email,
		&u.Status,
		&u.IsVerified,
		&u.LastLoginAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Update(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, status = $2, is_verified = $3, last_login_at = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, u.Email, u.Status, u.IsVerified, u.LastLoginAt, u.ID)
	return err
}

func (r *userRepository) AssignRole(ctx context.Context, userID string, roleName string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Fetch role ID
	var roleID string
	err = tx.QueryRow(ctx, "SELECT id FROM roles WHERE name = $1 LIMIT 1", roleName).Scan(&roleID)
	if err != nil {
		// If role doesn't exist, create it dynamically for robust setup
		hLevel := 10
		if roleName == "SUPER_ADMIN" {
			hLevel = 100
		} else if roleName == "ADMIN" {
			hLevel = 80
		} else if roleName == "DISPATCHER" {
			hLevel = 50
		} else if roleName == "TECHNICIAN" {
			hLevel = 20
		}
		createRoleQuery := `
			INSERT INTO roles (name, hierarchy_level, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			RETURNING id
		`
		err = tx.QueryRow(ctx, createRoleQuery, roleName, hLevel).Scan(&roleID)
		if err != nil {
			return err
		}
	}

	// Insert association
	assignQuery := `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err = tx.Exec(ctx, assignQuery, userID, roleID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *userRepository) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT p.code
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err == nil {
			permissions = append(permissions, code)
		}
	}
	return permissions, nil
}

func (r *userRepository) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT r.name
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			roles = append(roles, name)
		}
	}
	return roles, nil
}
