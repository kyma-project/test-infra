variable "region" {
  type        = string
  description = "Google Cloud region to deploy the signify secret rotator to."
}

variable "service_name" {
  type        = string
  description = "Name of the signify secret rotator service instance."
}

variable "application_name" {
  type        = string
  description = "Name of the application a signify secret rotator is a part of."
}

variable "signify_secret_rotator_account_id" {
  type        = string
  default     = "sa-signify-secret-rotator"
  description = "Service account ID of the signify secret rotator."
}

variable "cloud_run_service_listen_port" {
  type        = number
  description = "Port on which the signify secret rotator service listens."
  default     = 8080
}

variable "signify_secret_rotator_image" {
  type        = string
  description = "Image of the signify secret rotator."
  validation {
    condition     = can(regex("^europe-docker.pkg.dev/kyma.*", var.signify_secret_rotator_image))
    error_message = "The signify secret rotator image must be hosted in the Kyma Google Artifact Registry."
  }
}

variable "secret_manager_notifications_topic" {
  type        = string
  description = "Name of the topic to which the signify secret rotator subscribes to."
  default     = "secret-manager-notifications"
}

variable "signify_secret_rotator_dead_letter_topic_uri" {
  type        = string
  description = "URI of the topic to which the signify secret rotator publishes dead letter messages."
}

variable "create_secret_manager_notifications_topic" {
  type        = bool
  description = "Whether to create the topic to which the signify secret rotator subscribes to."
  default     = false
}

variable "secrets_rotator_sa_email" {
  type        = string
  description = "Secrets rotator application service account email."
}

variable "acknowledge_deadline" {
  type = number
  description = "This value in seconds is the maximum time after a subscriber receives a message before the subscriber should acknowledge the message. For push delivery, this value is also used to set the request timeout for the call to the push endpoint."
  default = 20
}

variable "time_to_live" {
  type = string
  description = "After that time the inactive pubsub subscription expires."
  default = "31556952s" // 1 year
}

variable "retry_policy" {
  type = object({
    minimum_backoff = string
    maximum_backoff = string
  })
  description = "A policy that specifies how Pub/Sub retries message delivery for subscription."
  default = {
    minimum_backoff = "300s"
    maximum_backoff = "600s"
  }
}

variable "dead_letter_maximum_delivery_attempts" {
  type = number
  description = "Maximum attempts of delivering the dead letter"
  default = 15
}
