package main

import (
	"fsm-backend/config"
	"fsm-backend/internal/domain"
	authHandler "fsm-backend/internal/auth/handler"
	authRepo "fsm-backend/internal/auth/repository"
	authService "fsm-backend/internal/auth/service"
	customerHandler "fsm-backend/internal/customer/handler"
	customerRepo "fsm-backend/internal/customer/repository"
	customerService "fsm-backend/internal/customer/service"
	smsProvider "fsm-backend/internal/notification/provider"
	techHandler "fsm-backend/internal/technician/handler"
	techRepo "fsm-backend/internal/technician/repository"
	techService "fsm-backend/internal/technician/service"
	ticketHandler "fsm-backend/internal/ticket/handler"
	ticketRepo "fsm-backend/internal/ticket/repository"
	ticketService "fsm-backend/internal/ticket/service"
	trackHandler "fsm-backend/internal/tracking/handler"
	trackRepo "fsm-backend/internal/tracking/repository"
	trackService "fsm-backend/internal/tracking/service"
	userRepo "fsm-backend/internal/user/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// AppDependencies holds pointers to all the initialised domain handlers.
type AppDependencies struct {
	Config            *config.Config
	SessionRepository domain.SessionRepository
	AuthHandler       *authHandler.AuthHandler
	CustomerHandler   *customerHandler.CustomerHandler
	TechnicianHandler *techHandler.TechnicianHandler
	TicketHandler     *ticketHandler.TicketHandler
	TrackingHandler   *trackHandler.TrackingHandler
}

// InitializeDependencies resolves repository, service, and handler layers.
func InitializeDependencies(pgConn *pgxpool.Pool, rdb *redis.Client) *AppDependencies {
	cfg := config.LoadConfig()

	// Repositories
	uRepository := userRepo.NewUserRepository(pgConn)
	custRepository := customerRepo.NewCustomerRepository(pgConn)
	sessionRepository := authRepo.NewSessionRepository(pgConn)
	otpRepository := authRepo.NewOTPRepository(pgConn)
	tRepository := techRepo.NewTechnicianRepository(pgConn)
	tktRepository := ticketRepo.NewTicketRepository(pgConn)

	// SMS Provider adapter
	sms := smsProvider.NewHormuudSMSProvider()

	// Services
	aService := authService.NewAuthService(
		uRepository,
		custRepository,
		sessionRepository,
		otpRepository,
		sms,
	)
	custService := customerService.NewCustomerService(custRepository)
	tService := techService.NewTechnicianService(tRepository, uRepository)
	tktService := ticketService.NewTicketService(tktRepository, tRepository, custRepository)

	trRedisRepo := trackRepo.NewTrackingRedisRepository(rdb)
	trService := trackService.NewTrackingService(trRedisRepo, tktRepository)

	// Handlers
	aHand := authHandler.NewAuthHandler(aService)
	custHand := customerHandler.NewCustomerHandler(custService)
	tHand := techHandler.NewTechnicianHandler(tService)
	tktHand := ticketHandler.NewTicketHandler(tktService)
	trHand := trackHandler.NewTrackingHandler(trService)

	return &AppDependencies{
		Config:            cfg,
		SessionRepository: sessionRepository,
		AuthHandler:       aHand,
		CustomerHandler:   custHand,
		TechnicianHandler: tHand,
		TicketHandler:     tktHand,
		TrackingHandler:   trHand,
	}
}
