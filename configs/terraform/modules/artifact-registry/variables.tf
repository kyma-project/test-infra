variable "repository_name" {
  type        = string
  description = "Artifact Registry repository name"
}

variable "owner" {
  type        = string
  description = "Owner inside SAP"
  default     = "neighbors"
}

variable "repoAdmin_serviceaccounts" {
  type = list(string)
  description = "Service Accounts with reapoAdmin access"
  default = []
}

variable "writer_serviceaccounts" {
  type        = list(string)
  description = "Service Accounts with write access"
  default = []
}

variable "reader_serviceaccounts" {
  type        = list(string)
  description = "Service Accounts with read access"
  default = []
}

variable "type" {
  type        = string
  description = "Environment for the resources"
  default     = "development"
}

variable "multi_region" {
  type        = bool
  description = "Is Artifact Registry location type Multi-region"
  default     = true
}

variable "primary_area" {
  type        = string
  description = "Location of primary area of the Artifact Registry for multi-region repositories"
  default     = "europe"

  validation {
    condition = var.primary_area != ""
    error_message = "When multi_region is true, primary_area must be set."
  }
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

variable "location" {
  type        = string
  description = "Location of the Artifact Registry for non multi-region repositories"
  default     = "europe"
}

variable "description" {
  type        = string
  description = "Description of the Artifact Registry"
  default     = "Artifact Registry for kyma-project"
}

variable "format" {
  type        = string
  description = "Format of the Artifact Registry"
  default     = "DOCKER"
}
