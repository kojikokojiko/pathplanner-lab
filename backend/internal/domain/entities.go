package domain

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Map struct {
	ID            string    `json:"id"`
	OwnerUserID   string    `json:"owner_user_id"`
	Name          string    `json:"name"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	Grid          string    `json:"grid"`
	ObstacleRatio float64   `json:"obstacle_ratio"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ParameterSet struct {
	ID         string          `json:"id"`
	Params     json.RawMessage `json:"params"`
	ParamsHash string          `json:"params_hash"`
}

type ExperimentStatus string

const (
	ExperimentStatusPending   ExperimentStatus = "PENDING"
	ExperimentStatusRunning   ExperimentStatus = "RUNNING"
	ExperimentStatusCompleted ExperimentStatus = "COMPLETED"
	ExperimentStatusFailed    ExperimentStatus = "FAILED"
	ExperimentStatusCancelled ExperimentStatus = "CANCELLED"
)

type Experiment struct {
	ID          string           `json:"id"`
	OwnerUserID string           `json:"owner_user_id"`
	MapID       string           `json:"map_id"`
	Name        string           `json:"name"`
	Start       Point            `json:"start"`
	Goal        Point            `json:"goal"`
	Seed        int64            `json:"seed"`
	Status      ExperimentStatus `json:"status"`
	Runs        []Run            `json:"runs,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type RunStatus string

const (
	RunStatusPending   RunStatus = "PENDING"
	RunStatusRunning   RunStatus = "RUNNING"
	RunStatusSucceeded RunStatus = "SUCCEEDED"
	RunStatusFailed    RunStatus = "FAILED"
)

type Run struct {
	ID           string          `json:"id"`
	ExperimentID string          `json:"experiment_id"`
	AlgorithmKey string          `json:"algorithm_key"`
	Params       json.RawMessage `json:"params"`
	Status       RunStatus       `json:"status"`
	Metrics      []RunMetric     `json:"metrics,omitempty"`
	Artifacts    []RunArtifact   `json:"artifacts,omitempty"`
	ErrorMsg     *string         `json:"error_msg,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type RunMetric struct {
	RunID string  `json:"run_id"`
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

type RunArtifact struct {
	RunID string          `json:"run_id"`
	Kind  string          `json:"kind"`
	Data  json.RawMessage `json:"data"`
}

type LLMReport struct {
	ID              string          `json:"id"`
	RunID           *string         `json:"run_id,omitempty"`
	CompareHash     *string         `json:"compare_hash,omitempty"`
	Model           string          `json:"model"`
	PromptVersion   string          `json:"prompt_version"`
	Summary         string          `json:"summary"`
	Recommendations json.RawMessage `json:"recommendations"`
	CreatedAt       time.Time       `json:"created_at"`
}
