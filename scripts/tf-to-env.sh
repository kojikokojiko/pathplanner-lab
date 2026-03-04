#!/bin/sh
# terraform output を読み取って .env に書き込む
# Usage: ./scripts/tf-to-env.sh dev|prod
set -e

ENV=${1:-dev}
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INFRA_DIR="$SCRIPT_DIR/../infra"
ENV_FILE="$SCRIPT_DIR/../.env"

cd "$INFRA_DIR"

echo "==> Switching to workspace: $ENV"
terraform workspace select "$ENV" 2>/dev/null || terraform workspace new "$ENV"

POOL_ID=$(terraform output -raw user_pool_id)
CLIENT_ID=$(terraform output -raw client_id)
DOMAIN=$(terraform output -raw cognito_domain)
REGION=$(terraform output -raw aws_region 2>/dev/null || echo "us-east-1")

echo "==> Writing Cognito config to .env"

# 既存の値を置換 (sed で in-place)
update_env() {
  KEY=$1
  VALUE=$2
  if grep -q "^${KEY}=" "$ENV_FILE"; then
    sed -i.bak "s|^${KEY}=.*|${KEY}=${VALUE}|" "$ENV_FILE" && rm -f "${ENV_FILE}.bak"
  else
    echo "${KEY}=${VALUE}" >> "$ENV_FILE"
  fi
}

update_env "COGNITO_USER_POOL_ID" "$POOL_ID"
update_env "COGNITO_CLIENT_ID" "$CLIENT_ID"
update_env "COGNITO_REGION" "$REGION"
update_env "VITE_COGNITO_DOMAIN" "$DOMAIN"
update_env "VITE_COGNITO_CLIENT_ID" "$CLIENT_ID"

echo ""
echo "==> Done! .env updated:"
echo "    COGNITO_USER_POOL_ID = $POOL_ID"
echo "    COGNITO_CLIENT_ID    = $CLIENT_ID"
echo "    VITE_COGNITO_DOMAIN  = $DOMAIN"
echo ""
echo "==> Run 'make dev' to apply the new settings."
