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
# GitHub Enterprise (github.tools.sap) Configuration
# ------------------------------------------------------------------------------
# These variables configure the connection to the internal GitHub Enterprise
# instance (github.tools.sap) used by SAP for internal repositories.
# The provider is configured in provider.tf and uses these variables.
# ------------------------------------------------------------------------------

variable "github_tools_sap_organization_name" {
  type        = string
  default     = "kyma"
  description = "The Kyma GitHub organization in internal GitHub Enterprise instance"
}

variable "github_tools_sap_token" {
  type        = string
  sensitive   = true
  description = "GitHub token for github.tools.sap provider. Passed via TF_VAR_github_tools_sap_token environment variable from GitHub Actions workflow."
}