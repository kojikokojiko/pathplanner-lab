package service

import (
	"context"
	"fmt"

	"github.com/pathplanner-lab/backend/internal/domain"
	"github.com/pathplanner-lab/backend/internal/repository"
)

type RunService struct {
	runRepo repository.RunRepository
	expRepo repository.ExperimentRepository
}

func NewRunService(runRepo repository.RunRepository, expRepo repository.ExperimentRepository) *RunService {
	return &RunService{runRepo: runRepo, expRepo: expRepo}
}

func (s *RunService) Get(ctx context.Context, id string, userID string) (*domain.Run, error) {
	run, err := s.runRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, ErrNotFound
	}
	exp, err := s.expRepo.FindByID(ctx, run.ExperimentID)
	if err != nil {
		return nil, err
	}
	if exp == nil || exp.OwnerUserID != userID {
		return nil, ErrForbidden
	}
	return run, nil
}

func (s *RunService) GetMetrics(ctx context.Context, runID string, userID string) ([]domain.RunMetric, error) {
	if _, err := s.Get(ctx, runID, userID); err != nil {
		return nil, err
	}
	metrics, err := s.runRepo.GetMetrics(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("get metrics: %w", err)
	}
	return metrics, nil
}

func (s *RunService) GetArtifacts(ctx context.Context, runID string, userID string) ([]domain.RunArtifact, error) {
	if _, err := s.Get(ctx, runID, userID); err != nil {
		return nil, err
	}
	artifacts, err := s.runRepo.GetArtifacts(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("get artifacts: %w", err)
	}
	return artifacts, nil
}
