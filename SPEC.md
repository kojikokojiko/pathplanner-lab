# SPEC.md — PathPlanner Lab (SaaS)

## 0. Purpose / Goal
Path planning アルゴリズムを「学習・比較・実験管理」できる Web SaaS を構築する。
単なるデモではなく、Webエンジニアリング（REST、DB設計、非同期処理、IaC、CI/CD、O11y）を前面に出す。

**Key Value**
- Map を作成/共有し、複数アルゴリズムで一括実験（Experiment）を実行
- 実験結果（Run）をメトリクスで比較
- LLM が結果を解説し、次に試すパラメータ/アルゴリズムを提案（教材化）

---

## 1. Scope
### 1.1 In Scope (MVP)
- Auth: メール + パスワード（シンプル） or OAuth は後回し
- Map CRUD（グリッド編集）
- Experiment 作成（1 Map に対して複数アルゴリズム/パラメータで Run を作成）
- 非同期実行（API → SQS → Worker）
- Run 結果保存（metrics + artifacts）
- Compare（複数 Run の比較テーブルを返す）
- LLM Review（Run/Compare を解説して保存）
- React ダッシュボード（一覧・詳細・比較・リプレイ）

### 1.2 Out of Scope (MVPではやらない)
- ROS/Gazebo 連携
- 3D 可視化
- D* Lite / RRT / JPS など高度アルゴリズム（追加チケット）
- マルチテナント組織、課金（将来対応）

---

## 2. User Stories (MVP)
1. ユーザーは Map を作成し、障害物/Start/Goal を配置できる
2. ユーザーは Map を選び、アルゴリズム（BFS/Dijkstra/A*/Weighted A*/Greedy）を複数選択し Experiment を作れる
3. システムは Run を非同期で実行し、進捗/完了ステータスを更新する
4. ユーザーは Run の結果（パス、探索ヒートマップ、メトリクス）を閲覧できる
5. ユーザーは複数 Run を比較できる（表 + LLM解説）
6. ユーザーは「LLM Review」を実行して、結果解説を保存/再参照できる

---

## 3. Domain Model (Concept)
- Map: グリッド世界（障害物密度、サイズ、グリッド表現）
- Algorithm: 実装されたアルゴリズム（key + version）
- ParameterSet: アルゴリズムのパラメータ定義（JSONB + hash）
- Experiment: ユーザーが作成する実験（Map + start/goal + runs）
- Run: Experiment の1実行単位（algorithm + parameter_set + result）
- RunMetric: key-value メトリクス（縦持ち）
- RunArtifact: リプレイ/ヒートマップ/パスなどの成果物（JSONB）
- LLMReport: Run/Compare に紐づく解説結果

---

## 4. Algorithms (MVP)
### 4.1 Supported algorithms
- BFS (4-neighbors) — baseline
- Dijkstra (4-neighbors) — optimal with uniform cost
- A* (Manhattan heuristic)
- Weighted A* (w>=1.0)
- Greedy Best-First (f = h only)

### 4.2 Parameters (JSONB examples)
- common:
  - neighbors: 4 | 8 (MVP default 4)
- A*:
  - heuristic: "manhattan"
  - tie_break: "low_g" | "high_g"
- Weighted A*:
  - heuristic: "manhattan"
  - w: 1.0〜3.0
- Greedy:
  - heuristic: "manhattan"

### 4.3 Metrics (minimum set)
- found (bool)
- path_length (int) ※セル数 or コスト
- expanded_nodes (int)
- time_ms (float)
- turns (int) ※任意（パスの方向転換回数）
- explored_ratio (float) ※expanded / (width*height)
- obstacle_ratio (float) ※map側からコピー

### 4.4 Artifacts (minimum set)
- PATH: 座標列 [{x,y},...]
- EXPANSION_HEATMAP: (x,y)->order or counts（圧縮可能）
- TIMELINE: stepごとの状態（MVPは PATH だけでも可）

---

