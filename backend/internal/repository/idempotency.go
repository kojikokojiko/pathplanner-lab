package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgIdempotencyRepository struct {
	pool *pgxpool.Pool
}

func NewIdempotencyRepository(pool *pgxpool.Pool) IdempotencyRepository {
	return &pgIdempotencyRepository{pool: pool}
}

func (r *pgIdempotencyRepository) Get(ctx context.Context, userID, key string) ([]byte, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT response FROM idempotency_keys WHERE user_id = $1 AND key = $2`, userID, key,
	)
	var rawJSON json.RawMessage
	err := row.Scan(&rawJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return rawJSON, nil
}

func (r *pgIdempotencyRepository) Set(ctx context.Context, userID, key string, response []byte) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO idempotency_keys (user_id, key, response)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, key) DO NOTHING`,
		userID, key, json.RawMessage(response),
	)
	return err
}
