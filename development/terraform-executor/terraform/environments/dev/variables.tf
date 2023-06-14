variable "terraform_executor_k8s_service_account" {
  type = object({
    name      = string
    namespace = string
  })

  default = {
    name      = "terraform-executor"
    namespace = "default"
  }

  description = "Details of terraform executor k8s service account."
}

variable "terraform_executor_gcp_service_account" {
  type = object({
    id         = string
    project_id = string
  })

  default = {
    id         = "terraform-executor"
    project_id = "sap-kyma-neighbors-dev"
  }

  description = "Details of terraform executor gcp service account."
}

variable "managed_k8s_cluster" {
  type = object({
    name     = string
    location = string
  })

  description = "Details of the managed k8s cluster."
}

variable "gcp_region" {
  type        = string
  default     = "europe-west4"
  description = "Default Google Cloud region to create resources."
}

variable "gcp_project_id" {
  type        = string
  default     = "sap-kyma-neighbors-dev"
  description = "Google Cloud project to create resources."
}

variable "external_secrets_sa" {
  type = object({
    name      = string
    namespace = string
  })

  default = {
    name      = "external-secrets"
    namespace = "external-secrets"
  }

  description = "Details of external secrets service account."
}
