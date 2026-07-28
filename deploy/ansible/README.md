# Veil Ansible (data plane)

Configures VM hosts for **stateful Compose** (Neo4j, NATS, crawl-db) and **scheduled scrape** batches. Control-plane HTTP services may run on Kubernetes via [Helm](../helm/veil/).

```bash
cd deploy/ansible
ansible-playbook -i inventories/stage playbooks/site.yml --check
```

Inventory hosts are filled from Terraform outputs (P5b). See [docs/deploy/deploy-platform-hybrid.md](../../docs/deploy/deploy-platform-hybrid.md).
