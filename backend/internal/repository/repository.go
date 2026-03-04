package repository

import (
	"context"

	"github.com/pathplanner-lab/backend/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	// Upsert inserts a Cognito-authenticated user by sub (id) and email.
	Upsert(ctx context.Context, id, email string) error
}

type MapRepository interface {
	Create(ctx context.Context, m *domain.Map) error
	FindByID(ctx context.Context, id string) (*domain.Map, error)
	FindByOwner(ctx context.Context, ownerID string, limit, offset int) ([]domain.Map, error)
	Update(ctx context.Context, m *domain.Map) error
	Delete(ctx context.Context, id string) error
}

type ExperimentRepository interface {
	Create(ctx context.Context, exp *domain.Experiment) error
	FindByID(ctx context.Context, id string) (*domain.Experiment, error)
	FindByOwner(ctx context.Context, ownerID string, limit, offset int) ([]domain.Experiment, error)
	UpdateStatus(ctx context.Context, id string, status domain.ExperimentStatus) error
	Cancel(ctx context.Context, id string) error
}

type RunRepository interface {
	Create(ctx context.Context, run *domain.Run) error
	FindByID(ctx context.Context, id string) (*domain.Run, error)
	FindByExperiment(ctx context.Context, experimentID string) ([]domain.Run, error)
	UpdateStatus(ctx context.Context, id string, status domain.RunStatus, errMsg *string) error
	SaveMetrics(ctx context.Context, runID string, metrics []domain.RunMetric) error
	SaveArtifacts(ctx context.Context, runID string, artifacts []domain.RunArtifact) error
	GetMetrics(ctx context.Context, runID string) ([]domain.RunMetric, error)
	GetArtifacts(ctx context.Context, runID string) ([]domain.RunArtifact, error)
}

type LLMReportRepository interface {
	Create(ctx context.Context, r *domain.LLMReport) error
	FindByRunID(ctx context.Context, runID string) ([]domain.LLMReport, error)
	FindByCompareHash(ctx context.Context, hash string) ([]domain.LLMReport, error)
}

type IdempotencyRepository interface {
	Get(ctx context.Context, userID, key string) ([]byte, error)
	Set(ctx context.Context, userID, key string, response []byte) error
}
