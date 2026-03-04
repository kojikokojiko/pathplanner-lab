package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pathplanner-lab/backend/internal/domain"
)

type pgExperimentRepository struct {
	pool *pgxpool.Pool
}

func NewExperimentRepository(pool *pgxpool.Pool) ExperimentRepository {
	return &pgExperimentRepository{pool: pool}
}

func (r *pgExperimentRepository) Create(ctx context.Context, exp *domain.Experiment) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO experiments (id, owner_user_id, map_id, name, start_x, start_y, goal_x, goal_y, seed, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		exp.ID, exp.OwnerUserID, exp.MapID, exp.Name,
		exp.Start.X, exp.Start.Y, exp.Goal.X, exp.Goal.Y,
		exp.Seed, exp.Status, exp.CreatedAt, exp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create experiment: %w", err)
	}
	return nil
}

func (r *pgExperimentRepository) FindByID(ctx context.Context, id string) (*domain.Experiment, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, owner_user_id, map_id, name, start_x, start_y, goal_x, goal_y, seed, status, created_at, updated_at
		 FROM experiments WHERE id = $1`, id,
	)
	exp := &domain.Experiment{}
	err := row.Scan(
		&exp.ID, &exp.OwnerUserID, &exp.MapID, &exp.Name,
		&exp.Start.X, &exp.Start.Y, &exp.Goal.X, &exp.Goal.Y,
		&exp.Seed, &exp.Status, &exp.CreatedAt, &exp.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find experiment: %w", err)
	}
	return exp, nil
}

func (r *pgExperimentRepository) FindByOwner(ctx context.Context, ownerID string, limit, offset int) ([]domain.Experiment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_user_id, map_id, name, start_x, start_y, goal_x, goal_y, seed, status, created_at, updated_at
		 FROM experiments WHERE owner_user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		ownerID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("find experiments by owner: %w", err)
	}
	defer rows.Close()

	var exps []domain.Experiment
	for rows.Next() {
		var exp domain.Experiment
		if err := rows.Scan(
			&exp.ID, &exp.OwnerUserID, &exp.MapID, &exp.Name,
			&exp.Start.X, &exp.Start.Y, &exp.Goal.X, &exp.Goal.Y,
			&exp.Seed, &exp.Status, &exp.CreatedAt, &exp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan experiment: %w", err)
		}
		exps = append(exps, exp)
	}
	return exps, rows.Err()
}

func (r *pgExperimentRepository) UpdateStatus(ctx context.Context, id string, status domain.ExperimentStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE experiments SET status=$1, updated_at=NOW() WHERE id=$2`, status, id,
	)
	return err
}

func (r *pgExperimentRepository) Cancel(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE experiments SET status='CANCELLED', updated_at=NOW()
		 WHERE id=$1 AND status IN ('PENDING','RUNNING')`, id,
	)
	return err
}
