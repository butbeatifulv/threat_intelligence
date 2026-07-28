locals {
  compose_file_args = concat(
    [
      "-f", "${var.repo_root}/deploy/discovery/compose.yml",
      "-f", "${var.repo_root}/deploy/pipeline/compose.yml",
      "-f", "${var.repo_root}/deploy/knowledge/compose.yml",
    ],
    var.enable_engage ? [
      "-f", "${var.repo_root}/deploy/engage/compose.yml",
      "-f", "${var.repo_root}/deploy/engage/compose.veil-stack.yml",
    ] : []
  )

  profile_path = "${var.repo_root}/deploy/profiles/${var.profile}.env"
  stack_script = "${var.repo_root}/scripts/ops/terraform-veil-stack.sh"
}

resource "local_file" "compose_env" {
  filename        = "${path.module}/generated/veil-compose.env"
  file_permission = "0644"
  content = templatefile("${path.module}/templates/compose.env.tpl", {
    profile               = var.profile
    project_name          = var.project_name
    pipeline_worker_scale = var.pipeline_worker_scale
    ingest_worker_scale   = var.ingest_worker_scale
    graph_pack_skip       = var.graph_pack_skip
    profile_path          = local.profile_path
  })

  lifecycle {
    precondition {
      condition     = fileexists(local.profile_path)
      error_message = "Deploy profile not found: ${local.profile_path}"
    }
  }
}

resource "null_resource" "compose_stack" {
  count = var.manage_compose ? 1 : 0

  triggers = {
    env_md5      = local_file.compose_env.content_md5
    env_path     = abspath(local_file.compose_env.filename)
    project_name = var.project_name
    engage       = var.enable_engage ? "1" : "0"
    engage_ev    = var.enable_engage_events ? "1" : "0"
    build        = var.compose_build ? "1" : "0"
    repo_root    = var.repo_root
  }

  provisioner "local-exec" {
    command     = "chmod +x '${local.stack_script}' && TERRAFORM_COMPOSE_ENV='${abspath(local_file.compose_env.filename)}' TF_ENABLE_ENGAGE='${var.enable_engage ? "1" : "0"}' TF_ENABLE_ENGAGE_EVENTS='${var.enable_engage_events ? "1" : "0"}' TF_COMPOSE_BUILD='${var.compose_build ? "1" : "0"}' '${local.stack_script}' up"
    working_dir = var.repo_root
  }

  provisioner "local-exec" {
    when    = destroy
    command = "chmod +x '${self.triggers.repo_root}/scripts/ops/terraform-veil-stack.sh' && TERRAFORM_COMPOSE_ENV='${self.triggers.env_path}' TF_ENABLE_ENGAGE='${self.triggers.engage}' TF_ENABLE_ENGAGE_EVENTS='${self.triggers.engage_ev}' '${self.triggers.repo_root}/scripts/ops/terraform-veil-stack.sh' down"
    on_failure = continue
  }
}
