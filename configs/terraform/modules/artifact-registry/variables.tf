###################################
# Artifact Registry related values
###################################
variable "artifact_registry_name" {
  type        = string
  description = "Artifact Registry name"
}

variable "artifact_registry_owner" {
  type        = string
  description = "Owner inside SAP"
  default     = "neighbors"
}

variable "artifact_registry_writer_serviceaccount" {
  type        = string
  description = "Service Account"
}

variable "artifact_registry_reader_serviceaccounts" {
  type        = list(string)
  description = "Service Accounts with read access (lifecycle-maneger)"
}

variable "artifact_registry_type" {
  type        = string
  description = "Environment for the resources"
  default     = "development"
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

variable "immutable_artifact_registry" {
  type        = bool
  description = "Is Artifact registry immutable"
  default     = false
}

variable "artifact_registry_public" {
  type        = bool
  description = "Is Artifact registry public"
  default     = false
}
