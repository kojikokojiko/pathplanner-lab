package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pathplanner-lab/backend/internal/domain"
)

type pgLLMReportRepository struct {
	pool *pgxpool.Pool
}

func NewLLMReportRepository(pool *pgxpool.Pool) LLMReportRepository {
	return &pgLLMReportRepository{pool: pool}
}

func (r *pgLLMReportRepository) Create(ctx context.Context, rep *domain.LLMReport) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO llm_reports (id, run_id, compare_hash, model, prompt_version, summary, recommendations, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		rep.ID, rep.RunID, rep.CompareHash, rep.Model, rep.PromptVersion, rep.Summary, rep.Recommendations, rep.CreatedAt,
	)
	return err
}

func (r *pgLLMReportRepository) FindByRunID(ctx context.Context, runID string) ([]domain.LLMReport, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, run_id, compare_hash, model, prompt_version, summary, recommendations, created_at
		 FROM llm_reports WHERE run_id = $1 ORDER BY created_at DESC`, runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []domain.LLMReport
	for rows.Next() {
		var rep domain.LLMReport
		if err := rows.Scan(&rep.ID, &rep.RunID, &rep.CompareHash, &rep.Model, &rep.PromptVersion, &rep.Summary, &rep.Recommendations, &rep.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan llm_report: %w", err)
		}
		reports = append(reports, rep)
	}
	return reports, rows.Err()
}

func (r *pgLLMReportRepository) FindByCompareHash(ctx context.Context, hash string) ([]domain.LLMReport, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, run_id, compare_hash, model, prompt_version, summary, recommendations, created_at
		 FROM llm_reports WHERE compare_hash = $1 ORDER BY created_at DESC`, hash,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []domain.LLMReport
	for rows.Next() {
		var rep domain.LLMReport
		if err := rows.Scan(&rep.ID, &rep.RunID, &rep.CompareHash, &rep.Model, &rep.PromptVersion, &rep.Summary, &rep.Recommendations, &rep.CreatedAt); err != nil {
			return nil, err
		}
		reports = append(reports, rep)
	}
	return reports, rows.Err()
}
