variable "secret_id" {
  type        = string
  description = "ID of the Secret Manager secret to create."
}

variable "rotation_period" {
  type        = string
  description = "Rotation period for the secret (e.g. '7776000s' for 90 days)."
  default     = "7776000s"
}

variable "next_rotation_time" {
  type        = string
  description = "RFC3339 timestamp for the next rotation. Ignored after initial creation via lifecycle ignore_changes."
}

variable "notification_topic" {
  type        = string
  description = "Full resource name of the Pub/Sub topic to publish rotation events to (e.g. 'projects/my-project/topics/secret-manager-notifications')."
}

variable "service_account_email" {
  type        = string
  description = "Email of the GCP service account whose key is stored in this secret."
}

variable "sa_keys_rotator_sa_email" {
  type        = string
  description = "Email of the SA keys rotator service account."
}

variable "sa_keys_cleaner_sa_email" {
  type        = string
  description = "Email of the SA keys cleaner service account."
}

variable "additional_secret_readers" {
  type        = list(string)
  description = "Additional service account emails that need secretAccessor access to this secret."
  default     = []
}

variable "labels" {
  type        = map(string)
  description = "Additional labels to attach to the secret. The 'type: service-account' label is always added."
  default     = {}
}
