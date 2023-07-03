
###################################
# Artifact Registry related values
###################################
variable "artifact_registry_owner" {
  type        = string
  description = "Owner inside SAP"
  default     = "neighbors"
}

variable "artifact_registry_module" {
  type        = string
  description = "Module name"
}

variable "artifact_registry_type" {
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