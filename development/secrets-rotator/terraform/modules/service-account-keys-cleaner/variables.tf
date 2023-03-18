variable "project" {
  type = object({
    id     = string
    number = number
  })
  description = "Google Cloud project ID to deploy the service account keys cleaner to."
}

variable "region" {
  type        = string
  description = "Google Cloud region to deploy the service account keys cleaner to."
}

variable "service_name" {
  type        = string
  description = "Name of the service account keys cleaner service instance."
}

variable "scheduler_name" {
  type        = string
  description = "Name of the service account keys cleaner scheduler instance."
}

variable "application_name" {
  type        = string
  description = "Name of the application a service account keys cleaner is a part of."
}

variable "service_account_keys_cleaner_account_id" {
  type        = string
  default     = "sa-keys-cleaner"
  description = "Service account ID of the service account keys cleaner."
}

variable "service_account_keys_cleaner_image" {
  type        = string
  description = "Image of the service account keys cleaner."
  validation {
    condition     = can(regex("^europe-docker.pkg.dev/kyma.*", var.service_account_keys_cleaner_image))
    error_message = "The service account keys cleaner image must be hosted in the Kyma Google Artifact Registry."
  }
}

variable "cloud_run_service_listen_port" {
  type        = number
  description = "Port on which the service account keys cleaner service listens."
  default     = 8080
}

variable "secrets_rotator_sa_email" {
  type        = string
  description = "Secrets rotator application service account email."
}

variable "scheduler_cron_schedule" {
  type        = string
  description = "Cron schedule for the service account keys cleaner scheduler."
  default     = "0 0 * * *"
}

variable "service_account_key_latest_version_min_age" {
  type        = number
  description = "Minimum age in hours the service account key latest version exist, before old version to be deleted."
}
