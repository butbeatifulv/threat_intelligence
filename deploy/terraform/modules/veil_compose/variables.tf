variable "repo_root" {
  type        = string
  description = "Absolute path to Veil repository root (docker build context)."
}

variable "project_name" {
  type        = string
  description = "Docker Compose project name."
  default     = "veil-tf"
}

variable "profile" {
  type        = string
  description = "Deploy profile name under deploy/profiles/<name>.env"
  default     = "smoke-minimal"
}

variable "enable_engage" {
  type        = bool
  description = "Include engage compose files and veil-stack overlay."
  default     = true
}

variable "enable_engage_events" {
  type        = bool
  description = "Start engage-events-worker (requires enable_engage)."
  default     = true
}

variable "pipeline_worker_scale" {
  type    = number
  default = 1
}

variable "ingest_worker_scale" {
  type    = number
  default = 1
}

variable "graph_pack_skip" {
  type        = string
  description = "1 skips graph-bootstrap pack import."
  default     = "1"
}

variable "manage_compose" {
  type        = bool
  description = "When true, terraform apply/up runs docker compose up; destroy runs down -v."
  default     = false
}

variable "compose_build" {
  type        = bool
  description = "Pass --build to compose up."
  default     = true
}
