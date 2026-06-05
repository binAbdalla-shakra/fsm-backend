package domain

import (
	"context"
	"time"
)

type Ticket struct {
	ID                    string     `json:"id"`
	TicketNumber          string     `json:"ticket_number"`
	CustomerID            string     `json:"customer_id"`
	TechnicianID          *string    `json:"technician_id,omitempty"`
	Title                 string     `json:"title"`
	Description           string     `json:"description"`
	Category              string     `json:"category"`
	Priority              string     `json:"priority"`
	Status                string     `json:"status"`
	Landmark              *string    `json:"landmark,omitempty"`
	Latitude              float64    `json:"latitude"`
	Longitude             float64    `json:"longitude"`
	BeforePhotoURL        *string    `json:"before_photo_url,omitempty"`
	AfterPhotoURL         *string    `json:"after_photo_url,omitempty"`
	OTPCode               string     `json:"otp_code"`
	RatingScore           *int       `json:"rating_score,omitempty"`
	RatingTags            []string   `json:"rating_tags,omitempty"`
	RatingComment         *string    `json:"rating_comment,omitempty"`
	RejectedTechnicianIDs []string   `json:"rejected_technician_ids,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	DeletedAt             *time.Time `json:"-"`

	// Dynamically Populated Fields for Frontend Compatibility
	CustomerName     string  `json:"customerName,omitempty"`
	CustomerPhone    string  `json:"customerPhone,omitempty"`
	Address          string  `json:"address,omitempty"`
	ServiceType      string  `json:"serviceType,omitempty"`
	TechnicianName      string  `json:"technicianName,omitempty"`
	TechnicianPhone     string  `json:"technicianPhone,omitempty"`
	TechnicianRating    float64 `json:"technicianRating,omitempty"`
	TechnicianLatitude  float64 `json:"technicianLatitude,omitempty"`
	TechnicianLongitude float64 `json:"technicianLongitude,omitempty"`
}

type TicketLog struct {
	ID          string    `json:"id"`
	TicketID    string    `json:"ticket_id"`
	OldStatus   *string   `json:"old_status,omitempty"`
	NewStatus   string    `json:"new_status"`
	Action      string    `json:"action"`
	Notes       *string   `json:"notes,omitempty"`
	PerformedBy string    `json:"performed_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuditLog struct {
	ID         string    `json:"id"`
	UserID     *string   `json:"user_id,omitempty"`
	Action     string    `json:"action"`
	TableName  string    `json:"table_name"`
	RecordID   string    `json:"record_id"`
	OldValues  *string   `json:"old_values,omitempty"`
	NewValues  *string   `json:"new_values,omitempty"`
	IPAddress  *string   `json:"ip_address,omitempty"`
	UserAgent  *string   `json:"user_agent,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type TicketRepository interface {
	Create(ctx context.Context, ticket *Ticket) error
	GetByID(ctx context.Context, id string) (*Ticket, error)
	GetByNumber(ctx context.Context, num string) (*Ticket, error)
	Update(ctx context.Context, ticket *Ticket) error
	AssignTechnician(ctx context.Context, ticketID string, techID string) error
	UpdateStatus(ctx context.Context, ticketID string, status string, notes string, performedBy string) error
	StartTicket(ctx context.Context, ticketID string, beforePhotoURL string) error
	CompleteTicket(ctx context.Context, ticketID string, afterPhotoURL string) error
	SubmitReview(ctx context.Context, ticketID string, rating int, tags []string, comment string) error
	LogProgress(ctx context.Context, log *TicketLog) error
	GetProgressLogs(ctx context.Context, ticketID string) ([]*TicketLog, error)
	GetByTechnicianID(ctx context.Context, techID string, statusFilter string) ([]*Ticket, error)
	GetByCustomerID(ctx context.Context, custID string, statusFilter string) ([]*Ticket, error)
	AcceptTicket(ctx context.Context, ticketID string) error
	RejectTicket(ctx context.Context, ticketID string, techID string) error
	TransitTicket(ctx context.Context, ticketID string) error
}

type AuditRepository interface {
	WriteLog(ctx context.Context, log *AuditLog) error
}
