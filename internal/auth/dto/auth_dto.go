package dto

type SendOTPRequest struct {
	Phone      string `json:"phone"`
	ActionType string `json:"action_type"` // SIGNUP, LOGIN
}

type SendOTPResponse struct {
	Message string `json:"message"`
	Phone   string `json:"phone"`
}

type SignUpPayload struct {
	Phone         string  `json:"phone"`
	Email         *string `json:"email,omitempty"`
	Name          string  `json:"name"`
	AccountNumber string  `json:"account_number"`
	Address       *string `json:"address,omitempty"`
	DeviceID      string  `json:"device_id"`
	DeviceName    string  `json:"device_name"`
}

type LoginPayload struct {
	Phone      string `json:"phone"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type VerifyOTPPayload struct {
	Phone      string `json:"phone"`
	Code       string `json:"code"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type RefreshTokenPayload struct {
	DeviceID     string `json:"device_id"`
	RefreshToken string `json:"refresh_token"`
}
