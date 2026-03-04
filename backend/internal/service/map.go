package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/pathplanner-lab/backend/internal/domain"
	"github.com/pathplanner-lab/backend/internal/repository"
)

type MapService struct {
	mapRepo repository.MapRepository
}

func NewMapService(mapRepo repository.MapRepository) *MapService {
	return &MapService{mapRepo: mapRepo}
}

type CreateMapInput struct {
	Name   string
	Width  int
	Height int
	Grid   string
}

func (s *MapService) Create(ctx context.Context, ownerID string, input CreateMapInput) (*domain.Map, error) {
	if err := validateMapInput(input); err != nil {
		return nil, err
	}
	grid := normalizeGrid(input.Grid, input.Width, input.Height)
	ratio := calcObstacleRatio(grid, input.Width, input.Height)

	m := &domain.Map{
		ID:            uuid.New().String(),
		OwnerUserID:   ownerID,
		Name:          input.Name,
		Width:         input.Width,
		Height:        input.Height,
		Grid:          grid,
		ObstacleRatio: ratio,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := s.mapRepo.Create(ctx, m); err != nil {
		return nil, fmt.Errorf("create map: %w", err)
	}
	return m, nil
}

func (s *MapService) Get(ctx context.Context, id string, userID string) (*domain.Map, error) {
	m, err := s.mapRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrNotFound
	}
	if m.OwnerUserID != userID {
		return nil, ErrForbidden
	}
	return m, nil
}

func (s *MapService) List(ctx context.Context, ownerID string, limit, offset int) ([]domain.Map, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.mapRepo.FindByOwner(ctx, ownerID, limit, offset)
}

func (s *MapService) Update(ctx context.Context, id string, ownerID string, input CreateMapInput) (*domain.Map, error) {
	m, err := s.mapRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrNotFound
	}
	if m.OwnerUserID != ownerID {
		return nil, ErrForbidden
	}
	if err := validateMapInput(input); err != nil {
		return nil, err
	}
	m.Name = input.Name
	m.Width = input.Width
	m.Height = input.Height
	m.Grid = normalizeGrid(input.Grid, input.Width, input.Height)
	m.ObstacleRatio = calcObstacleRatio(m.Grid, m.Width, m.Height)
	m.UpdatedAt = time.Now()
	if err := s.mapRepo.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MapService) Delete(ctx context.Context, id string, ownerID string) error {
	m, err := s.mapRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return ErrNotFound
	}
	if m.OwnerUserID != ownerID {
		return ErrForbidden
	}
	return s.mapRepo.Delete(ctx, id)
}

func validateMapInput(input CreateMapInput) error {
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}
	if input.Width < 2 || input.Width > 200 {
		return fmt.Errorf("width must be between 2 and 200")
	}
	if input.Height < 2 || input.Height > 200 {
		return fmt.Errorf("height must be between 2 and 200")
	}
	return nil
}

func normalizeGrid(grid string, width, height int) string {
	lines := strings.Split(grid, "\n")
	result := make([]string, height)
	for y := 0; y < height; y++ {
		if y < len(lines) {
			line := lines[y]
			count := utf8.RuneCountInString(line)
			if count > width {
				runes := []rune(line)
				result[y] = string(runes[:width])
			} else if count < width {
				result[y] = line + strings.Repeat(".", width-count)
			} else {
				result[y] = line
			}
		} else {
			result[y] = strings.Repeat(".", width)
		}
	}
	return strings.Join(result, "\n")
}

func calcObstacleRatio(grid string, width, height int) float64 {
	total := width * height
	if total == 0 {
		return 0
	}
	obstacles := strings.Count(grid, "#")
	return float64(obstacles) / float64(total)
}
