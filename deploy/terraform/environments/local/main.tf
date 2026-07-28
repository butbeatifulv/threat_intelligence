terraform {
  required_version = ">= 1.5.0"

  required_providers {
    local = {
      source  = "hashicorp/local"
      version = ">= 2.4.0"
    }
    null = {
      source  = "hashicorp/null"
      version = ">= 3.2.0"
    }
  }
}

locals {
  repo_root = abspath("${path.module}/../../../..")
}

module "veil" {
  source = "../../modules/veil_compose"

  repo_root    = local.repo_root
  project_name = var.project_name
  profile      = var.profile

  enable_engage         = var.enable_engage
  enable_engage_events  = var.enable_engage_events
  pipeline_worker_scale = var.pipeline_worker_scale
  ingest_worker_scale   = var.ingest_worker_scale
  graph_pack_skip       = var.graph_pack_skip

  manage_compose = var.manage_compose
  compose_build  = var.compose_build
}

output "compose_env_path" {
  value = module.veil.compose_env_path
}

output "profile_path" {
  value = module.veil.profile_path
}
