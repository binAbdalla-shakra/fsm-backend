package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"fsm-backend/constants"
	"fsm-backend/internal/tracking/dto"
	"time"

	"github.com/redis/go-redis/v9"
)

type TrackingRedisRepository interface {
	CacheLocation(ctx context.Context, ping *dto.LocationPing) error
	PublishLocationUpdate(ctx context.Context, ticketID string, ping *dto.LocationPing) error
	SubscribeTicketChannel(ctx context.Context, ticketID string) *redis.PubSub
}

type trackingRedisRepository struct {
	rdb *redis.Client
}

func NewTrackingRedisRepository(rdb *redis.Client) TrackingRedisRepository {
	return &trackingRedisRepository{rdb: rdb}
}

func (r *trackingRedisRepository) CacheLocation(ctx context.Context, ping *dto.LocationPing) error {
	pipe := r.rdb.Pipeline()

	pipe.GeoAdd(ctx, constants.RedisGeoKeyPrefix, &redis.GeoLocation{
		Name:      ping.TechnicianID,
		Longitude: ping.Longitude,
		Latitude:  ping.Latitude,
	})

	hashKey := fmt.Sprintf("%s%s", constants.RedisStatusPrefix, ping.TechnicianID)
	pipe.HSet(ctx, hashKey, map[string]interface{}{
		"latitude":     ping.Latitude,
		"longitude":    ping.Longitude,
		"heading":      ping.Heading,
		"speed":        ping.Speed,
		"last_ping_at": time.Now().Format(time.RFC3339),
	})

	_, err := pipe.Exec(ctx)
	return err
}

func (r *trackingRedisRepository) PublishLocationUpdate(ctx context.Context, ticketID string, ping *dto.LocationPing) error {
	channel := fmt.Sprintf("%s%s", constants.RedisChannelPrefix, ticketID)
	
	bytes, err := json.Marshal(ping)
	if err != nil {
		return err
	}

	return r.rdb.Publish(ctx, channel, bytes).Err()
}

func (r *trackingRedisRepository) SubscribeTicketChannel(ctx context.Context, ticketID string) *redis.PubSub {
	channel := fmt.Sprintf("%s%s", constants.RedisChannelPrefix, ticketID)
	return r.rdb.Subscribe(ctx, channel)
}
