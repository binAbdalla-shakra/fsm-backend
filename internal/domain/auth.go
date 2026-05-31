package domain

import (
	"context"
	"time"
)

type SignUpRequest struct {
	Phone         string  `json:"phone"`
	Email         *string `json:"email,omitempty"`
	Name          string  `json:"name"`
	AccountNumber string  `json:"account_number"`
	Address       *string `json:"address,omitempty"`
	DeviceID      string  `json:"device_id"`
	DeviceName    string  `json:"device_name"`
}

type LoginRequest struct {
	Phone      string `json:"phone"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type OTPVerifyRequest struct {
	Phone      string `json:"phone"`
	Code       string `json:"code"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type TechLoginVerifyRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type DeviceSession struct {
	ID         string    `json:"id"`
	DeviceID   string    `json:"device_id"`
	DeviceName string    `json:"device_name"`
	IPAddress  string    `json:"ip_address"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type OTPVerification struct {
	ID         string    `json:"id"`
	Phone      string    `json:"phone"`
	CodeHash   string    `json:"-"`
	ActionType string    `json:"action_type"`
	IsVerified bool      `json:"is_verified"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session *UserSession) error
	GetSession(ctx context.Context, userID string, deviceID string) (*UserSession, error)
	GetSessionByRefreshToken(ctx context.Context, tokenHash string) (*UserSession, error)
	DeleteSession(ctx context.Context, userID string, deviceID string) error
	DeleteAllUserSessions(ctx context.Context, userID string) error
	UpdateSession(ctx context.Context, session *UserSession) error
	GetUserSessions(ctx context.Context, userID string) ([]*DeviceSession, error)
}

type OTPRepository interface {
	SaveOTP(ctx context.Context, phone string, codeHash string, action string, durationMinutes int) error
	VerifyOTP(ctx context.Context, phone string, code string, action string) (bool, error)
}

type SMSProvider interface {
	SendSMS(ctx context.Context, phone string, message string) error
}
