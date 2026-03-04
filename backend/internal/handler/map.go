package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pathplanner-lab/backend/internal/domain"
	"github.com/pathplanner-lab/backend/internal/middleware"
	"github.com/pathplanner-lab/backend/internal/service"
)

type MapHandler struct {
	mapSvc *service.MapService
}

func NewMapHandler(mapSvc *service.MapService) *MapHandler {
	return &MapHandler{mapSvc: mapSvc}
}

func (h *MapHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	var req struct {
		Name   string `json:"name"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Grid   string `json:"grid"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON")
		return
	}
	m, err := h.mapSvc.Create(r.Context(), userID, service.CreateMapInput{
		Name:   req.Name,
		Width:  req.Width,
		Height: req.Height,
		Grid:   req.Grid,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *MapHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)
	maps, err := h.mapSvc.List(r.Context(), userID, limit, offset)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if maps == nil {
		maps = []domain.Map{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": maps, "limit": limit, "offset": offset})
}

func (h *MapHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	mapID := chi.URLParam(r, "mapId")
	m, err := h.mapSvc.Get(r.Context(), mapID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *MapHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	mapID := chi.URLParam(r, "mapId")
	var req struct {
		Name   string `json:"name"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Grid   string `json:"grid"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "invalid JSON")
		return
	}
	m, err := h.mapSvc.Update(r.Context(), mapID, userID, service.CreateMapInput{
		Name:   req.Name,
		Width:  req.Width,
		Height: req.Height,
		Grid:   req.Grid,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *MapHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	mapID := chi.URLParam(r, "mapId")
	if err := h.mapSvc.Delete(r.Context(), mapID, userID); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
