variable "environment" {
  type        = string
  description = "Environment name (stage, prod)."
}

variable "region" {
  type    = string
  default = "eu-central-1"
}

variable "cluster_enabled" {
  type        = bool
  description = "When true, provision managed Kubernetes (EKS/GKE module stub)."
  default     = false
}

variable "vm_count" {
  type        = number
  description = "Data-plane VM count for Ansible targets when not using managed services."
  default     = 2
}

variable "blob_bucket_name" {
  type        = string
  description = "S3-compatible bucket for scrape blobs / graph pack artifacts."
  default     = ""
}
