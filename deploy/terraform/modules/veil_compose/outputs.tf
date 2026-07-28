output "compose_env_path" {
  description = "Path to generated compose env file (source before docker compose)."
  value       = local_file.compose_env.filename
}

output "compose_file_args" {
  description = "Docker compose -f arguments for this stack shape."
  value       = local.compose_file_args
}

output "profile_path" {
  value = local.profile_path
}

output "project_name" {
  value = var.project_name
}
