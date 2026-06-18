package main

import (
	"context"
	"log"
	"fsm-backend/internal/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// RegisterRoutes maps application routes using the dependencies wrapper.
func RegisterRoutes(app *fiber.App, deps *AppDependencies) {
	// 1. Initialize DB structures if they do not exist
	ctx := context.Background()
	_, err := deps.DbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS company_settings (
			id SERIAL PRIMARY KEY,
			name VARCHAR(150) NOT NULL DEFAULT 'FSM Operations',
			email VARCHAR(150) NOT NULL DEFAULT 'info@fsmcorp.com',
			phone VARCHAR(50) NOT NULL DEFAULT '+1 555-0100',
			address TEXT NOT NULL DEFAULT '123 Main St, New York',
			logo_url TEXT DEFAULT '',
			sla_target NUMERIC(5, 2) DEFAULT 95.0,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO company_settings (id, name, email, phone, address, logo_url, sla_target)
		VALUES (1, 'FSM Operations', 'info@fsmcorp.com', '+1 555-0100', '123 Main St, New York', '', 95.0)
		ON CONFLICT (id) DO NOTHING;
	`)
	if err != nil {
		log.Printf("DB INIT ERROR (company_settings): %v", err)
	}

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

	// =========================================================================
	// FSM SYSTEM PORTAL HANDLERS (Isolated under /portal to prevent conflicts)
	// =========================================================================
	portal := api.Group("/portal")

	// CUSTOMERS CRUD
	portal.Get("/customers", deps.PortalHandler.GetCustomers)
	portal.Post("/customers", deps.PortalHandler.CreateCustomer)
	portal.Put("/customers/:id", deps.PortalHandler.UpdateCustomer)
	portal.Delete("/customers/:id", deps.PortalHandler.DeleteCustomer)

	// TECHNICIANS CRUD
	portal.Get("/technicians", deps.PortalHandler.GetTechnicians)
	portal.Post("/technicians", deps.PortalHandler.CreateTechnician)
	portal.Put("/technicians/:id", deps.PortalHandler.UpdateTechnician)
	portal.Delete("/technicians/:id", deps.PortalHandler.DeleteTechnician)

	// ROLES CRUD
	portal.Get("/roles", deps.PortalHandler.GetRoles)
	portal.Post("/roles", deps.PortalHandler.CreateRole)
	portal.Put("/roles/:id", deps.PortalHandler.UpdateRole)
	portal.Delete("/roles/:id", deps.PortalHandler.DeleteRole)

	// USERS CRUD
	portal.Get("/users", deps.PortalHandler.GetUsers)
	portal.Post("/users", deps.PortalHandler.CreateUser)
	portal.Put("/users/:id", deps.PortalHandler.UpdateUser)
	portal.Delete("/users/:id", deps.PortalHandler.DeleteUser)

	// ROLE PERMISSIONS MAPPING
	portal.Get("/role-permissions/:roleId", deps.PortalHandler.GetRolePermissions)
	portal.Post("/role-permissions/:roleId", deps.PortalHandler.SaveRolePermissions)

	// DASHBOARD STATS
	portal.Get("/fsm/dashboard-stats", deps.PortalHandler.GetDashboardStats)

	// COMPANY SETTINGS ENDPOINTS
	portal.Get("/company-settings", deps.PortalHandler.GetCompanySettings)
	portal.Post("/company-settings", deps.PortalHandler.UpdateCompanySettings)

	// PORTAL TICKETS ENDPOINTS
	portal.Get("/tickets", deps.PortalHandler.GetTickets)
	portal.Post("/tickets", deps.PortalHandler.CreateTicket)
	portal.Get("/technicians/locations", deps.PortalHandler.GetTechnicianLocations)
	portal.Post("/tickets/:id/complete", deps.PortalHandler.CompleteTicket)
	portal.Post("/tickets/:id/cancel", deps.PortalHandler.CancelTicket)

	// =========================================================================
	// END OF FSM SYSTEM PORTAL HANDLERS
	// =========================================================================

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
