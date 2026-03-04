package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pathplanner-lab/backend/internal/domain"
	"github.com/pathplanner-lab/backend/internal/llm"
	"github.com/pathplanner-lab/backend/internal/repository"
)

type LLMService struct {
	llmClient *llm.Client
	runRepo   repository.RunRepository
	expRepo   repository.ExperimentRepository
	mapRepo   repository.MapRepository
	llmRepo   repository.LLMReportRepository
}

func NewLLMService(
	llmClient *llm.Client,
	runRepo repository.RunRepository,
	expRepo repository.ExperimentRepository,
	mapRepo repository.MapRepository,
	llmRepo repository.LLMReportRepository,
) *LLMService {
	return &LLMService{
		llmClient: llmClient,
		runRepo:   runRepo,
		expRepo:   expRepo,
		mapRepo:   mapRepo,
		llmRepo:   llmRepo,
	}
}

func (s *LLMService) ReviewRun(ctx context.Context, runID string, userID string, mode string, promptVersion string) (*domain.LLMReport, error) {
	run, err := s.runRepo.FindByID(ctx, runID)
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
	m, err := s.mapRepo.FindByID(ctx, exp.MapID)
	if err != nil || m == nil {
		return nil, fmt.Errorf("map not found")
	}
	metrics, err := s.runRepo.GetMetrics(ctx, runID)
	if err != nil {
		return nil, err
	}

	mapSummary := llm.MapSummary{
		Width: m.Width, Height: m.Height, ObstacleRatio: m.ObstacleRatio,
	}
	result, err := s.llmClient.ReviewRun(ctx, run, metrics, mapSummary, mode)
	if err != nil {
		return nil, fmt.Errorf("llm review: %w", err)
	}

	recsJSON, _ := json.Marshal(result.Recommendations)

	report := &domain.LLMReport{
		ID:              uuid.New().String(),
		RunID:           &runID,
		Model:           "claude-opus-4-6",
		PromptVersion:   promptVersion,
		Summary:         result.Summary,
		Recommendations: recsJSON,
		CreatedAt:       time.Now(),
	}
	if err := s.llmRepo.Create(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func (s *LLMService) ReviewCompare(ctx context.Context, runIDs []string, userID string, mode string, promptVersion string) (*domain.LLMReport, error) {
	if len(runIDs) == 0 {
		return nil, fmt.Errorf("runIds required")
	}
	compareHash := computeCompareHash(runIDs)

	var rows []map[string]interface{}
	for _, runID := range runIDs {
		run, err := s.runRepo.FindByID(ctx, runID)
		if err != nil || run == nil {
			continue
		}
		metrics, _ := s.runRepo.GetMetrics(ctx, runID)
		mMap := make(map[string]float64)
		for _, m := range metrics {
			mMap[m.Key] = m.Value
		}
		rows = append(rows, map[string]interface{}{
			"run_id":    runID,
			"algorithm": run.AlgorithmKey,
			"metrics":   mMap,
		})
	}

	result, err := s.llmClient.ReviewCompare(ctx, rows, llm.MapSummary{}, mode)
	if err != nil {
		return nil, err
	}

	recsJSON, _ := json.Marshal(result.Recommendations)

	report := &domain.LLMReport{
		ID:              uuid.New().String(),
		CompareHash:     &compareHash,
		Model:           "claude-opus-4-6",
		PromptVersion:   promptVersion,
		Summary:         result.Summary,
		Recommendations: recsJSON,
		CreatedAt:       time.Now(),
	}
	if err := s.llmRepo.Create(ctx, report); err != nil {
		return nil, err
	}
	return report, nil
}

func computeCompareHash(runIDs []string) string {
	sorted := make([]string, len(runIDs))
	copy(sorted, runIDs)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}
