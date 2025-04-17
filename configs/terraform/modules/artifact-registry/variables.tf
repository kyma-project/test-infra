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

  validation {
    condition = contains(["development", "production"], var.type)
    error_message = "Type must be either 'development' or 'production'."
  }
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
  default = ""
}

variable "format" {
  type        = string
  description = "Format of the Artifact Registry"
  default     = "DOCKER"
}

variable "mode" {
  type        = string
  description = "Mode of the Artifact Registry"
  default     = "STANDARD_REPOSITORY"

  validation {
    condition     = contains(["STANDARD_REPOSITORY", "VIRTUAL_REPOSITORY", "REMOTE_REPOSITORY"], var.mode)
    error_message = "Mode must be either 'STANDARD_REPOSITORY', 'VIRTUAL_REPOSITORY', or 'REMOTE_REPOSITORY'."
  }
}

variable "cleanup_policy_dry_run" {
  type        = bool
  description = "Is cleanup policy dry run"
  default     = false
}

variable "cleanup_policies" {
  type = list(object({
    id     = string
    action = string
    condition = optional(object({
      tag_state = optional(string)
      older_than = optional(string)
      tag_prefixes = optional(list(string))
    }))
  }))
  default = []
}

variable "remote_repository_config" {
  type = object({
    description = optional(string)
    docker_repository = optional(object({
      public_repository = string
    }))
    upstream_credentials = optional(object({
      username_password_credentials = optional(object({
        username                = string
        password_secret_version = string
      }))
    }))
  })
  default = null
}
