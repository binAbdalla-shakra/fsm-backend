package dto

import "time"

type CreateTicketRequest struct {
	CustomerID    string  `json:"customer_id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	SkillRequired string  `json:"skill_required"`
	Landmark      string  `json:"landmark"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
}

type StartTicketRequest struct {
	TechnicianID   string `json:"technician_id"`
	BeforePhotoRaw []byte `json:"before_photo_raw"`
}

type CompleteTicketRequest struct {
	OTP           string `json:"otp"`
	AfterPhotoRaw []byte `json:"after_photo_raw"`
}

type DispatchResponse struct {
	TicketID     string  `json:"ticket_id"`
	TechnicianID string  `json:"technician_id"`
	TechName     string  `json:"tech_name"`
	TechPhone    string  `json:"tech_phone"`
	Distance     float64 `json:"distance_meters"`
}

type TicketResponse struct {
	ID             string    `json:"id"`
	CustomerID     string    `json:"customer_id"`
	TechnicianID   *string   `json:"technician_id,omitempty"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	SkillRequired  string    `json:"skill_required"`
	Status         string    `json:"status"`
	Landmark       string    `json:"landmark"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	OTP            string    `json:"otp"`
	BeforePhotoURL *string   `json:"before_photo_url,omitempty"`
	AfterPhotoURL  *string   `json:"after_photo_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
