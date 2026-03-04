package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pathplanner-lab/backend/internal/middleware"
	"github.com/pathplanner-lab/backend/internal/service"
)

type RunHandler struct {
	runSvc *service.RunService
}

func NewRunHandler(runSvc *service.RunService) *RunHandler {
	return &RunHandler{runSvc: runSvc}
}

func (h *RunHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	runID := chi.URLParam(r, "runId")
	run, err := h.runSvc.Get(r.Context(), runID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (h *RunHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	runID := chi.URLParam(r, "runId")
	metrics, err := h.runSvc.GetMetrics(r.Context(), runID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": metrics})
}

func (h *RunHandler) GetArtifacts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	runID := chi.URLParam(r, "runId")
	artifacts, err := h.runSvc.GetArtifacts(r.Context(), runID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": artifacts})
}

func (h *RunHandler) LLMReview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	runID := chi.URLParam(r, "runId")

	var req struct {
		Mode          string `json:"mode"`
		PromptVersion string `json:"promptVersion"`
	}
	req.Mode = "teacher"
	req.PromptVersion = "v1"
	_ = decodeJSON(r, &req)

	// This will be wired to LLMService in router
	_ = userID
	_ = runID
	writeProblem(w, http.StatusNotImplemented, "Not Implemented", "use /v1/runs/{runId}/llm-review via router")
}
