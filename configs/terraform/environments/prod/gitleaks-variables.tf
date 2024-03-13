variable "gitleaks_gcp_service_account" {
  type = object({
    id         = string
    project_id = string
  })

  default = {
    id         = "gitleaks-secret-accesor"
    project_id = "sap-kyma-prow"
  }

  description = "Details of gitleaks secret accesor gcp service account."
}

variable "gitleaks_repositories" {
  type    = set(string)
  default = ["test-infra"]

  description = "List of repositories that can use gitleaks secrets accesor service account"
}

variable "gitleaks_workflow_name" {
  type        = string
  default     = "gitleaks"
  description = "Name of the gitleaks workflow"
}
