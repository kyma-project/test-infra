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



variable "kyma-project-github-org" {
  type        = string
  default     = "kyma-project"
  description = "The GitHub organization where the Kyma project is hosted"
}

# TODO (dekiel): should be moved to internal-github.tf
variable "github-tools-sap-organization-name" {
  type        = string
  default     = "kyma"
  description = "The GitHub organization where the tools-sap is hosted"
}

variable "github_tools_sap_token" {
  type        = string
  sensitive   = true
  description = "GitHub token for github.tools.sap provider. Passed via TF_VAR_GITHUB_TOOLS_SAP_TOKEN environment variable from GitHub Actions workflow."
}