## 5. API Design (REST, v1)
### 5.1 Auth
- POST /v1/auth/register
- POST /v1/auth/login
- POST /v1/auth/logout
- GET  /v1/me

Auth方式: Cookie session or JWT（MVPは JWT 推奨、APIの単純さ重視）

### 5.2 Maps
- POST /v1/maps
- GET  /v1/maps
- GET  /v1/maps/{mapId}
- PUT  /v1/maps/{mapId}
- DELETE /v1/maps/{mapId}

Map payload example:
{
  "name": "maze-01",
  "width": 40,
  "height": 25,
  "grid": "S...#\n.#..#\n...G.\n..."
}

### 5.3 Experiments (async)
- POST /v1/experiments
  - returns 202 + experimentId
  - supports Idempotency-Key header
Request:
{
  "mapId": "...",
  "start": {"x":1,"y":1},
  "goal": {"x":38,"y":22},
  "runs": [
    {"algorithmKey":"ASTAR", "params":{"neighbors":4,"heuristic":"manhattan"}},
    {"algorithmKey":"WEIGHTED_ASTAR", "params":{"neighbors":4,"heuristic":"manhattan","w":1.4}}
  ],
  "seed": 42
}

- GET /v1/experiments
- GET /v1/experiments/{experimentId}
- GET /v1/experiments/{experimentId}/runs
- POST /v1/experiments/{experimentId}:cancel

### 5.4 Runs
- GET /v1/runs/{runId}
- GET /v1/runs/{runId}/metrics
- GET /v1/runs/{runId}/artifacts

### 5.5 Compare
- POST /v1/compare
Request:
{
  "runIds": ["...", "...", "..."]
}
Response:
{
  "table": [
    {"runId":"...", "algorithm":"ASTAR", "time_ms":12.3, "expanded_nodes":1450, "path_length":82},
    ...
  ],
  "insights": null
}

### 5.6 LLM Review
- POST /v1/runs/{runId}:llm-review
- POST /v1/compare:llm-review
Request:
{
  "mode":"teacher", // teacher | concise
  "promptVersion":"v1"
}
Response:
{
  "reportId":"...",
  "summary":"...",
  "recommendations":[...]
}

---

## 6. Database Design (Postgres)
### 6.1 Tables (MVP)
- users
- maps
- algorithms
- parameter_sets
- experiments
- runs
- run_metrics
- run_artifacts
- llm_reports
- idempotency_keys (optional but recommended)

### 6.2 Key constraints / indexes
- parameter_sets.params_hash UNIQUE (dedupe)
- runs(experiment_id) index
- run_metrics(run_id) index
- maps(owner_user_id) index
- experiments(owner_user_id, created_at) index
- idempotency_keys(user_id, key) UNIQUE

### 6.3 grid storage
MVP:
- maps.grid as TEXT (newline separated)
Later:
- RLE compress or bytea

---

## 7. Backend (Go) Architecture
### 7.1 Style
- Clean-ish layering:
  - handler (HTTP)
  - service (usecases)
  - repository (db)
  - domain (entities)
  - worker (job consumers)
- Validation: request DTO validation
- Error: problem+json style error response

### 7.2 Directory example
/backend
  /cmd
    /api
    /worker
  /internal
    /handler
    /service
    /repository
    /domain
    /jobs
    /algorithms
    /llm
    /observability
    /config

### 7.3 Worker jobs (SQS)
Message schema:
{
  "jobType":"RUN_SIMULATION",
  "runId":"uuid",
  "experimentId":"uuid",
  "mapId":"uuid"
}

Worker flow:
- fetch run + map + params
- execute algorithm
- write metrics/artifacts
- mark run status SUCCEEDED/FAILED

---

## 8. Frontend (React + Vite + Tailwind)
### 8.1 Requirements
- Container/Presentational 徹底
- Data fetching: TanStack Query
- Routing: React Router
- Forms: react-hook-form (optional)
- Auth: token management (httpOnly cookie preferred)

