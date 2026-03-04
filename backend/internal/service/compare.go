package service

import (
	"context"
	"fmt"

	"github.com/pathplanner-lab/backend/internal/domain"
	"github.com/pathplanner-lab/backend/internal/repository"
)

type CompareRow struct {
	RunID         string  `json:"run_id"`
	Algorithm     string  `json:"algorithm"`
	Status        string  `json:"status"`
	Found         bool    `json:"found"`
	PathLength    float64 `json:"path_length"`
	ExpandedNodes float64 `json:"expanded_nodes"`
	TimeMs        float64 `json:"time_ms"`
	Turns         float64 `json:"turns"`
	ExploredRatio float64 `json:"explored_ratio"`
}

type CompareResult struct {
	Table    []CompareRow `json:"table"`
	Insights *string      `json:"insights"`
}

type CompareService struct {
	runRepo repository.RunRepository
}

func NewCompareService(runRepo repository.RunRepository) *CompareService {
	return &CompareService{runRepo: runRepo}
}

func (s *CompareService) Compare(ctx context.Context, runIDs []string) (*CompareResult, error) {
	if len(runIDs) == 0 {
		return nil, fmt.Errorf("at least one run ID required")
	}
	if len(runIDs) > 20 {
		return nil, fmt.Errorf("too many run IDs (max 20)")
	}

	var rows []CompareRow
	for _, runID := range runIDs {
		run, err := s.runRepo.FindByID(ctx, runID)
		if err != nil {
			return nil, err
		}
		if run == nil {
			continue
		}
		row := CompareRow{
			RunID:     run.ID,
			Algorithm: run.AlgorithmKey,
			Status:    string(run.Status),
		}
		if run.Status == domain.RunStatusSucceeded {
			metrics, err := s.runRepo.GetMetrics(ctx, runID)
			if err == nil {
				for _, m := range metrics {
					switch m.Key {
					case "found":
						row.Found = m.Value > 0
					case "path_length":
						row.PathLength = m.Value
					case "expanded_nodes":
						row.ExpandedNodes = m.Value
					case "time_ms":
						row.TimeMs = m.Value
					case "turns":
						row.Turns = m.Value
					case "explored_ratio":
						row.ExploredRatio = m.Value
					}
				}
			}
		}
		rows = append(rows, row)
	}

	return &CompareResult{Table: rows}, nil
}
