variable "release_log_uploader_service_account_name" {
  type        = string
  description = "The service account name of the release-log-uploader service account."
  default     = "release-log-uploader"
}

variable "release_log_uploader_compliancy_workflow_ref_public_github" {
  type        = string
  description = "The value of GitHub OIDC token job_workflow_ref claim of the release log upload workflow in kyma-project/compliancy repository on github.com."
  default     = "compliancy/.github/workflows/release-log-upload.yaml@refs/heads/main"
}

variable "release_log_uploader_compliancy_workflow_ref_internal_github" {
  type        = string
  description = "The value of GitHub OIDC token job_workflow_ref claim of the release log upload workflow in kyma/compliancy repository on github.tools.sap."
  default     = "compliancy/.github/workflows/release-log-upload.yaml@refs/heads/main"
}

variable "release_log_uploader_logs_bucket_name" {
  type        = string
  description = "Name of the GCS bucket where release logs are stored."
  default     = "kyma_release_test_logs"
}
