variable "project_id" {
  type        = string
  description = "The project id to create Workload Identity Pool"
}

variable "pool_id" {
  type        = string
  description = "Workload Identity Pool id"
}

variable "provider_id" {
  type        = string
  description = "Workload Identity Provider id"
}

variable "issuer_uri" {
  type        = string
  description = "Workload Identity Issuer URL"
}

variable "attribute_mapping" {
  type        = map(string)
  description = "Workload Identity attribute mapping"
}

variable "sa_mapping" {
  type = map(object({
    sa_name   = string
    attribute = string
  }))
}


