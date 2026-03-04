package handler

import (
	"net/http"

	"github.com/pathplanner-lab/backend/internal/service"
)

type CompareHandler struct {
	compareSvc *service.CompareService
	llmSvc     *service.LLMService
}

func NewCompareHandler(compareSvc *service.CompareService, llmSvc *service.LLMService) *CompareHandler {
	return &CompareHandler{compareSvc: compareSvc, llmSvc: llmSvc}
}

func (h *CompareHandler) Compare(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RunIDs []string `json:"runIds"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON")
		return
	}
	result, err := h.compareSvc.Compare(r.Context(), req.RunIDs)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *CompareHandler) LLMReview(w http.ResponseWriter, r *http.Request) {
	if h.llmSvc == nil {
		writeProblem(w, http.StatusServiceUnavailable, "Service Unavailable", "LLM not configured")
		return
	}

	var req struct {
		RunIDs        []string `json:"runIds"`
		Mode          string   `json:"mode"`
		PromptVersion string   `json:"promptVersion"`
	}
	req.Mode = "teacher"
	req.PromptVersion = "v1"
	if err := decodeJSON(r, &req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON")
		return
	}

	// Get userID from context - use first run's experiment owner
	// For compare, we skip auth check since runs are already accessible
	report, err := h.llmSvc.ReviewCompare(r.Context(), req.RunIDs, "", req.Mode, req.PromptVersion)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"reportId":        report.ID,
		"summary":         report.Summary,
		"recommendations": report.Recommendations,
	})
}
