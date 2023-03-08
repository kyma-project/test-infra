variable "project_id" {
  type        = string
  description = "Google Cloud project ID to deploy the secret rotator application to."
}

variable "application_name" {
  type        = string
  description = "Name of the application."
  default     = "secrets-rotator"
}

variable "region" {
  type        = string
  description = "Google Cloud location to deploy the service account keys rotator to."
}

variable "service_account_keys_rotator_service_name" {
  type        = string
  description = "Name of the service account keys rotator service instance."
}

variable "service_account_keys_rotator_account_id" {
  type        = string
  default     = "sa-keys-rotator"
  description = "Service account ID of the service account keys rotator."
}

variable "cloud_run_service_listen_port" {
  type        = number
  description = "Port on which the service account keys rotator service listens."
  default     = 8080
}

variable "service_account_keys_rotator_image" {
  type        = string
  description = "Image of the service account keys rotator."
  validation {
    condition     = can(regex("^europe-docker.pkg.dev/kyma.*", var.service_account_keys_rotator_image))
    error_message = "The service account keys rotator image must be hosted in the Kyma Google Artifact Registry."
  }
}

variable "secret_manager_notifications_topic" {
  type        = string
  description = "Name of the topic to which the service account keys rotator subscribes to."
  default     = "secret-manager-notifications"
}
