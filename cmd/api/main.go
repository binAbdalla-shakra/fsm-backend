package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"fsm-backend/config"
	"fsm-backend/exceptions"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	log.Println("Starting Field Smart Management (FSM) Single-Tenant API Service...")

	// 1. Load System Configuration
	cfg := config.LoadConfig()
	ctx := context.Background()

	// 2. Establish Database Connections
	pgConn, err := config.ConnectPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("DATABASE CONNECTION ERROR: %v", err)
		os.Exit(1)
	}
	defer pgConn.Close()
	log.Println("Connected to PostgreSQL successfully.")

	rdb, err := config.ConnectRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Printf("REDIS CONNECTION ERROR: %v", err)
		os.Exit(1)
	}
	defer rdb.Close()
	log.Println("Connected to Redis successfully.")

	// 3. Initialize Dependency Injection (Decoupled to dependencies.go)
	deps := InitializeDependencies(pgConn, rdb)

	// 4. Setup Fiber Application
	app := fiber.New(fiber.Config{
		ErrorHandler: exceptions.FiberErrorHandler,
	})

	// Add Standard Middlewares
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Device-ID, X-Device-Name",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	app.Use(logger.New())
	app.Use(recover.New())

	// 5. Register Routes (Decoupled to routes.go)
	RegisterRoutes(app, deps)

	// 6. Start HTTP Server
	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("FSM Service Listening on HTTP %s", serverAddr)
	if err := app.Listen(serverAddr); err != nil {
		log.Fatalf("Fiber failed to serve: %v", err)
	}
}
