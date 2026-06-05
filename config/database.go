package config

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// ConnectPostgres initializes a connection pool to the PostgreSQL database.
func ConnectPostgres(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, databaseURL)
}

// ConnectRedis initializes and pings the Redis client.
func ConnectRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, err
	}

	return rdb, nil
}
