package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pathplanner-lab/backend/internal/domain"
)

type pgRunRepository struct {
	pool *pgxpool.Pool
}

func NewRunRepository(pool *pgxpool.Pool) RunRepository {
	return &pgRunRepository{pool: pool}
}

func (r *pgRunRepository) Create(ctx context.Context, run *domain.Run) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO runs (id, experiment_id, algorithm_key, params, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		run.ID, run.ExperimentID, run.AlgorithmKey, run.Params, run.Status, run.CreatedAt, run.UpdatedAt,
	)
	return err
}

func (r *pgRunRepository) FindByID(ctx context.Context, id string) (*domain.Run, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, experiment_id, algorithm_key, params, status, error_msg, created_at, updated_at
		 FROM runs WHERE id = $1`, id,
	)
	run := &domain.Run{}
	err := row.Scan(&run.ID, &run.ExperimentID, &run.AlgorithmKey, &run.Params, &run.Status, &run.ErrorMsg, &run.CreatedAt, &run.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find run: %w", err)
	}
	return run, nil
}

func (r *pgRunRepository) FindByExperiment(ctx context.Context, experimentID string) ([]domain.Run, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, experiment_id, algorithm_key, params, status, error_msg, created_at, updated_at
		 FROM runs WHERE experiment_id = $1 ORDER BY created_at ASC`,
		experimentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []domain.Run
	for rows.Next() {
		var run domain.Run
		if err := rows.Scan(&run.ID, &run.ExperimentID, &run.AlgorithmKey, &run.Params, &run.Status, &run.ErrorMsg, &run.CreatedAt, &run.UpdatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (r *pgRunRepository) UpdateStatus(ctx context.Context, id string, status domain.RunStatus, errMsg *string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE runs SET status=$1, error_msg=$2, updated_at=NOW() WHERE id=$3`,
		status, errMsg, id,
	)
	return err
}

func (r *pgRunRepository) SaveMetrics(ctx context.Context, runID string, metrics []domain.RunMetric) error {
	for _, m := range metrics {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO run_metrics (run_id, key, value) VALUES ($1, $2, $3)
			 ON CONFLICT (run_id, key) DO UPDATE SET value = EXCLUDED.value`,
			runID, m.Key, m.Value,
		)
		if err != nil {
			return fmt.Errorf("save metric %s: %w", m.Key, err)
		}
	}
	return nil
}

func (r *pgRunRepository) SaveArtifacts(ctx context.Context, runID string, artifacts []domain.RunArtifact) error {
	for _, a := range artifacts {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO run_artifacts (run_id, kind, data) VALUES ($1, $2, $3)`,
			runID, a.Kind, a.Data,
		)
		if err != nil {
			return fmt.Errorf("save artifact %s: %w", a.Kind, err)
		}
	}
	return nil
}

func (r *pgRunRepository) GetMetrics(ctx context.Context, runID string) ([]domain.RunMetric, error) {
	rows, err := r.pool.Query(ctx, `SELECT run_id, key, value FROM run_metrics WHERE run_id = $1`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []domain.RunMetric
	for rows.Next() {
		var m domain.RunMetric
		if err := rows.Scan(&m.RunID, &m.Key, &m.Value); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

func (r *pgRunRepository) GetArtifacts(ctx context.Context, runID string) ([]domain.RunArtifact, error) {
	rows, err := r.pool.Query(ctx, `SELECT run_id, kind, data FROM run_artifacts WHERE run_id = $1`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []domain.RunArtifact
	for rows.Next() {
		var a domain.RunArtifact
		if err := rows.Scan(&a.RunID, &a.Kind, &a.Data); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, a)
	}
	return artifacts, rows.Err()
}
