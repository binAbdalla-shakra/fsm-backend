package provider

import (
	"context"
	"fsm-backend/internal/domain"
	"log"
)

type hormuudSMSProvider struct{}

func NewHormuudSMSProvider() domain.SMSProvider {
	return &hormuudSMSProvider{}
}

func (p *hormuudSMSProvider) SendSMS(ctx context.Context, phone string, message string) error {
	// In production, we'd fire an HTTP client call to Hormuud SMS Gateway API.
	// For simulation, we log it to stdout.
	log.Printf("[Hormuud SMS API] Sending to %s: %s", phone, message)
	return nil
}
