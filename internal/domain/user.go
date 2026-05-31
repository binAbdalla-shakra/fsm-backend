package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string     `json:"id"`
	Phone        string     `json:"phone"`
	Email        *string    `json:"email,omitempty"`
	PasswordHash *string    `json:"-"`
	Status       string     `json:"status"`
	IsVerified   bool       `json:"is_verified"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

type Role struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    *string    `json:"description,omitempty"`
	HierarchyLevel int        `json:"hierarchy_level"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Permission struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GroupName   string `json:"group_name"`
}

type Customer struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	AccountNumber string     `json:"account_number"`
	PlanType      string     `json:"plan_type"`
	CurrentSpeed  string     `json:"current_speed"`
	Address       *string    `json:"address,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	User          *User      `json:"user,omitempty"`
}

type Technician struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Status         string     `json:"status"`
	Workload       int        `json:"workload"`
	Skills         []string   `json:"skills"`
	Latitude       float64    `json:"latitude"`
	Longitude      float64    `json:"longitude"`
	ZoneAssignment *string    `json:"zone_assignment,omitempty"`
	Rating         float64    `json:"rating"`
	TasksCompleted int        `json:"tasks_completed"`
	ShiftStart     *string    `json:"shift_start,omitempty"`
	ShiftEnd       *string    `json:"shift_end,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	User           *User      `json:"user,omitempty"`
}

type UserSession struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	DeviceID         string    `json:"device_id"`
	DeviceName       string    `json:"device_name"`
	RefreshTokenHash string    `json:"-"`
	IPAddress        string    `json:"ip_address"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	Update(ctx context.Context, user *User) error
	AssignRole(ctx context.Context, userID string, roleName string) error
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

type CustomerRepository interface {
	Create(ctx context.Context, cust *Customer) error
	GetByID(ctx context.Context, id string) (*Customer, error)
	GetByUserID(ctx context.Context, userID string) (*Customer, error)
	GetByAccountNumber(ctx context.Context, accNum string) (*Customer, error)
}

type TechnicianRepository interface {
	Create(ctx context.Context, tech *Technician) error
	GetByID(ctx context.Context, id string) (*Technician, error)
	GetByUserID(ctx context.Context, userID string) (*Technician, error)
	UpdateStatusAndLocation(ctx context.Context, id string, status string, lat float64, lon float64) error
	UpdateWorkload(ctx context.Context, id string, change int) error
	FindNearestMatching(ctx context.Context, lon float64, lat float64, skill string, maxDistance float64) ([]*Technician, []float64, error)
}
