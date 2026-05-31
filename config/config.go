package config

import (
	"os"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	RedisURL    string
}

func LoadConfig() *Config {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/fsm?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	return &Config{
		ServerPort:  serverPort,
		DatabaseURL: databaseURL,
		RedisURL:    redisURL,
	}
}
