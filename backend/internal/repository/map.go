package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pathplanner-lab/backend/internal/domain"
)

type pgMapRepository struct {
	pool *pgxpool.Pool
}

func NewMapRepository(pool *pgxpool.Pool) MapRepository {
	return &pgMapRepository{pool: pool}
}

func (r *pgMapRepository) Create(ctx context.Context, m *domain.Map) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO maps (id, owner_user_id, name, width, height, grid, obstacle_ratio, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		m.ID, m.OwnerUserID, m.Name, m.Width, m.Height, m.Grid, m.ObstacleRatio, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create map: %w", err)
	}
	return nil
}

func (r *pgMapRepository) FindByID(ctx context.Context, id string) (*domain.Map, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, owner_user_id, name, width, height, grid, obstacle_ratio, created_at, updated_at
		 FROM maps WHERE id = $1`, id,
	)
	m := &domain.Map{}
	err := row.Scan(&m.ID, &m.OwnerUserID, &m.Name, &m.Width, &m.Height, &m.Grid, &m.ObstacleRatio, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find map by id: %w", err)
	}
	return m, nil
}

func (r *pgMapRepository) FindByOwner(ctx context.Context, ownerID string, limit, offset int) ([]domain.Map, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_user_id, name, width, height, grid, obstacle_ratio, created_at, updated_at
		 FROM maps WHERE owner_user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		ownerID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("find maps by owner: %w", err)
	}
	defer rows.Close()

	var maps []domain.Map
	for rows.Next() {
		var m domain.Map
		if err := rows.Scan(&m.ID, &m.OwnerUserID, &m.Name, &m.Width, &m.Height, &m.Grid, &m.ObstacleRatio, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan map: %w", err)
		}
		maps = append(maps, m)
	}
	return maps, rows.Err()
}

func (r *pgMapRepository) Update(ctx context.Context, m *domain.Map) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE maps SET name=$1, width=$2, height=$3, grid=$4, obstacle_ratio=$5, updated_at=$6
		 WHERE id=$7 AND owner_user_id=$8`,
		m.Name, m.Width, m.Height, m.Grid, m.ObstacleRatio, m.UpdatedAt, m.ID, m.OwnerUserID,
	)
	if err != nil {
		return fmt.Errorf("update map: %w", err)
	}
	return nil
}

func (r *pgMapRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM maps WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete map: %w", err)
	}
	return nil
}
