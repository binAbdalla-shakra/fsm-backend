package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"fsm-backend/exceptions"
	"fsm-backend/internal/auth/dto"
	"fsm-backend/internal/domain"
	"fsm-backend/pkg/jwt"
	"time"
)

type AuthService interface {
	SendOTP(ctx context.Context, req *dto.SendOTPRequest) error
	VerifyOTPAndLogin(ctx context.Context, req *dto.VerifyOTPPayload, ipAddress string) (*domain.TokenResponse, error)
	SignUpCustomer(ctx context.Context, payload *dto.SignUpPayload) error
	RefreshSession(ctx context.Context, payload *dto.RefreshTokenPayload, ipAddress string) (*domain.TokenResponse, error)
	Logout(ctx context.Context, userID string, deviceID string) error
}

type authService struct {
	userRepo    domain.UserRepository
	custRepo    domain.CustomerRepository
	sessionRepo domain.SessionRepository
	otpRepo     domain.OTPRepository
	smsProvider domain.SMSProvider
	jwtSecret   string
}

func NewAuthService(
	userRepo domain.UserRepository,
	custRepo domain.CustomerRepository,
	sessionRepo domain.SessionRepository,
	otpRepo domain.OTPRepository,
	smsProvider domain.SMSProvider,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		custRepo:    custRepo,
		sessionRepo: sessionRepo,
		otpRepo:     otpRepo,
		smsProvider: smsProvider,
		jwtSecret:   "fsm-super-secret-key-12345", // standard default secret key
	}
}

func (s *authService) SendOTP(ctx context.Context, req *dto.SendOTPRequest) error {
	if req.Phone == "" || (req.ActionType != "SIGNUP" && req.ActionType != "LOGIN") {
		return exceptions.NewBadRequestError("Invalid phone number or action type", "INVALID_OTP_REQUEST")
	}

	// Generate standard 4-digit code (e.g. "4321" mock or randomized)
	code := "4321" // Simulation verification code
	hasher := sha256.New()
	hasher.Write([]byte(code))
	codeHash := hex.EncodeToString(hasher.Sum(nil))

	// Cache OTP
	err := s.otpRepo.SaveOTP(ctx, req.Phone, codeHash, req.ActionType, 5)
	if err != nil {
		return exceptions.NewInternalServerError(err.Error())
	}

	// Dispatch SMS
	message := fmt.Sprintf("Your FSM verification code is: %s. Valid for 5 minutes.", code)
	err = s.smsProvider.SendSMS(ctx, req.Phone, message)
	if err != nil {
		return exceptions.NewInternalServerError("Failed to send verification SMS via SMS Gateway")
	}

	return nil
}

func (s *authService) VerifyOTPAndLogin(ctx context.Context, req *dto.VerifyOTPPayload, ipAddress string) (*domain.TokenResponse, error) {
	if req.Phone == "" || req.Code == "" || req.DeviceID == "" {
		return nil, exceptions.NewBadRequestError("Required credentials payload fields are missing", "MISSING_VERIFICATION_FIELDS")
	}

	// 1. Verify OTP in database
	matched, err := s.otpRepo.VerifyOTP(ctx, req.Phone, req.Code, "LOGIN")
	if err != nil {
		return nil, exceptions.NewInternalServerError(err.Error())
	}
	if !matched {
		// Attempting signup fallback check
		signupMatched, err := s.otpRepo.VerifyOTP(ctx, req.Phone, req.Code, "SIGNUP")
		if err != nil || !signupMatched {
			return nil, exceptions.NewBadRequestError("Invalid or expired OTP verification code", "INVALID_OTP")
		}
	}

	// 2. Fetch User
	user, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err != nil {
		return nil, exceptions.NewInternalServerError(err.Error())
	}
	if user == nil {
		return nil, exceptions.NewNotFoundError("User profile is not registered. Please complete registration.", "USER_NOT_FOUND")
	}

	// Update status
	user.IsVerified = true
	user.Status = "ACTIVE"
	now := time.Now()
	user.LastLoginAt = &now
	_ = s.userRepo.Update(ctx, user)

	// Fetch roles and permissions
	roles, _ := s.userRepo.GetUserRoles(ctx, user.ID)
	permissions, _ := s.userRepo.GetUserPermissions(ctx, user.ID)

	primaryRole := "CUSTOMER"
	if len(roles) > 0 {
		primaryRole = roles[0]
	}

	// 3. Issue Access JWT
	accessToken, err := jwt.GenerateToken(user.ID, primaryRole, permissions, s.jwtSecret, 15) // 15 mins expiry
	if err != nil {
		return nil, exceptions.NewInternalServerError("Failed to sign access token JWT")
	}

	// 4. Issue Refresh Token
	rawRefreshToken := fmt.Sprintf("rf_%d_%s", time.Now().UnixNano(), user.ID)
	rfHasher := sha256.New()
	rfHasher.Write([]byte(rawRefreshToken))
	rfHash := hex.EncodeToString(rfHasher.Sum(nil))

	// Register device session in DB
	session := domain.UserSession{
		UserID:           user.ID,
		DeviceID:         req.DeviceID,
		DeviceName:       req.DeviceName,
		RefreshTokenHash: rfHash,
		IPAddress:        ipAddress,
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour), // 7 days expiry
	}

	err = s.sessionRepo.CreateSession(ctx, &session)
	if err != nil {
		return nil, exceptions.NewInternalServerError("Failed to create active session database record")
	}

	return &domain.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

