# Veil Terraform (Compose IaC)

Declarative wrapper around existing **Docker Compose** deploy files. Terraform does not replace Compose; it generates env artifacts and optionally drives `docker compose` via `scripts/ops/terraform-veil-stack.sh`.

## Layout

| Path | Purpose |
|------|---------|
| `modules/veil_compose/` | Reusable module: profile + scales + optional engage overlay |
| `environments/local/` | Local/dev stack (default profile `smoke-minimal`) |
| `environments/stage/` | Stage foundation stub + compose env (`fast-rich`) |
| `environments/prod/` | Prod foundation stub + compose env (`secure-graph`) |
| `modules/foundation/` | VPC/EKS/VM/S3 placeholder outputs for Ansible/Helm |

## Quick start (generate env only)

```bash
cd deploy/terraform/environments/local
terraform init
terraform apply
```

Default `manage_compose = false` — apply writes `modules/veil_compose/generated/veil-compose.env` only. Run the stack manually:

```bash
export TERRAFORM_COMPOSE_ENV="$(terraform output -raw compose_env_path)"
source "$TERRAFORM_COMPOSE_ENV"
source "$VEIL_PROFILE_PATH"
./scripts/ops/terraform-veil-stack.sh up   # from repository root
```

Or use existing ops scripts: `./scripts/ops/compose-up-full.sh` with the same profile (`source deploy/profiles/smoke-minimal.env`).

## Managed compose (terraform drives up/down)

```bash
terraform apply -var='manage_compose=true'
terraform destroy   # runs compose down -v when manage_compose was true
```

Requires Docker on the host running Terraform.

## Variables (local)

See [environments/local/variables.tf](environments/local/variables.tf) and [terraform.tfvars.example](environments/local/terraform.tfvars.example).

| Variable | Default | Notes |
|----------|---------|-------|
| `profile` | `smoke-minimal` | Must exist under `deploy/profiles/<name>.env` |
| `enable_engage` | `true` | Adds engage + veil-stack compose files |
| `manage_compose` | `false` | Set `true` to run stack on apply |

## Relation to platform smokes

- P3 closed loop: `make test-platform-closed-loop` (engage only, no scrape)
- P4b full loop: `make test-platform-full-loop` (scrape → graph → engage) — [docs/architecture/platform-full-loop-smoke.md](../../docs/architecture/platform-full-loop-smoke.md)
