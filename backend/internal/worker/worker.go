package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/pathplanner-lab/backend/internal/algorithms"
	"github.com/pathplanner-lab/backend/internal/domain"
	"github.com/pathplanner-lab/backend/internal/repository"
)

type Job struct {
	JobType      string `json:"jobType"`
	RunID        string `json:"runId"`
	ExperimentID string `json:"experimentId"`
	MapID        string `json:"mapId"`
}

type MessageSource interface {
	Receive(ctx context.Context) ([]RawMessage, error)
	Delete(ctx context.Context, receiptHandle string) error
}

type RawMessage struct {
	Body          string
	ReceiptHandle string
}

type Worker struct {
	source  MessageSource
	runRepo repository.RunRepository
	expRepo repository.ExperimentRepository
	mapRepo repository.MapRepository
	logger  *zap.Logger
}

func New(
	source MessageSource,
	runRepo repository.RunRepository,
	expRepo repository.ExperimentRepository,
	mapRepo repository.MapRepository,
	logger *zap.Logger,
) *Worker {
	return &Worker{
		source:  source,
		runRepo: runRepo,
		expRepo: expRepo,
		mapRepo: mapRepo,
		logger:  logger,
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.logger.Info("worker started")
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker stopping")
			return
		default:
		}

		msgs, err := w.source.Receive(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			w.logger.Error("receive messages", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		for _, msg := range msgs {
			w.processMessage(ctx, msg)
		}
	}
}

func (w *Worker) processMessage(ctx context.Context, msg RawMessage) {
	var job Job
	if err := json.Unmarshal([]byte(msg.Body), &job); err != nil {
		w.logger.Error("unmarshal job", zap.Error(err), zap.String("body", msg.Body))
		_ = w.source.Delete(ctx, msg.ReceiptHandle)
		return
	}

	w.logger.Info("processing job", zap.String("runId", job.RunID), zap.String("jobType", job.JobType))

	if err := w.executeRun(ctx, job); err != nil {
		w.logger.Error("execute run failed", zap.Error(err), zap.String("runId", job.RunID))
		errMsg := err.Error()
		_ = w.runRepo.UpdateStatus(ctx, job.RunID, domain.RunStatusFailed, &errMsg)
	}

	_ = w.source.Delete(ctx, msg.ReceiptHandle)
	w.updateExperimentStatus(ctx, job.ExperimentID)
}

func (w *Worker) executeRun(ctx context.Context, job Job) error {
	// Mark run as running
	if err := w.runRepo.UpdateStatus(ctx, job.RunID, domain.RunStatusRunning, nil); err != nil {
		return fmt.Errorf("mark running: %w", err)
	}

	run, err := w.runRepo.FindByID(ctx, job.RunID)
	if err != nil || run == nil {
		return fmt.Errorf("find run: %w", err)
	}

	exp, err := w.expRepo.FindByID(ctx, job.ExperimentID)
	if err != nil || exp == nil {
		return fmt.Errorf("find experiment: %w", err)
	}

	m, err := w.mapRepo.FindByID(ctx, job.MapID)
	if err != nil || m == nil {
		return fmt.Errorf("find map: %w", err)
	}

	// Parse algorithm params
	params := algorithms.ParseParams(run.Params)
	grid := algorithms.ParseGrid(m.Grid, m.Width, m.Height)

	algo, err := algorithms.Get(run.AlgorithmKey)
	if err != nil {
		return fmt.Errorf("get algorithm: %w", err)
	}

	input := algorithms.AlgorithmInput{
		Grid:   grid,
		Start:  algorithms.Point{X: exp.Start.X, Y: exp.Start.Y},
		Goal:   algorithms.Point{X: exp.Goal.X, Y: exp.Goal.Y},
		Params: params,
	}

	startTime := time.Now()
	output := algo.Run(input)
	_ = startTime

	// Build metrics
	totalCells := m.Width * m.Height
	exploredRatio := 0.0
	if totalCells > 0 {
		exploredRatio = float64(output.ExpandedNodes) / float64(totalCells)
	}

	pathLength := 0
	if output.Found {
		pathLength = len(output.Path) - 1
	}

	foundVal := 0.0
	if output.Found {
		foundVal = 1.0
	}

	metrics := []domain.RunMetric{
		{RunID: run.ID, Key: "found", Value: foundVal},
		{RunID: run.ID, Key: "path_length", Value: float64(pathLength)},
		{RunID: run.ID, Key: "expanded_nodes", Value: float64(output.ExpandedNodes)},
		{RunID: run.ID, Key: "time_ms", Value: float64(output.TimeMicros) / 1000.0},
		{RunID: run.ID, Key: "turns", Value: float64(output.Turns)},
		{RunID: run.ID, Key: "explored_ratio", Value: exploredRatio},
		{RunID: run.ID, Key: "obstacle_ratio", Value: m.ObstacleRatio},
	}

	if err := w.runRepo.SaveMetrics(ctx, run.ID, metrics); err != nil {
		return fmt.Errorf("save metrics: %w", err)
	}

	// Build artifacts
	var artifacts []domain.RunArtifact

	if output.Found {
		pathData, _ := json.Marshal(output.Path)
		artifacts = append(artifacts, domain.RunArtifact{
			RunID: run.ID,
			Kind:  "PATH",
			Data:  pathData,
		})
	}

	if len(output.ExploredOrder) > 0 {
		heatmapData, _ := json.Marshal(output.ExploredOrder)
		artifacts = append(artifacts, domain.RunArtifact{
			RunID: run.ID,
			Kind:  "EXPANSION_HEATMAP",
			Data:  heatmapData,
		})
	}

	if len(artifacts) > 0 {
		if err := w.runRepo.SaveArtifacts(ctx, run.ID, artifacts); err != nil {
			return fmt.Errorf("save artifacts: %w", err)
		}
	}

	if err := w.runRepo.UpdateStatus(ctx, run.ID, domain.RunStatusSucceeded, nil); err != nil {
		return fmt.Errorf("mark succeeded: %w", err)
	}

	w.logger.Info("run completed",
		zap.String("runId", run.ID),
		zap.String("algorithm", run.AlgorithmKey),
		zap.Bool("found", output.Found),
		zap.Int("expanded", output.ExpandedNodes),
	)
	return nil
}

func (w *Worker) updateExperimentStatus(ctx context.Context, experimentID string) {
	runs, err := w.runRepo.FindByExperiment(ctx, experimentID)
	if err != nil {
		return
	}
	allDone := true
	anyFailed := false
	for _, r := range runs {
		if r.Status == domain.RunStatusPending || r.Status == domain.RunStatusRunning {
			allDone = false
			break
		}
		if r.Status == domain.RunStatusFailed {
			anyFailed = true
		}
	}
	if allDone {
		status := domain.ExperimentStatusCompleted
		if anyFailed {
			status = domain.ExperimentStatusFailed
		}
		_ = w.expRepo.UpdateStatus(ctx, experimentID, status)
	}
}
