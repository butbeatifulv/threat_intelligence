variable "project_name" {
  type    = string
  default = "veil-local"
}

variable "profile" {
  type    = string
  default = "smoke-minimal"
}

variable "enable_engage" {
  type    = bool
  default = true
}

variable "enable_engage_events" {
  type    = bool
  default = true
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
  type    = string
  default = "1"
}

variable "manage_compose" {
  type        = bool
  description = "When true, terraform apply runs docker compose up; destroy runs down -v."
  default     = false
}

variable "compose_build" {
  type    = bool
  default = true
}
