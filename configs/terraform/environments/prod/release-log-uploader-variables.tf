variable "release_log_uploader_service_account_name" {
  type        = string
  description = "The service account name of the release-log-uploader service account."
  default     = "release-log-uploader"
}

variable "release_log_uploader_workflow_name" {
  type        = string
  description = "The name of the GitHub Actions workflow (as defined in the 'name:' field) that uploads release logs."
  default     = "Release report"
}

variable "release_log_uploader_logs_bucket_name" {
  type        = string
  description = "Name of the GCS bucket where release logs are stored."
  default     = "kyma_release_test_logs"
}
