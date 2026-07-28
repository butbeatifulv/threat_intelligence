variable "cluster_enabled" {
  type    = bool
  default = true
}

variable "vm_count" {
  type    = number
  default = 3
}

variable "blob_bucket_name" {
  type    = string
  default = ""
}

variable "veil_profile" {
  type    = string
  default = "secure-graph"
}