### 8.2 Pages (MVP)
- /login
- /maps (list)
- /maps/:id (editor)
- /experiments (list)
- /experiments/:id (detail + runs)
- /runs/:id (run detail + artifacts)
- /compare (run selection + comparison table + LLM report)

### 8.3 Map editor behavior
- grid click: toggle obstacle
- place start/goal mode
- validate: start/goal必須、障害物上不可

---

## 9. LLM Integration
### 9.1 Where LLM is used (MVP)
- Run review:
  - explain why the path looks like that
  - highlight efficiency vs baseline
  - recommend next parameter tweaks
- Compare review:
  - algorithm selection advice based on metrics

### 9.2 Input to LLM (structured JSON)
- map summary: size, obstacle_ratio
- algorithm & params
- metrics
- baseline (optional): dijkstra/astar as reference
- artifacts summary (NOT full grid unless small)

### 9.3 Storage
- llm_reports stores:
  - model, prompt_version
  - summary text
  - recommendations json
  - linked to runId or compare hash

---

## 10. Infrastructure (AWS)
### 10.1 Components
- ECS Fargate:
  - api-service
  - worker-service
- ALB (api)
- RDS Postgres
- SQS queue (run-jobs)
- ECR (images)
- CloudWatch logs (base)
- Secrets Manager or SSM Parameter Store for secrets

Optional later:
- ElastiCache Redis (rate-limit / caching)

### 10.2 Terraform modules (suggested)
/infra
  /modules
    /vpc
    /ecs
    /rds
    /sqs
    /alb
    /ecr
  /envs
    /dev
    /prod

---

## 11. CI/CD (GitHub Actions)
Pipelines:
1) PR:
- frontend: lint/test/build
- backend: go test + lint
2) main merge:
- build docker images (api/worker)
- push to ECR
- terraform plan/apply (dev)
- deploy ECS (force new deployment)
- run DB migration

DB migration:
- goose or migrate tool

---

## 12. Observability (New Relic)
### 12.1 Signals
- API latency p50/p95
- Error rate
- Worker job duration
- SQS queue depth
- DB query timings

### 12.2 Tracing
- OpenTelemetry instrumentation in Go
- Correlation IDs:
  - request_id
  - experiment_id
  - run_id

### 12.3 Structured logs
JSON logs with:
- level, msg
- request_id
- user_id
- run_id
- duration_ms

---

## 13. Non-functional Requirements
- Idempotency: POST /experiments supports Idempotency-Key
- Rate limiting: user per minute (simple)
- Security: validate map size, prevent large payload attacks
- Determinism: seedをRunに保存し再現可能にする（MVPは固定seedでも可）

---

## 14. Milestones
### M0: Skeleton (Day 1-2)
- repo structure, docker compose (frontend, api, db)
- basic health endpoints
- DB migrations baseline

### M1: Maps (Day 3-5)
- Map CRUD + editor UI
- map validation

### M2: Experiments + Worker (Day 6-10)
- POST /experiments (202)
- SQS + worker consumption
- A* + Dijkstra実装（まず2つでOK）
- run metrics/artifacts保存

### M3: Compare + LLM (Day 11-14)
- compare endpoint + UI
- llm review endpoint + report保存

### M4: AWS deploy + IaC + O11y (Day 15+)
- Terraform dev env
- ECS/RDS/SQS
- CI/CD
- New Relic

---

## 15. Definition of Done (MVP)
- 新規ユーザーがMapを作成 → Experiment作成 → Run完了 → Compare → LLM解説
- API, DB, Worker, UI が最低限のエラーハンドリングを持つ
- AWS上で動作し、New Relicで主要メトリクスが見える
- READMEにローカル起動/デプロイ手順がある

---

## 16. Claude Code Prompting Tips
- まず backend の DB DDL と repository 層を生成
- 次に handler/service を生成
- その後 worker とアルゴリズム実装
- 最後に frontend の画面とAPI接続

生成順序を守る（迷子になりにくい）