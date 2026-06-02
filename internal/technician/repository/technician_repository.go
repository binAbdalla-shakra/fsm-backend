package repository

import (
	"context"
	"errors"
	"fsm-backend/internal/domain"

	"github.com/jackc/pgx/v5"
)

type technicianRepository struct {
	db *pgx.Conn
}

func NewTechnicianRepository(db *pgx.Conn) domain.TechnicianRepository {
	return &technicianRepository{db: db}
}

func (r *technicianRepository) Create(ctx context.Context, t *domain.Technician) error {
	query := `
		INSERT INTO technicians (user_id, status, workload, skills, location, zone_assignment, rating, tasks_completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326), $7, $8, 0, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, t.UserID, t.Status, t.Workload, t.Skills, t.Longitude, t.Latitude, t.ZoneAssignment, t.Rating).Scan(
		&t.ID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
}

func (r *technicianRepository) GetByID(ctx context.Context, id string) (*domain.Technician, error) {
	query := `
		SELECT id, user_id, status, workload, skills, ST_Y(location) AS latitude, ST_X(location) AS longitude,
		       zone_assignment, rating, tasks_completed, created_at, updated_at
		FROM technicians
		WHERE id = $1 AND deleted_at IS NULL
	`
	var t domain.Technician
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.UserID,
		&t.Status,
		&t.Workload,
		&t.Skills,
		&t.Latitude,
		&t.Longitude,
		&t.ZoneAssignment,
		&t.Rating,
		&t.TasksCompleted,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *technicianRepository) GetByUserID(ctx context.Context, userID string) (*domain.Technician, error) {
	query := `
		SELECT id, user_id, status, workload, skills, ST_Y(location) AS latitude, ST_X(location) AS longitude,
		       zone_assignment, rating, tasks_completed, created_at, updated_at
		FROM technicians
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	var t domain.Technician
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&t.ID,
		&t.UserID,
		&t.Status,
		&t.Workload,
		&t.Skills,
		&t.Latitude,
		&t.Longitude,
		&t.ZoneAssignment,
		&t.Rating,
		&t.TasksCompleted,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *technicianRepository) UpdateStatusAndLocation(ctx context.Context, id string, status string, lat float64, lon float64) error {
	query := `
		UPDATE technicians
		SET status = $1,
		    location = ST_SetSRID(ST_MakePoint($2, $3), 4326),
		    updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, status, lon, lat, id)
	return err
}

func (r *technicianRepository) UpdateWorkload(ctx context.Context, id string, change int) error {
	query := `
		UPDATE technicians
		SET workload = GREATEST(0, workload + $1),
		    updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, change, id)
	return err
}

func (r *technicianRepository) FindNearestMatching(ctx context.Context, lon float64, lat float64, skill string, maxDistance float64) ([]*domain.Technician, []float64, error) {
	query := `
		SELECT id, user_id, status, workload, skills, ST_Y(location) AS latitude, ST_X(location) AS longitude,
		       zone_assignment, rating, tasks_completed, created_at, updated_at,
		       ST_DistanceSphere(location, ST_SetSRID(ST_MakePoint($1, $2), 4326)) AS distance_meters
		FROM technicians
		WHERE status = 'ONLINE'
		  AND workload = 0
		  AND $3 = ANY(skills)
		  AND ST_DWithin(location::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $4)
		ORDER BY distance_meters ASC
		LIMIT 5
	`

	rows, err := r.db.Query(ctx, query, lon, lat, skill, maxDistance)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var technicians []*domain.Technician
	var distances []float64

	for rows.Next() {
		var t domain.Technician
		var dist float64
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Status,
			&t.Workload,
			&t.Skills,
			&t.Latitude,
			&t.Longitude,
			&t.ZoneAssignment,
			&t.Rating,
			&t.TasksCompleted,
			&t.CreatedAt,
			&t.UpdatedAt,
			&dist,
		)
		if err != nil {
			return nil, nil, err
		}
		technicians = append(technicians, &t)
		distances = append(distances, dist)
	}

	return technicians, distances, nil
}

func (r *technicianRepository) GetStats(ctx context.Context, technicianID string) (*domain.TechnicianStats, error) {
	query := `
		SELECT 
			COUNT(*)::integer AS total,
			COUNT(*) FILTER (WHERE status != 'COMPLETED')::integer AS active,
			COUNT(*) FILTER (WHERE status = 'COMPLETED')::integer AS completed,
			COALESCE(AVG(rating_score), 5.0)::double precision AS avg_rating
		FROM tickets
		WHERE technician_id = $1 AND deleted_at IS NULL
	`
	var stats domain.TechnicianStats
	err := r.db.QueryRow(ctx, query, technicianID).Scan(&stats.TotalAssigned, &stats.ActiveTickets, &stats.CompletedTickets, &stats.AverageRating)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
