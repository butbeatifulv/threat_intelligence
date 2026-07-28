# Foundation module — cloud-agnostic stub.
# Implement provider-specific resources (aws_vpc, aws_eks, google_container_cluster, etc.)
# in provider wrappers; keep outputs stable for Ansible inventory and Helm values.

locals {
  name_prefix = "veil-${var.environment}"
}

# Placeholder outputs until cloud provider modules land (P5b).
output "environment" {
  value = var.environment
}

output "name_prefix" {
  value = local.name_prefix
}

output "cluster_enabled" {
  value = var.cluster_enabled
}

output "ansible_inventory_hint" {
  value = "deploy/ansible/inventories/${var.environment}/hosts.yml"
}

output "helm_values_hint" {
  value = "deploy/helm/veil/values-${var.environment}.yaml"
}
