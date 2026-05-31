package service

import (
	"context"
	"fsm-backend/internal/domain"
	"fsm-backend/internal/tracking/dto"
	"fsm-backend/internal/tracking/repository"

	"github.com/redis/go-redis/v9"
)

type TrackingService interface {
	ProcessTechPing(ctx context.Context, ping *dto.LocationPing, ticketID string) error
	GetSubscription(ctx context.Context, ticketID string) *redis.PubSub
}

type trackingService struct {
	redisRepo  repository.TrackingRedisRepository
	ticketRepo domain.TicketRepository
}

func NewTrackingService(redisRepo repository.TrackingRedisRepository, ticketRepo domain.TicketRepository) TrackingService {
	return &trackingService{
		redisRepo:  redisRepo,
		ticketRepo: ticketRepo,
	}
}

func (s *trackingService) ProcessTechPing(ctx context.Context, ping *dto.LocationPing, ticketID string) error {
	err := s.redisRepo.CacheLocation(ctx, ping)
	if err != nil {
		return err
	}

	if ticketID != "" {
		err = s.redisRepo.PublishLocationUpdate(ctx, ticketID, ping)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *trackingService) GetSubscription(ctx context.Context, ticketID string) *redis.PubSub {
	return s.redisRepo.SubscribeTicketChannel(ctx, ticketID)
}