func (s *authService) SignUpCustomer(ctx context.Context, payload *dto.SignUpPayload) error {
	if payload.Phone == "" || payload.Name == "" || payload.AccountNumber == "" {
		return exceptions.NewBadRequestError("Missing signup information fields", "INVALID_SIGNUP_PAYLOAD")
	}

	// 1. Create Core User profile
	user := domain.User{
		Phone:      payload.Phone,
		Email:      payload.Email,
		Status:     "PENDING_VERIFICATION",
		IsVerified: false,
	}

	err := s.userRepo.Create(ctx, &user)
	if err != nil {
		return exceptions.NewConflictError("Phone number or email is already registered", "USER_ALREADY_EXISTS")
	}

	// Assign role
	_ = s.userRepo.AssignRole(ctx, user.ID, "CUSTOMER")

	// 2. Create Customer Profile details
	customer := domain.Customer{
		UserID:        user.ID,
		AccountNumber: payload.AccountNumber,
		PlanType:      "STANDARD",
		CurrentSpeed:  "100 Mbps",
		Address:       payload.Address,
	}

	err = s.custRepo.Create(ctx, &customer)
	if err != nil {
		return exceptions.NewConflictError("Customer account number already mapped", "CUSTOMER_CONFLICT")
	}

	// 3. Immediately trigger Verification OTP SMS
	sendReq := dto.SendOTPRequest{
		Phone:      payload.Phone,
		ActionType: "SIGNUP",
	}
	return s.SendOTP(ctx, &sendReq)
}

func (s *authService) RefreshSession(ctx context.Context, payload *dto.RefreshTokenPayload, ipAddress string) (*domain.TokenResponse, error) {
	if payload.RefreshToken == "" || payload.DeviceID == "" {
		return nil, exceptions.NewBadRequestError("Missing refresh token or device header info", "INVALID_REFRESH_REQUEST")
	}

	// Hash input refresh token
	hasher := sha256.New()
	hasher.Write([]byte(payload.RefreshToken))
	rfHash := hex.EncodeToString(hasher.Sum(nil))

	// Find active session
	session, err := s.sessionRepo.GetSessionByRefreshToken(ctx, rfHash)
	if err != nil {
		return nil, exceptions.NewInternalServerError(err.Error())
	}
	if session == nil {
		// Potential breach: token has already been rotated or is falsified. Delete all sessions!
		return nil, exceptions.NewBadRequestError("Session has expired or token is already invalidated", "UNAUTHORIZED_REFRESH")
	}

	if session.ExpiresAt.Before(time.Now()) {
		_ = s.sessionRepo.DeleteSession(ctx, session.UserID, payload.DeviceID)
		return nil, exceptions.NewBadRequestError("Session refresh token expired. Please login again.", "EXPIRED_SESSION")
	}

	// 1. Fetch User roles & permissions
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil || user == nil {
		return nil, exceptions.NewBadRequestError("User not found", "USER_NOT_FOUND")
	}

	roles, _ := s.userRepo.GetUserRoles(ctx, user.ID)
	permissions, _ := s.userRepo.GetUserPermissions(ctx, user.ID)

	primaryRole := "CUSTOMER"
	if len(roles) > 0 {
		primaryRole = roles[0]
	}

	// 2. Generate new Access JWT
	newAccessToken, err := jwt.GenerateToken(user.ID, primaryRole, permissions, s.jwtSecret, 15)
	if err != nil {
		return nil, exceptions.NewInternalServerError("Failed to sign access token")
	}

	// 3. Rotate Refresh Token (RTR)
	newRawRefreshToken := fmt.Sprintf("rf_%d_%s", time.Now().UnixNano(), user.ID)
	newHasher := sha256.New()
	newHasher.Write([]byte(newRawRefreshToken))
	newRfHash := hex.EncodeToString(newHasher.Sum(nil))

	session.RefreshTokenHash = newRfHash
	session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	session.IPAddress = ipAddress

	err = s.sessionRepo.UpdateSession(ctx, session)
	if err != nil {
		return nil, exceptions.NewInternalServerError("Failed to rotate refresh token session")
	}

	return &domain.TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRawRefreshToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID string, deviceID string) error {
	return s.sessionRepo.DeleteSession(ctx, userID, deviceID)
}
