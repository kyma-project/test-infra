variable "project_id" {
  type        = string
  description = "The GCP project id which contains workload identity federation"
}

variable "pool_id" {
  type        = string
  description = "Name of the workload identity federation pool"
}

variable "provider_id" {
  type        = string
  description = "Name of the workload identity provider"
}

variable "issuer_uri" {
  type        = string
  description = "Token issuer url for worklaod identity federation"
  default     = "https://token.actions.githubusercontent.com"
}

variable "attribute_mapping" {
  type        = map(any)
  description = "Workload Identity Pool  attributes mapping"
}

variable "sa_mapping" {
  type = map(object({
    sa_name   = string,
    attribute = string
  }))
  description = "Mapping of service accounts and corresponding workload identity federation attributes"
}
