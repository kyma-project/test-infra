variable "gcp_region" {
  type        = string
  default     = "europe-west4"
  description = "Default Google Cloud region to create resources."
}

variable "gcp_scheduler_region" {
  type        = string
  default     = "europe-west3"
  description = "Additional Google Cloud Region to create resources."
}

variable "gcp_project_id" {
  type        = string
  default     = "sap-kyma-prow"
  description = "Google Cloud project to create resources."
}

variable "kyma_project_gcp_region" {
  type        = string
  description = "Default Google Cloud region to create resources in kyma-project"
  default     = "europe-west4"
}

variable "kyma_project_gcp_project_id" {
  type        = string
  description = "Google Cloud project to create resources"
  default     = "kyma-project"
}



variable "kyma_project_github_org" {
  type        = string
  default     = "kyma-project"
  description = "The GitHub organization where the Kyma project is hosted"
}

# ------------------------------------------------------------------------------
# Internal GitHub Enterprise Configuration
# ------------------------------------------------------------------------------
# These variables configure the connection to the internal GitHub Enterprise
# instance used by SAP for internal repositories.
# The provider is configured in provider.tf and uses these variables.
# ------------------------------------------------------------------------------

variable "internal_github_organization_name" {
  type        = string
  default     = "kyma"
  description = "The Kyma GitHub organization in internal GitHub Enterprise instance"
}

variable "internal_github_token" {
  type        = string
  sensitive   = true
  description = "GitHub token for internal GitHub provider. Passed via TF_VAR_internal_github_token environment variable from GitHub Actions workflow."
}

variable "internal_github_base_url" {
  type        = string
  default     = "https://github.tools.sap/api/v3"
  description = "Base URL for the internal GitHub Enterprise API endpoint"
}