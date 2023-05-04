data "google_project" "project" {
  project_id = var.project_id
}

resource "google_pubsub_topic" "secrets_rotator_dead_letter" {
  name    = format("%s-%s", var.application_name, "dead-letter")
  project = data.google_project.project.project_id

  labels = {
    application = var.application_name
  }

  message_retention_duration = "86600s"
}

output "secrets_rotator_dead_letter_topic" {
  value = google_pubsub_topic.secrets_rotator_dead_letter
}

resource "google_service_account" "secrets-rotator" {
  account_id   = "secrets-rotator"
  project      = data.google_project.project.project_id
  display_name = "Identity of the secrets rotator application"
}

output "secrets-rotator" {
  value = google_service_account.secrets-rotator
}

data "google_pubsub_topic" "secret-manager-notifications-topic" {
  name    = var.secret_manager_notifications_topic
  project = data.google_project.project.project_id
}

output "secret-manager-notifications-topic" {
  value = data.google_pubsub_topic.secret-manager-notifications-topic
}

module "service_account_keys_rotator" {
  source = "../../modules/rotate-service-account"

  application_name = var.application_name
  service_name     = var.service_account_keys_rotator_service_name
  project = {
    id     = data.google_project.project.project_id
    number = data.google_project.project.number
  }
  region = var.region

  service_account_keys_rotator_account_id            = var.service_account_keys_rotator_account_id
  service_account_keys_rotator_dead_letter_topic_uri = google_pubsub_topic.secrets_rotator_dead_letter.id
  service_account_keys_rotator_image                 = var.service_account_keys_rotator_image
  cloud_run_service_listen_port                      = var.cloud_run_service_listen_port
  secret_manager_notifications_topic                 = var.secret_manager_notifications_topic
  secrets_rotator_sa_email                           = google_service_account.secrets-rotator.email
}

output "service_account_keys_rotator" {
  value = module.service_account_keys_rotator
}

module "service_account_keys_cleaner" {
  source = "../../modules/service-account-keys-cleaner"

  application_name = var.application_name
  service_name     = var.service_account_keys_cleaner_service_name
  project = {
    id     = data.google_project.project.project_id
    number = data.google_project.project.number
  }
  region                                     = var.region
  service_account_keys_cleaner_account_id    = var.service_account_keys_cleaner_account_id
  service_account_keys_cleaner_image         = var.service_account_keys_cleaner_image
  cloud_run_service_listen_port              = var.cloud_run_service_listen_port
  scheduler_name                             = var.service_account_keys_cleaner_service_name
  secrets_rotator_sa_email                   = google_service_account.secrets-rotator.email
  scheduler_cron_schedule                    = var.service_account_keys_cleaner_scheduler_cron_schedule
  service_account_key_latest_version_min_age = var.service_account_key_latest_version_min_age
}

output "service_account_keys_cleaner" {
  value = module.service_account_keys_cleaner
}
