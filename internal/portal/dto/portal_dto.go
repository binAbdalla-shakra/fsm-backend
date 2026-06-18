package dto

import "time"

type CreatePortalTicketRequest struct {
	CustomerID    string  `json:"customer_id"`
	TechnicianID  string  `json:"technician_id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	SkillRequired string  `json:"skill_required"`
	Landmark      string  `json:"landmark"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
}

type PortalTicketResponse struct {
	ID             string    `json:"_id"`
	TicketNumber   string    `json:"ticket_number"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	Category       string    `json:"category"`
	Landmark       string    `json:"landmark"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	OTPCode        string    `json:"otp_code"`
	CreatedAt      time.Time `json:"created_at"`
	CustomerID     string    `json:"customer_id"`
	CustomerName   string    `json:"customer_name"`
	CustomerPhone  string    `json:"customer_phone"`
	TechnicianID   string    `json:"technician_id"`
	TechnicianName string    `json:"technician_name"`
}

type TechnicianLocationResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Phone      string  `json:"phone"`
	WorkStatus string  `json:"workStatus"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Zone       string  `json:"zone"`
}

type CreateCustomerRequest struct {
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	Status        string `json:"status"`
}

type CreateTechnicianRequest struct {
	Name           string `json:"name"`
	Skill          string `json:"skill"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	WorkStatus     string `json:"workStatus"`
	Status         string `json:"status"`
	ZoneAssignment string `json:"zone_assignment"`
}

type CreateRoleRequest struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type CreateUserRequest struct {
	Username string   `json:"username"`
	FullName string   `json:"fullName"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Status   string   `json:"status"`
	Roles    []string `json:"roles"`
}

type UpdateUserRequest struct {
	Username string   `json:"username"`
	FullName string   `json:"fullName"`
	Email    string   `json:"email"`
	Status   string   `json:"status"`
	Roles    []string `json:"roles"`
}

type UpdateCompanySettingsRequest struct {
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
	Address   string  `json:"address"`
	LogoURL   string  `json:"logo_url"`
	SLATarget float64 `json:"sla_target"`
}
