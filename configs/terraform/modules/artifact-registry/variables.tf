###################################
# Artifact Registry related values
###################################
variable "registry_name" {
  type        = string
  description = "Artifact Registry name"
}

variable "owner" {
  type        = string
  description = "Owner inside SAP"
  default     = "neighbors"
}

variable "writer_serviceaccounts" {
  type        = list(string)
  description = "Service Accounts with reapoAdmin access"
}

variable "reader_serviceaccounts" {
  type        = list(string)
  description = "Service Accounts with read access (lifecycle-maneger)"
}

variable "type" {
  type        = string
  description = "Environment for the resources"
  default     = "development"
}

variable "multi_region" {
  type        = bool
  description = "Is Location type Multi-region"
  default     = true
}

variable "primary_area" {
  type        = string
  description = "Location type Multi-region"
  default     = "europe"
}

variable "immutable_tags" {
  type        = bool
  description = "Is Artifact registry immutable"
  default     = false
}

variable "public" {
  type        = bool
  description = "Is Artifact registry public"
  default     = false
}
# (2025-03-04)