-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE maps (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    width           INT NOT NULL CHECK (width > 0 AND width <= 200),
    height          INT NOT NULL CHECK (height > 0 AND height <= 200),
    grid            TEXT NOT NULL,
    obstacle_ratio  DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_maps_owner_user_id ON maps(owner_user_id);

CREATE TABLE algorithms (
    key         TEXT PRIMARY KEY,
    version     TEXT NOT NULL DEFAULT 'v1',
    description TEXT
);

INSERT INTO algorithms (key, version, description) VALUES
    ('BFS',           'v1', 'Breadth-First Search'),
    ('DIJKSTRA',      'v1', 'Dijkstra shortest path'),
    ('ASTAR',         'v1', 'A* with Manhattan heuristic'),
    ('WEIGHTED_ASTAR','v1', 'Weighted A* (f = g + w*h)'),
    ('GREEDY',        'v1', 'Greedy Best-First Search');

CREATE TABLE parameter_sets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    params      JSONB NOT NULL,
    params_hash TEXT NOT NULL UNIQUE
);

CREATE TABLE experiments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    map_id          UUID NOT NULL REFERENCES maps(id) ON DELETE CASCADE,
    name            TEXT NOT NULL DEFAULT '',
    start_x         INT NOT NULL,
    start_y         INT NOT NULL,
    goal_x          INT NOT NULL,
    goal_y          INT NOT NULL,
    seed            BIGINT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING','RUNNING','COMPLETED','FAILED','CANCELLED')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_experiments_owner_user_id_created ON experiments(owner_user_id, created_at DESC);

CREATE TABLE runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id   UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    algorithm_key   TEXT NOT NULL REFERENCES algorithms(key),
    params          JSONB NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING','RUNNING','SUCCEEDED','FAILED')),
    error_msg       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_runs_experiment_id ON runs(experiment_id);

CREATE TABLE run_metrics (
    run_id  UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    key     TEXT NOT NULL,
    value   DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (run_id, key)
);

CREATE INDEX idx_run_metrics_run_id ON run_metrics(run_id);

CREATE TABLE run_artifacts (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id  UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    kind    TEXT NOT NULL,
    data    JSONB NOT NULL
);

CREATE INDEX idx_run_artifacts_run_id ON run_artifacts(run_id);

CREATE TABLE llm_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id          UUID REFERENCES runs(id) ON DELETE SET NULL,
    compare_hash    TEXT,
    model           TEXT NOT NULL,
    prompt_version  TEXT NOT NULL DEFAULT 'v1',
    summary         TEXT NOT NULL,
    recommendations JSONB NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_llm_reports_run_id ON llm_reports(run_id);

CREATE TABLE idempotency_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key         TEXT NOT NULL,
    response    JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, key)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS llm_reports;
DROP TABLE IF EXISTS run_artifacts;
DROP TABLE IF EXISTS run_metrics;
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS experiments;
DROP TABLE IF EXISTS parameter_sets;
DROP TABLE IF EXISTS algorithms;
DROP TABLE IF EXISTS maps;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
