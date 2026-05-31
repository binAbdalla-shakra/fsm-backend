package dto

import "time"

type CreateTechnicianRequest struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Skills    []string `json:"skills"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
}

type UpdateStatusRequest struct {
	Status    string  `json:"status"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TechnicianResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Status     string    `json:"status"`
	Workload   int       `json:"workload"`
	Skills     []string  `json:"skills"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	LastPingAt time.Time `json:"last_ping_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
