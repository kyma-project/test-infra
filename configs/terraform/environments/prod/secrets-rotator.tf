resource "google_pubsub_topic" "secrets_rotator_dead_letter" {
  name = format("%s-%s", var.secrets_rotator_name, "dead-letter")

  labels = {
    application = var.secrets_rotator_name
  }

  message_retention_duration = "86600s"
}

resource "google_service_account" "secrets-rotator" {
  account_id   = "secrets-rotator"
  display_name = "secrets-rotator"
  description  = "Identity of the secrets rotator application"
}

data "google_pubsub_topic" "secret-manager-notifications-topic" {
  name = var.secret_manager_notifications_topic
}

module "service_account_keys_rotator" {
  source = "../../modules/rotate-service-account"

  application_name = var.secrets_rotator_name
  service_name     = var.service_account_keys_rotator_service_name

  region                                             = var.gcp_region
  service_account_keys_rotator_account_id            = var.service_account_keys_rotator_account_id
  service_account_keys_rotator_dead_letter_topic_uri = google_pubsub_topic.secrets_rotator_dead_letter.id
  service_account_keys_rotator_image                 = var.service_account_keys_rotator_image
  cloud_run_service_listen_port                      = var.secrets_rotator_cloud_run_listen_port
  secret_manager_notifications_topic                 = var.secret_manager_notifications_topic
  secrets_rotator_sa_email                           = google_service_account.secrets-rotator.email
}

output "service_account_keys_rotator" {
  value = module.service_account_keys_rotator
}

resource "google_project_iam_member" "service_account_keys_rotator_workloads_project" {
  provider = google.workloads
  project  = var.workloads_project_id
  role     = "roles/iam.serviceAccountKeyAdmin"
  member   = "serviceAccount:${module.service_account_keys_rotator.service_account_keys_rotator_service_account.email}"
}

module "service_account_keys_cleaner" {
  source = "../../modules/service-account-keys-cleaner"

  application_name = var.secrets_rotator_name
  service_name     = var.service_account_keys_cleaner_service_name

  region                                     = var.gcp_region
  scheduler_region                           = var.gcp_scheduler_region
  service_account_keys_cleaner_account_id    = var.service_account_keys_cleaner_account_id
  service_account_keys_cleaner_image         = var.service_account_keys_cleaner_image
  cloud_run_service_listen_port              = var.secrets_rotator_cloud_run_listen_port
  scheduler_name                             = var.service_account_keys_cleaner_service_name
  secrets_rotator_sa_email                   = google_service_account.secrets-rotator.email
  scheduler_cron_schedule                    = var.service_account_keys_cleaner_scheduler_cron_schedule
  service_account_key_latest_version_min_age = var.service_account_key_latest_version_min_age
}

output "service_account_keys_cleaner" {
  value = module.service_account_keys_cleaner
}

resource "google_project_iam_member" "service_account_keys_cleaner_workloads_project" {
  provider = google.workloads
  project  = var.workloads_project_id
  role     = "roles/iam.serviceAccountKeyAdmin"
  member   = "serviceAccount:${module.service_account_keys_cleaner.service_account_keys_cleaner_service_account.email}"
}
