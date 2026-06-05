package repository

import (
	"context"
	"errors"
	"fsm-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ticketRepository struct {
	db *pgxpool.Pool
}

func NewTicketRepository(db *pgxpool.Pool) domain.TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ctx context.Context, t *domain.Ticket) error {
	query := `
		INSERT INTO tickets (ticket_number, customer_id, title, description, category, priority, status, landmark, location, otp_code, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, ST_SetSRID(ST_MakePoint($9, $10), 4326), $11, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		t.TicketNumber, t.CustomerID, t.Title, t.Description, t.Category, t.Priority, t.Status, t.Landmark, t.Longitude, t.Latitude, t.OTPCode,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *ticketRepository) GetByID(ctx context.Context, id string) (*domain.Ticket, error) {
	query := `
		SELECT t.id, t.ticket_number, t.customer_id, t.technician_id, t.title, t.description, t.category, t.priority, t.status, t.landmark,
		       ST_Y(t.location) AS latitude, ST_X(t.location) AS longitude, t.before_photo_url, t.after_photo_url, t.otp_code,
		       t.rating_score, t.rating_tags, t.rating_comment, t.rejected_technician_ids, t.created_at, t.updated_at,
		       COALESCE(u.name, 'Customer'), u.phone, COALESCE(c.address, ''),
		       COALESCE(tu.name, ''), COALESCE(tu.phone, ''), COALESCE(tech.rating, 5.0),
		       COALESCE(ST_Y(tech.location), 0.0) AS tech_latitude, COALESCE(ST_X(tech.location), 0.0) AS tech_longitude
		FROM tickets t
		JOIN customers c ON t.customer_id = c.id
		JOIN users u ON c.user_id = u.id
		LEFT JOIN technicians tech ON t.technician_id = tech.id
		LEFT JOIN users tu ON tech.user_id = tu.id
		WHERE t.id = $1 AND t.deleted_at IS NULL
	`
	var t domain.Ticket
	var rejTechs []string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TicketNumber, &t.CustomerID, &t.TechnicianID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status, &t.Landmark,
		&t.Latitude, &t.Longitude, &t.BeforePhotoURL, &t.AfterPhotoURL, &t.OTPCode,
		&t.RatingScore, &t.RatingTags, &t.RatingComment, &rejTechs, &t.CreatedAt, &t.UpdatedAt,
		&t.CustomerName, &t.CustomerPhone, &t.Address,
		&t.TechnicianName, &t.TechnicianPhone, &t.TechnicianRating, &t.TechnicianLatitude, &t.TechnicianLongitude,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	t.RejectedTechnicianIDs = rejTechs
	t.ServiceType = t.Category
	return &t, nil
}

func (r *ticketRepository) GetByNumber(ctx context.Context, num string) (*domain.Ticket, error) {
	query := `
		SELECT t.id, t.ticket_number, t.customer_id, t.technician_id, t.title, t.description, t.category, t.priority, t.status, t.landmark,
		       ST_Y(t.location) AS latitude, ST_X(t.location) AS longitude, t.before_photo_url, t.after_photo_url, t.otp_code,
		       t.rating_score, t.rating_tags, t.rating_comment, t.rejected_technician_ids, t.created_at, t.updated_at,
		       COALESCE(u.name, 'Customer'), u.phone, COALESCE(c.address, ''),
		       COALESCE(tu.name, ''), COALESCE(tu.phone, ''), COALESCE(tech.rating, 5.0),
		       COALESCE(ST_Y(tech.location), 0.0) AS tech_latitude, COALESCE(ST_X(tech.location), 0.0) AS tech_longitude
		FROM tickets t
		JOIN customers c ON t.customer_id = c.id
		JOIN users u ON c.user_id = u.id
		LEFT JOIN technicians tech ON t.technician_id = tech.id
		LEFT JOIN users tu ON tech.user_id = tu.id
		WHERE t.ticket_number = $1 AND t.deleted_at IS NULL
	`
	var t domain.Ticket
	var rejTechs []string
	err := r.db.QueryRow(ctx, query, num).Scan(
		&t.ID, &t.TicketNumber, &t.CustomerID, &t.TechnicianID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status, &t.Landmark,
		&t.Latitude, &t.Longitude, &t.BeforePhotoURL, &t.AfterPhotoURL, &t.OTPCode,
		&t.RatingScore, &t.RatingTags, &t.RatingComment, &rejTechs, &t.CreatedAt, &t.UpdatedAt,
		&t.CustomerName, &t.CustomerPhone, &t.Address,
		&t.TechnicianName, &t.TechnicianPhone, &t.TechnicianRating, &t.TechnicianLatitude, &t.TechnicianLongitude,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	t.RejectedTechnicianIDs = rejTechs
	t.ServiceType = t.Category
	return &t, nil
}

func (r *ticketRepository) Update(ctx context.Context, t *domain.Ticket) error {
	query := `
		UPDATE tickets
		SET title = $1, description = $2, category = $3, priority = $4, landmark = $5, location = ST_SetSRID(ST_MakePoint($6, $7), 4326), updated_at = NOW()
		WHERE id = $8 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, t.Title, t.Description, t.Category, t.Priority, t.Landmark, t.Longitude, t.Latitude, t.ID)
	return err
}

