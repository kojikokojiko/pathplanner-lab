Apply Terraform infrastructure changes.

Ask the user which environment: dev or prod

Then run:
- `make infra-dev` for dev
- `make infra-prod` for prod

After apply completes, run `make env-dev` or `make env-prod` to sync outputs to .env

Remind the user to restart the dev environment with `make dev` if .env was updated.
