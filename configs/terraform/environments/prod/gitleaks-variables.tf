# Details about service account in Google Cloud Platform that should have access to gitleaks secrets.
# Such access is should be granted manually to that service account
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

# Name of the workflow that runs gitleaks scans mapped to gitleaks service account
# via workload identity federation
variable "gitleaks_workflow_name" {
  type        = string
  default     = "gitleaks"
  description = "Name of the gitleaks workflow"
}
