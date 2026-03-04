# PathPlanner Lab

Path planning アルゴリズムを学習・比較・実験管理できる Web SaaS。

## アーキテクチャ

```
frontend/    React + Vite + Tailwind
backend/     Go (Clean Architecture: handler/service/repository/domain)
migrations/  Goose SQL migrations
infra/       Terraform (coming)
```

## ローカル起動

### 前提条件

- Docker & Docker Compose
- Go 1.23+
- Node.js 20+

### 1. 環境変数設定

```bash
cp frontend/.env.example frontend/.env
# ANTHROPIC_API_KEY を設定（LLM機能を使う場合）
export ANTHROPIC_API_KEY=sk-ant-...
```

### 2. 依存関係インストール

```bash
# Go依存関係
cd backend && go mod tidy && cd ..

# Node依存関係
cd frontend && npm install && cd ..
```

### 3. Docker Compose で起動

```bash
# DB + LocalStack (SQS) を起動してマイグレーション実行
make up

# または手動で
docker compose up -d db localstack
sleep 5
make migrate
docker compose up -d api worker
```

### 4. フロントエンド開発サーバー起動

```bash
make frontend-dev
# → http://localhost:5173 でアクセス
```

### サービス一覧

| Service | Port | 説明 |
|---------|------|------|
| frontend | 5173 | React Dev Server |
| api | 8080 | Go REST API |
| db | 5432 | PostgreSQL |
| localstack | 4566 | SQS (LocalStack) |

## API エンドポイント

```
POST /v1/auth/register    ユーザー登録
POST /v1/auth/login       ログイン
GET  /v1/me               自分の情報

POST   /v1/maps           マップ作成
GET    /v1/maps           マップ一覧
GET    /v1/maps/:id       マップ取得
PUT    /v1/maps/:id       マップ更新
DELETE /v1/maps/:id       マップ削除

POST /v1/experiments            実験作成 (202 Accepted)
GET  /v1/experiments            実験一覧
GET  /v1/experiments/:id        実験詳細
GET  /v1/experiments/:id/runs   実験のRun一覧
POST /v1/experiments/:id/cancel キャンセル

GET /v1/runs/:id          Run詳細
GET /v1/runs/:id/metrics  メトリクス
GET /v1/runs/:id/artifacts アーティファクト
POST /v1/runs/:id/llm-review LLMレビュー

POST /v1/compare          複数Run比較
POST /v1/compare/llm-review LLM比較レビュー
```

## アルゴリズム

| Key | 説明 |
|-----|------|
| BFS | Breadth-First Search |
| DIJKSTRA | Dijkstra最短経路 |
| ASTAR | A* (Manhattan heuristic) |
| WEIGHTED_ASTAR | Weighted A* (w >= 1.0) |
| GREEDY | Greedy Best-First |

## ミルストーン

- [x] M0: スケルトン (DB, docker-compose, API基盤)
- [x] M1: Map CRUD + グリッドエディタ
- [x] M2: Experiments + Worker + アルゴリズム5種
- [x] M3: Compare + LLM Review
- [ ] M4: AWS Terraform + CI/CD + New Relic

## DB マイグレーション (手動)

```bash
# goose がインストールされていない場合
go install github.com/pressly/goose/v3/cmd/goose@latest

# マイグレーション実行
goose -dir migrations postgres "postgres://pathplanner:secret@localhost:5432/pathplanner?sslmode=disable" up
```
