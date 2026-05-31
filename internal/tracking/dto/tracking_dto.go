package dto

type LocationPing struct {
	TechnicianID string  `json:"technician_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Heading      float64 `json:"heading"`
	Speed        float64 `json:"speed"`
}

type ClientSubscribe struct {
	TicketID string `json:"ticket_id"`
}
