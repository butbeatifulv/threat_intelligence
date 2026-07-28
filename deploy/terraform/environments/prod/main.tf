terraform {
  required_version = ">= 1.5.0"
}

locals {
  repo_root = abspath("${path.module}/../../../..")
}

module "foundation" {
  source = "../../modules/foundation"

  environment      = "prod"
  cluster_enabled  = var.cluster_enabled
  vm_count         = var.vm_count
  blob_bucket_name = var.blob_bucket_name
}

module "veil_compose" {
  source = "../../modules/veil_compose"

  repo_root      = local.repo_root
  project_name   = "veil-prod"
  profile        = var.veil_profile
  manage_compose = false
}

output "foundation" {
  value = module.foundation
}

output "compose_env_path" {
  value = module.veil_compose.compose_env_path
}