func (r *ticketRepository) AssignTechnician(ctx context.Context, ticketID string, techID string) error {
	query := `
		UPDATE tickets
		SET technician_id = $1, status = 'AUTO_DISPATCHING', updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, techID, ticketID)
	return err
}

func (r *ticketRepository) UpdateStatus(ctx context.Context, ticketID string, status string, notes string, performedBy string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var oldStatus string
	err = tx.QueryRow(ctx, "SELECT status FROM tickets WHERE id = $1", ticketID).Scan(&oldStatus)
	if err != nil {
		return err
	}

	// Update ticket
	_, err = tx.Exec(ctx, "UPDATE tickets SET status = $1, updated_at = NOW() WHERE id = $2", status, ticketID)
	if err != nil {
		return err
	}

	// Insert status audit log
	logQuery := `
		INSERT INTO ticket_logs (ticket_id, old_status, new_status, action, notes, performed_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err = tx.Exec(ctx, logQuery, ticketID, oldStatus, status, "STATUS_UPDATE", notes, performedBy)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ticketRepository) StartTicket(ctx context.Context, ticketID string, beforePhotoURL string) error {
	query := `
		UPDATE tickets
		SET status = 'IN_PROGRESS', before_photo_url = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, beforePhotoURL, ticketID)
	return err
}

func (r *ticketRepository) CompleteTicket(ctx context.Context, ticketID string, afterPhotoURL string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Fetch tech profile
	var techID *string
	err = tx.QueryRow(ctx, "SELECT technician_id FROM tickets WHERE id = $1", ticketID).Scan(&techID)
	if err != nil {
		return err
	}

	// Update status
	_, err = tx.Exec(ctx, "UPDATE tickets SET status = 'COMPLETED', after_photo_url = $1, updated_at = NOW() WHERE id = $2", afterPhotoURL, ticketID)
	if err != nil {
		return err
	}

	if techID != nil {
		// Update tasks count and decrement workload
		_, err = tx.Exec(ctx, "UPDATE technicians SET workload = GREATEST(0, workload - 1), tasks_completed = tasks_completed + 1 WHERE id = $1", *techID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ticketRepository) SubmitReview(ctx context.Context, ticketID string, rating int, tags []string, comment string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Update ticket review
	query := `
		UPDATE tickets
		SET rating_score = $1, rating_tags = $2, rating_comment = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING technician_id
	`
	var techID *string
	err = tx.QueryRow(ctx, query, rating, tags, comment, ticketID).Scan(&techID)
	if err != nil {
		return err
	}

	if techID != nil {
		// 2. Re-calculate average rating for this technician across all their rated tickets
		var avgRating float64
		err = tx.QueryRow(ctx, `
			SELECT COALESCE(AVG(rating_score), 5.0)::double precision
			FROM tickets
			WHERE technician_id = $1 AND rating_score IS NOT NULL AND deleted_at IS NULL
		`, *techID).Scan(&avgRating)
		if err != nil {
			return err
		}

		// 3. Update the technician's rating
		_, err = tx.Exec(ctx, `
			UPDATE technicians
			SET rating = $1,
			    updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL
		`, avgRating, *techID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ticketRepository) LogProgress(ctx context.Context, l *domain.TicketLog) error {
	query := `
		INSERT INTO ticket_logs (ticket_id, old_status, new_status, action, notes, performed_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, l.TicketID, l.OldStatus, l.NewStatus, l.Action, l.Notes, l.PerformedBy).Scan(&l.ID, &l.CreatedAt)
}

func (r *ticketRepository) GetProgressLogs(ctx context.Context, ticketID string) ([]*domain.TicketLog, error) {
	query := `
		SELECT id, ticket_id, old_status, new_status, action, notes, performed_by, created_at
		FROM ticket_logs
		WHERE ticket_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.TicketLog
	for rows.Next() {
		var l domain.TicketLog
		err := rows.Scan(&l.ID, &l.TicketID, &l.OldStatus, &l.NewStatus, &l.Action, &l.Notes, &l.PerformedBy, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

func (r *ticketRepository) GetByTechnicianID(ctx context.Context, techID string, statusFilter string) ([]*domain.Ticket, error) {
	query := `
		SELECT id, ticket_number, customer_id, technician_id, title, description, category, priority, status, landmark,
		       ST_Y(location) AS latitude, ST_X(location) AS longitude, before_photo_url, after_photo_url, otp_code,
		       rating_score, rating_tags, rating_comment, created_at, updated_at
		FROM tickets
		WHERE technician_id = $1 AND deleted_at IS NULL
	`
	var rows pgx.Rows
	var err error
	if statusFilter != "" {
		query += " AND status = $2 ORDER BY created_at DESC"
		rows, err = r.db.Query(ctx, query, techID, statusFilter)
	} else {
		query += " ORDER BY created_at DESC"
		rows, err = r.db.Query(ctx, query, techID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		err := rows.Scan(
			&t.ID, &t.TicketNumber, &t.CustomerID, &t.TechnicianID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status, &t.Landmark,
			&t.Latitude, &t.Longitude, &t.BeforePhotoURL, &t.AfterPhotoURL, &t.OTPCode,
			&t.RatingScore, &t.RatingTags, &t.RatingComment, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func (r *ticketRepository) GetByCustomerID(ctx context.Context, custID string, statusFilter string) ([]*domain.Ticket, error) {
	query := `
		SELECT id, ticket_number, customer_id, technician_id, title, description, category, priority, status, landmark,
		       ST_Y(location) AS latitude, ST_X(location) AS longitude, before_photo_url, after_photo_url, otp_code,
		       rating_score, rating_tags, rating_comment, created_at, updated_at
		FROM tickets
		WHERE customer_id = $1 AND deleted_at IS NULL
	`
	var rows pgx.Rows
	var err error
	if statusFilter != "" {
		query += " AND status = $2 ORDER BY created_at DESC"
		rows, err = r.db.Query(ctx, query, custID, statusFilter)
	} else {
		query += " ORDER BY created_at DESC"
		rows, err = r.db.Query(ctx, query, custID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		err := rows.Scan(
			&t.ID, &t.TicketNumber, &t.CustomerID, &t.TechnicianID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status, &t.Landmark,
			&t.Latitude, &t.Longitude, &t.BeforePhotoURL, &t.AfterPhotoURL, &t.OTPCode,
			&t.RatingScore, &t.RatingTags, &t.RatingComment, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

func (r *ticketRepository) AcceptTicket(ctx context.Context, ticketID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var techID *string
	err = tx.QueryRow(ctx, "SELECT technician_id FROM tickets WHERE id = $1", ticketID).Scan(&techID)
	if err != nil || techID == nil {
		return errors.New("technician not assigned to ticket")
	}

	// Update ticket status to DISPATCHED
	_, err = tx.Exec(ctx, "UPDATE tickets SET status = 'DISPATCHED', updated_at = NOW() WHERE id = $1", ticketID)
	if err != nil {
		return err
	}

	// Increase workload for the technician
	_, err = tx.Exec(ctx, "UPDATE technicians SET workload = workload + 1 WHERE id = $1", *techID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ticketRepository) RejectTicket(ctx context.Context, ticketID string, techID string) error {
	// Reset technician_id, add technician to rejected_technician_ids array, and reset status to REPORTED
	query := `
		UPDATE tickets
		SET technician_id = NULL,
		    status = 'REPORTED',
		    rejected_technician_ids = array_append(rejected_technician_ids, $1::UUID),
		    updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, techID, ticketID)
	return err
}

func (r *ticketRepository) TransitTicket(ctx context.Context, ticketID string) error {
	query := `
		UPDATE tickets
		SET status = 'ON_THE_WAY', updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, ticketID)
	return err
}
