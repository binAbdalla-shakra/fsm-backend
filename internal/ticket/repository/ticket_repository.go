package repository

import (
	"context"
	"errors"
	"fsm-backend/internal/domain"

	"github.com/jackc/pgx/v5"
)

type ticketRepository struct {
	db *pgx.Conn
}

func NewTicketRepository(db *pgx.Conn) domain.TicketRepository {
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
		SELECT id, ticket_number, customer_id, technician_id, title, description, category, priority, status, landmark,
		       ST_Y(location) AS latitude, ST_X(location) AS longitude, before_photo_url, after_photo_url, otp_code,
		       rating_score, rating_tags, rating_comment, created_at, updated_at
		FROM tickets
		WHERE id = $1 AND deleted_at IS NULL
	`
	var t domain.Ticket
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TicketNumber, &t.CustomerID, &t.TechnicianID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status, &t.Landmark,
		&t.Latitude, &t.Longitude, &t.BeforePhotoURL, &t.AfterPhotoURL, &t.OTPCode,
		&t.RatingScore, &t.RatingTags, &t.RatingComment, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *ticketRepository) GetByNumber(ctx context.Context, num string) (*domain.Ticket, error) {
	query := `
		SELECT id, ticket_number, customer_id, technician_id, title, description, category, priority, status, landmark,
		       ST_Y(location) AS latitude, ST_X(location) AS longitude, before_photo_url, after_photo_url, otp_code,
		       rating_score, rating_tags, rating_comment, created_at, updated_at
		FROM tickets
		WHERE ticket_number = $1 AND deleted_at IS NULL
	`
	var t domain.Ticket
	err := r.db.QueryRow(ctx, query, num).Scan(
		&t.ID, &t.TicketNumber, &t.CustomerID, &t.TechnicianID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status, &t.Landmark,
		&t.Latitude, &t.Longitude, &t.BeforePhotoURL, &t.AfterPhotoURL, &t.OTPCode,
		&t.RatingScore, &t.RatingTags, &t.RatingComment, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Fetch current technician if exists
	var oldTechID *string
	err = tx.QueryRow(ctx, "SELECT technician_id FROM tickets WHERE id = $1", ticketID).Scan(&oldTechID)
	if err != nil {
		return err
	}

	if oldTechID != nil {
		// Decrease workload for previous technician
		_, err = tx.Exec(ctx, "UPDATE technicians SET workload = GREATEST(0, workload - 1) WHERE id = $1", *oldTechID)
		if err != nil {
			return err
		}
	}

	// Update ticket status
	_, err = tx.Exec(ctx, "UPDATE tickets SET technician_id = $1, status = 'DISPATCHED', updated_at = NOW() WHERE id = $2", techID, ticketID)
	if err != nil {
		return err
	}

	// Increase workload for matched technician
	_, err = tx.Exec(ctx, "UPDATE technicians SET workload = workload + 1 WHERE id = $1", techID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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
	query := `
		UPDATE tickets
		SET rating_score = $1, rating_tags = $2, rating_comment = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, rating, tags, comment, ticketID)
	return err
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
