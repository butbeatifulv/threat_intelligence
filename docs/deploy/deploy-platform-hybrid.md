# Veil hybrid deploy platform (P5)

Mature production model: **Terraform** provisions cloud resources, **Ansible** configures hosts and runs the stateful Compose data plane, **Helm** runs the Kubernetes control plane (HTTP APIs and autoscaled workers).

## Prerequisites

- Pentest execution runs in **[veneno](https://github.com/butbeautifulv/veneno)**; veil provides graph read + `engage.events` ingest bridge ([veil-veneno-split.md](../architecture/veil-veneno-split.md), [mcp-agents.md](../agents/mcp-agents.md)).
- Application config: [deploy/profiles/](../deploy/profiles/) + [versions.env](../versions.env).

## Layer responsibilities

| Layer | Directory | Owns |
|-------|-----------|------|
| Terraform | [deploy/terraform/](../deploy/terraform/) | VPC, EKS/VM, S3, IAM, secrets outputs, optional managed DB |
| Ansible | [deploy/ansible/](../deploy/ansible/) | OS hardening, Docker, Compose deploy, scrape cron, TLS on VM |
| Helm | [deploy/helm/veil/](../deploy/helm/veil/) | `api`, `engage-api`, `pipeline_worker`, `ingest_worker`, `scrape` CronJob |
| Compose (dev/CI) | [deploy/discovery|pipeline|graph|engage/](../deploy/) | Source of service definitions; Helm values generated from same env keys |

## Workload placement

| Service | Stage/prod placement |
|---------|----------------------|
| Neo4j, NATS, crawl-db | Ansible + Compose on data nodes (or managed equivalents) |
| `api`, nginx secure | Helm Deployment + Ingress |
| `engage-api`, `engage-events-worker` | Helm (veil-stack values: NATS/Neo4j endpoints from TF outputs) |
| `pipeline_worker`, `ingest_worker` | Helm Deployment + HPA (NATS consumer lag) |
| `scrape_worker` | Helm **CronJob** per profile **or** Ansible `cron` + `compose run` |
| `engage-runner` | Dedicated node pool / VM — **not** co-scheduled with graph API |

## Deploy flow (stage)

```bash
# 1. Infrastructure
cd deploy/terraform/environments/stage && terraform init && terraform apply

# 2. Configure data plane hosts
cd deploy/ansible && ansible-playbook -i inventories/stage playbooks/site.yml

# 3. Control plane on EKS
helm upgrade --install veil deploy/helm/veil \
  -f deploy/helm/veil/values.yaml \
  -f deploy/helm/veil/values-stage.yaml \
  --set global.imageTag=$(grep APP_VERSION ../../versions.env | cut -d= -f2)
```

## Local / CI

Unchanged: `docker compose`, `make test-platform-*`, [deploy/terraform/README.md](../deploy/terraform/README.md) for local env generation.

## Plan

Implementation phases: [.cursor/plans/veil_deploy_platform_p5_hybrid.plan.md](../.cursor/plans/veil_deploy_platform_p5_hybrid.plan.md).
