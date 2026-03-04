.PHONY: dev up down build migrate logs \
	infra-init infra-dev infra-prod infra-destroy-dev env-dev env-prod

# ─── 開発 (ホットリロード付き) ──────────────────────
dev:
	docker compose -f docker-compose.dev.yml up

dev-build:
	docker compose -f docker-compose.dev.yml up --build

dev-down:
	docker compose -f docker-compose.dev.yml down

# ─── 本番ビルド (Docker image) ──────────────────────
up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

# ─── DB マイグレーション ──────────────────────────
migrate:
	docker compose -f docker-compose.dev.yml run --rm migrate

migrate-status:
	docker run --rm --network host \
		-v $(PWD)/migrations:/migrations \
		ghcr.io/pressly/goose:v3 \
		goose -dir /migrations postgres \
		"postgres://pathplanner:secret@localhost:5432/pathplanner?sslmode=disable" status

# ─── ログ ─────────────────────────────────────────
logs:
	docker compose -f docker-compose.dev.yml logs -f api worker

logs-api:
	docker compose -f docker-compose.dev.yml logs -f api

logs-worker:
	docker compose -f docker-compose.dev.yml logs -f worker

# ─── Terraform (Cognito) ───────────────────────────
infra-init:
	cd infra && terraform init

# dev 環境の Cognito を作成・更新
infra-dev:
	cd infra && terraform workspace select dev 2>/dev/null || terraform workspace new dev
	cd infra && terraform apply -auto-approve -var-file=envs/dev.tfvars

# prod 環境の Cognito を作成・更新
infra-prod:
	cd infra && terraform workspace select prod 2>/dev/null || terraform workspace new prod
	cd infra && terraform apply -auto-approve -var-file=envs/prod.tfvars

# dev 環境を削除
infra-destroy-dev:
	cd infra && terraform workspace select dev && terraform destroy -var-file=envs/dev.tfvars

# Terraform outputs を .env に書き込む
env-dev:
	./scripts/tf-to-env.sh dev

env-prod:
	./scripts/tf-to-env.sh prod
