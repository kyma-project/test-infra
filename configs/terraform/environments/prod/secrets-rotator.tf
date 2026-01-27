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

module "signify_secret_rotator" {
  source = "../../modules/signify-secret-rotator"

  application_name = var.secrets_rotator_name
  service_name     = var.signify_secret_rotator_service_name

  region                                       = var.gcp_region
  signify_secret_rotator_account_id            = var.signify_secret_rotator_account_id
  signify_secret_rotator_dead_letter_topic_uri = google_pubsub_topic.secrets_rotator_dead_letter.id
  signify_secret_rotator_image                 = var.signify_secret_rotator_image
  cloud_run_service_listen_port                = var.secrets_rotator_cloud_run_listen_port
  secret_manager_notifications_topic           = var.secret_manager_notifications_topic
  secrets_rotator_sa_email                     = google_service_account.secrets-rotator.email
}


// ### dead letter monitoring ###

resource "google_pubsub_subscription" "secrets-rotator-dead-letter" {
  name  = "secrets-rotator-dead-letter"
  topic = google_pubsub_topic.secrets_rotator_dead_letter.id

  expiration_policy {
    ttl = ""
  }
  message_retention_duration = "864000s" // 10 days

  retry_policy {
    minimum_backoff = "1s" // fast start, so the incident is closable ASAP
    maximum_backoff = "600s"
  }

  cloud_storage_config {
    bucket = google_storage_bucket.secret-rotator-dead-letters-bucket.name
  }
}

// bucket

resource "google_storage_bucket" "secret-rotator-dead-letters-bucket" {
  name          = "secret-rotator-dead-letters"
  location      = "EU"
  force_destroy = true

  uniform_bucket_level_access = true
}

resource "google_storage_bucket_iam_member" "dead-letter-bucket-access" {
  bucket = google_storage_bucket.secret-rotator-dead-letters-bucket.name
  for_each = toset(["roles/storage.legacyBucketReader", "roles/storage.objectCreator"])
  role   = each.value
  member = "serviceAccount:service-${data.google_project.current.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

// alert

resource "google_monitoring_alert_policy" "dead-letter-alert" {
  display_name = "secrets-rotator-dead-letter"
  combiner     = "OR"
  severity = "ERROR"
  notification_channels = [
    "projects/${var.gcp_project_id}/notificationChannels/1439001756543663676",
    "projects/${var.gcp_project_id}/notificationChannels/17517114611086313455"
  ]
  conditions {
    display_name = "Cloud Pub/Sub Subscription - Dead letter message count"
    condition_threshold {
      filter     = "resource.type = \"pubsub_subscription\" AND (resource.labels.subscription_id = \"secrets-rotator-service-account-keys-rotator\" AND resource.labels.project_id = \"sap-kyma-prow\") AND metric.type = \"pubsub.googleapis.com/subscription/dead_letter_message_count\""
      duration   = "60s"
      comparison = "COMPARISON_GT"
      aggregations {
        alignment_period   = "60s"
        per_series_aligner = "ALIGN_COUNT"
      }
    }
  }
}