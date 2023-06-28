###################################
# Artifact Registry related values
###################################
variable "module" {
  type        = string
  description = "Module name"
}

variable "type" {
  type        = string
  description = "Environment for the resources"
}

variable "artifact_registry_multi_region" {
  type        = bool
  description = "Is Location type Multi-region"
  default     = true
}

variable "artifact_registry_primary_area" {
  type        = string
  description = "Location type Multi-region"
  default     = "europe"
}


variable "artifact_registry_prefix" {
  type        = string
  description = "Naming prefix for all Artifact registry"
  default     = "modules"
}

variable "artifact_registry_count" {
  type        = number
  description = "Number of Artifact registries to create"
  default     = 2
}

variable "artifact_registry_names" {
  type        = list(string)
  description = "Artifact Registry names"
  default     = ["ocim", "internal"]
}

variable "immutable_artifact_registry" {
  type        = bool
  description = "Is Artifact registry immutable"
  default     = false
}