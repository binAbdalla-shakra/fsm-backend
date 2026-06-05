package main

import (
	"fsm-backend/internal/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// RegisterRoutes maps application routes using the dependencies wrapper.
func RegisterRoutes(app *fiber.App, deps *AppDependencies) {
	// Serve static docs directory for swagger spec
	app.Static("/docs", "./docs")

	// Mount Swagger UI
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL: "/docs/swagger.json",
	}))

	api := app.Group("/api/v1")

	// 1. Auth endpoints (Public)
	api.Post("/auth/signup", deps.AuthHandler.SignUp)
	api.Post("/auth/login", deps.AuthHandler.RequestOTP)
	api.Post("/auth/otp/verify", deps.AuthHandler.VerifyOTP)
	api.Post("/auth/refresh", deps.AuthHandler.Refresh)

	// Auth endpoints (Protected)
	authMiddleware := middleware.AuthRequired(deps.Config, deps.SessionRepository)
	api.Post("/auth/logout", authMiddleware, deps.AuthHandler.Logout)

	// 2. Customer resource endpoints (Protected)
	api.Get("/customers/profile", authMiddleware, deps.CustomerHandler.GetMe)
	api.Get("/customers/tickets", authMiddleware, deps.TicketHandler.GetByCustomer)

	// 3. Technician resource endpoints (Protected or Public for demo registration)
	api.Post("/technicians", deps.TechnicianHandler.Register)
	api.Get("/technicians/profile", authMiddleware, deps.TechnicianHandler.GetMe)
	api.Put("/technicians/status", authMiddleware, deps.TechnicianHandler.UpdateStatus)
	api.Get("/technicians/tickets", authMiddleware, deps.TicketHandler.GetByTechnician)

	// 4. Ticket resource endpoints (Protected)
	api.Post("/tickets", authMiddleware, deps.TicketHandler.Report)
	api.Get("/tickets/:id", authMiddleware, deps.TicketHandler.GetByID)
	api.Post("/tickets/:id/dispatch", authMiddleware, deps.TicketHandler.AutoDispatch)
	api.Post("/tickets/:id/assign", authMiddleware, deps.TicketHandler.DirectAssign)
	api.Post("/tickets/:id/start", authMiddleware, deps.TicketHandler.Start)
	api.Post("/tickets/:id/complete", authMiddleware, deps.TicketHandler.Complete)
	api.Post("/tickets/:id/review", authMiddleware, deps.TicketHandler.Review)
	api.Post("/tickets/:id/accept", authMiddleware, deps.TicketHandler.Accept)
	api.Post("/tickets/:id/reject", authMiddleware, deps.TicketHandler.Reject)
	api.Post("/tickets/:id/transit", authMiddleware, deps.TicketHandler.Transit)
	api.Get("/tickets/:id/logs", authMiddleware, deps.TicketHandler.GetLogs)

	// 5. WebSockets Telemetry Routes
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Upgrades technician connection to push GPS updates
	app.Get("/ws/tracking/tech/:tech_id", websocket.New(deps.TrackingHandler.TechTelemetryWS))
	// Upgrades customer connection to receive live technician route tracking
	app.Get("/ws/tracking/customer/:ticket_id", websocket.New(deps.TrackingHandler.CustomerTrackWS))
}
