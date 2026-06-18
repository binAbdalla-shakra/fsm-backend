package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
)

func main() {
	log.Println("Reading .env configuration...")
	envBytes, err := os.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}

	var dbURL string
	lines := strings.Split(string(envBytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "DATABASE_URL=") {
			dbURL = strings.TrimPrefix(line, "DATABASE_URL=")
			dbURL = strings.Trim(dbURL, `"'`)
			break
		}
	}

	if dbURL == "" {
		log.Fatal("DATABASE_URL not found in .env")
	}

	ctx := context.Background()
	log.Printf("Connecting to database: %s", dbURL)
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	log.Println("Successfully connected to database.")

	// 1. Read down migration to drop all existing tables Cascade
	downPath := "db/migration/000001_init.down.sql"
	log.Printf("Reading down-migration file: %s", downPath)
	downSQL, err := os.ReadFile(downPath)
	if err != nil {
		log.Fatalf("Error reading down-migration: %v", err)
	}

	log.Println("Executing down-migration to drop existing tables...")
	_, err = conn.Exec(ctx, string(downSQL))
	if err != nil {
		log.Fatalf("Error executing down-migration: %v", err)
	}
	log.Println("Existing tables dropped successfully.")

	// 2. Read up migration to create new tables and schemas
	upPath := "db/migration/000001_init.up.sql"
	log.Printf("Reading up-migration file: %s", upPath)
	upSQL, err := os.ReadFile(upPath)
	if err != nil {
		log.Fatalf("Error reading up-migration: %v", err)
	}

	log.Println("Executing up-migration to create tables...")
	_, err = conn.Exec(ctx, string(upSQL))
	if err != nil {
		log.Fatalf("Error executing up-migration: %v", err)
	}
	log.Println("Up-migration completed successfully. All tables created.")

	// 3. Seed initial Roles & Permissions
	log.Println("Seeding initial Roles and Permissions...")
	err = seedRolesAndPermissions(ctx, conn)
	if err != nil {
		log.Fatalf("Error seeding roles and permissions: %v", err)
	}
	log.Println("Database migration and seeding completed successfully.")
}

func seedRolesAndPermissions(ctx context.Context, conn *pgx.Conn) error {
	// Let's seed permissions
	permissions := []struct {
		Code      string
		Name      string
		Group     string
		Desc      string
	}{
		{"ticket:create", "Create Ticket", "Ticket Management", "Allows creating a new ticket"},
		{"ticket:dispatch", "Dispatch Ticket", "Ticket Management", "Allows dispatching tickets to technicians"},
		{"ticket:update", "Update Ticket Status", "Ticket Management", "Allows changing ticket status"},
		{"tech:telemetry", "Submit Telemetry", "Technician Tracking", "Allows sending dynamic GPS telemetry updates"},
		{"system:config", "Configure System", "System Config", "Allows changing core system parameters"},
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, p := range permissions {
		_, err := tx.Exec(ctx, `
			INSERT INTO permissions (code, name, group_name, description)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (code) DO NOTHING`,
			p.Code, p.Name, p.Group, p.Desc,
		)
		if err != nil {
			return fmt.Errorf("failed to seed permission %s: %w", p.Code, err)
		}
	}

	roles := []struct {
		Name      string
		Level     int
		Desc      string
		PermCodes []string
	}{
		{"Super Admin", 100, "Super administrator with all permissions", []string{"ticket:create", "ticket:dispatch", "ticket:update", "tech:telemetry", "system:config"}},
		{"Admin", 80, "Administrator with management capabilities", []string{"ticket:create", "ticket:dispatch", "ticket:update", "tech:telemetry"}},
		{"Dispatcher", 50, "Operational dispatcher for ticketing and routing", []string{"ticket:create", "ticket:dispatch", "ticket:update"}},
		{"Technician", 20, "Field Service Technician", []string{"ticket:update", "tech:telemetry"}},
		{"Customer", 10, "Client subscriber reporting issues", []string{"ticket:create", "ticket:update"}},
	}

	for _, r := range roles {
		var roleID string
		err := tx.QueryRow(ctx, `
			INSERT INTO roles (name, hierarchy_level, description)
			VALUES ($1, $2, $3)
			ON CONFLICT (name) DO UPDATE SET hierarchy_level = EXCLUDED.hierarchy_level
			RETURNING id`,
			r.Name, r.Level, r.Desc,
		).Scan(&roleID)
		if err != nil {
			// If not returned (DO NOTHING case), query it
			err = tx.QueryRow(ctx, "SELECT id FROM roles WHERE name = $1", r.Name).Scan(&roleID)
			if err != nil {
				return fmt.Errorf("failed to retrieve role %s: %w", r.Name, err)
			}
		}

		for _, code := range r.PermCodes {
			var permID string
			err = tx.QueryRow(ctx, "SELECT id FROM permissions WHERE code = $1", code).Scan(&permID)
			if err != nil {
				return fmt.Errorf("failed to find permission %s for role %s: %w", code, r.Name, err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO role_permissions (role_id, permission_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING`,
				roleID, permID,
			)
			if err != nil {
				return fmt.Errorf("failed to link role %s and permission %s: %w", r.Name, code, err)
			}
		}
	}

	// Seed a default Admin User for testing
	var adminUserID string
	err = tx.QueryRow(ctx, `
		INSERT INTO users (phone, email, name, status, is_verified)
		VALUES ('+252617770300', 'admin@fsm.com', 'FSM Admin', 'ACTIVE', TRUE)
		ON CONFLICT (phone) DO UPDATE SET status = 'ACTIVE'
		RETURNING id`,
	).Scan(&adminUserID)
	if err != nil {
		// If already exists, select the ID
		err = tx.QueryRow(ctx, "SELECT id FROM users WHERE phone = '+252617770300'").Scan(&adminUserID)
		if err != nil {
			return fmt.Errorf("failed to find or create default admin user: %w", err)
		}
	}

	var adminRoleID string
	err = tx.QueryRow(ctx, "SELECT id FROM roles WHERE name = 'Admin' LIMIT 1").Scan(&adminRoleID)
	if err != nil {
		return fmt.Errorf("failed to find Admin role: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`,
		adminUserID, adminRoleID,
	)
	if err != nil {
		return fmt.Errorf("failed to link admin user to Admin role: %w", err)
	}

	return tx.Commit(ctx)
}
