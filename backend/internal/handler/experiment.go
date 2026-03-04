package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pathplanner-lab/backend/internal/domain"
	"github.com/pathplanner-lab/backend/internal/middleware"
	"github.com/pathplanner-lab/backend/internal/repository"
	"github.com/pathplanner-lab/backend/internal/service"
)

type ExperimentHandler struct {
	expSvc       *service.ExperimentService
	idempotency  repository.IdempotencyRepository
}

func NewExperimentHandler(expSvc *service.ExperimentService, idempotency repository.IdempotencyRepository) *ExperimentHandler {
	return &ExperimentHandler{expSvc: expSvc, idempotency: idempotency}
}

func (h *ExperimentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	idempKey := r.Header.Get("Idempotency-Key")

	// Idempotency check
	if idempKey != "" && h.idempotency != nil {
		existing, err := h.idempotency.Get(r.Context(), userID, idempKey)
		if err == nil && existing != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(existing)
			return
		}
	}

	var req struct {
		MapID string             `json:"mapId"`
		Name  string             `json:"name"`
		Start domain.Point       `json:"start"`
		Goal  domain.Point       `json:"goal"`
		Runs  []service.RunInput `json:"runs"`
		Seed  int64              `json:"seed"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON")
		return
	}

	exp, err := h.expSvc.Create(r.Context(), userID, service.CreateExperimentInput{
		MapID: req.MapID,
		Name:  req.Name,
		Start: req.Start,
		Goal:  req.Goal,
		Runs:  req.Runs,
		Seed:  req.Seed,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respBytes, _ := json.Marshal(exp)
	if idempKey != "" && h.idempotency != nil {
		_ = h.idempotency.Set(r.Context(), userID, idempKey, respBytes)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write(respBytes)
}

func (h *ExperimentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)
	exps, err := h.expSvc.List(r.Context(), userID, limit, offset)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if exps == nil {
		exps = []domain.Experiment{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": exps, "limit": limit, "offset": offset})
}

func (h *ExperimentHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	expID := chi.URLParam(r, "experimentId")
	exp, err := h.expSvc.Get(r.Context(), expID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, exp)
}

func (h *ExperimentHandler) GetRuns(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	expID := chi.URLParam(r, "experimentId")
	runs, err := h.expSvc.GetRuns(r.Context(), expID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if runs == nil {
		runs = []domain.Run{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": runs})
}

func (h *ExperimentHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	expID := chi.URLParam(r, "experimentId")
	if err := h.expSvc.Cancel(r.Context(), expID, userID); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}
