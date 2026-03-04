package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pathplanner-lab/backend/internal/domain"
	"github.com/pathplanner-lab/backend/internal/repository"
)

type RunInput struct {
	AlgorithmKey string          `json:"algorithmKey"`
	Params       json.RawMessage `json:"params"`
}

type CreateExperimentInput struct {
	MapID string
	Name  string
	Start domain.Point
	Goal  domain.Point
	Runs  []RunInput
	Seed  int64
}

type ExperimentService struct {
	expRepo  repository.ExperimentRepository
	runRepo  repository.RunRepository
	mapRepo  repository.MapRepository
	sqsSvc   SQSPublisher
}

type SQSPublisher interface {
	Publish(ctx context.Context, msg interface{}) error
}

type SQSMessage struct {
	JobType      string `json:"jobType"`
	RunID        string `json:"runId"`
	ExperimentID string `json:"experimentId"`
	MapID        string `json:"mapId"`
}

func NewExperimentService(
	expRepo repository.ExperimentRepository,
	runRepo repository.RunRepository,
	mapRepo repository.MapRepository,
	sqsSvc SQSPublisher,
) *ExperimentService {
	return &ExperimentService{expRepo: expRepo, runRepo: runRepo, mapRepo: mapRepo, sqsSvc: sqsSvc}
}

func (s *ExperimentService) Create(ctx context.Context, ownerID string, input CreateExperimentInput) (*domain.Experiment, error) {
	m, err := s.mapRepo.FindByID(ctx, input.MapID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("map not found")
	}
	if m.OwnerUserID != ownerID {
		return nil, ErrForbidden
	}
	if len(input.Runs) == 0 {
		return nil, fmt.Errorf("at least one run is required")
	}

	now := time.Now()
	exp := &domain.Experiment{
		ID:          uuid.New().String(),
		OwnerUserID: ownerID,
		MapID:       input.MapID,
		Name:        input.Name,
		Start:       input.Start,
		Goal:        input.Goal,
		Seed:        input.Seed,
		Status:      domain.ExperimentStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.expRepo.Create(ctx, exp); err != nil {
		return nil, fmt.Errorf("create experiment: %w", err)
	}

	for _, ri := range input.Runs {
		params := ri.Params
		if params == nil {
			params = json.RawMessage(`{}`)
		}
		run := &domain.Run{
			ID:           uuid.New().String(),
			ExperimentID: exp.ID,
			AlgorithmKey: ri.AlgorithmKey,
			Params:       params,
			Status:       domain.RunStatusPending,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.runRepo.Create(ctx, run); err != nil {
			return nil, fmt.Errorf("create run: %w", err)
		}
		// Enqueue to SQS
		if s.sqsSvc != nil {
			_ = s.sqsSvc.Publish(ctx, SQSMessage{
				JobType:      "RUN_SIMULATION",
				RunID:        run.ID,
				ExperimentID: exp.ID,
				MapID:        input.MapID,
			})
		}
	}

	if err := s.expRepo.UpdateStatus(ctx, exp.ID, domain.ExperimentStatusRunning); err != nil {
		return nil, err
	}
	exp.Status = domain.ExperimentStatusRunning
	return exp, nil
}

func (s *ExperimentService) Get(ctx context.Context, id string, userID string) (*domain.Experiment, error) {
	exp, err := s.expRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if exp == nil {
		return nil, ErrNotFound
	}
	if exp.OwnerUserID != userID {
		return nil, ErrForbidden
	}
	return exp, nil
}

func (s *ExperimentService) List(ctx context.Context, ownerID string, limit, offset int) ([]domain.Experiment, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.expRepo.FindByOwner(ctx, ownerID, limit, offset)
}

func (s *ExperimentService) GetRuns(ctx context.Context, experimentID string, userID string) ([]domain.Run, error) {
	exp, err := s.expRepo.FindByID(ctx, experimentID)
	if err != nil {
		return nil, err
	}
	if exp == nil {
		return nil, ErrNotFound
	}
	if exp.OwnerUserID != userID {
		return nil, ErrForbidden
	}
	return s.runRepo.FindByExperiment(ctx, experimentID)
}

func (s *ExperimentService) Cancel(ctx context.Context, id string, userID string) error {
	exp, err := s.expRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if exp == nil {
		return ErrNotFound
	}
	if exp.OwnerUserID != userID {
		return ErrForbidden
	}
	return s.expRepo.Cancel(ctx, id)
}
