data "google_project" "project" {
  provider = google
}

resource "google_service_account" "service_account_keys_rotator" {
  account_id  = var.service_account_keys_rotator_account_id
  description = "Identity of the service account keys rotator service."
}

// roles/iam.serviceAccountKeyAdmin is required to be able to create new keys for the service account
resource "google_project_iam_member" "service_account_keys_rotator" {
  project = data.google_project.project.project_id
  role    = "roles/iam.serviceAccountKeyAdmin"
  member  = "serviceAccount:${google_service_account.service_account_keys_rotator.email}"
}

// roles/secretmanager.secretAccessor is required to be able to access the secret version payload in secret manager
resource "google_project_iam_member" "service_account_keys_rotator_secret_version_accessor" {
  project = data.google_project.project.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.service_account_keys_rotator.email}"
}

// roles/secretmanager.secretVersionAdder is required to be able to add new versions to the secret in secret manager
resource "google_project_iam_member" "service_account_keys_rotator_secret_version_adder" {
  project = data.google_project.project.project_id
  role    = "roles/secretmanager.secretVersionAdder"
  member  = "serviceAccount:${google_service_account.service_account_keys_rotator.email}"
}

// roles/secretmanager.viewer is required to be able to access the secret in secret manager and read its metadata
resource "google_project_iam_member" "service_account_keys_rotator_secret_version_viewer" {
  project = data.google_project.project.project_id
  role    = "roles/secretmanager.viewer"
  member  = "serviceAccount:${google_service_account.service_account_keys_rotator.email}"
}

resource "google_cloud_run_service_iam_member" "service_account_keys_rotator_invoker" {
  location = google_cloud_run_service.service_account_keys_rotator.location
  service  = google_cloud_run_service.service_account_keys_rotator.name
  project  = google_cloud_run_service.service_account_keys_rotator.project
  role     = "roles/run.invoker"
  member   = "serviceAccount:${var.secrets_rotator_sa_email}"
}

resource "google_project_service_identity" "pubsub_identity_agent" {
  provider = google-beta
  service  = "pubsub.googleapis.com"
}

resource "google_project_iam_binding" "pubsub_project_token_creator" {
  project = data.google_project.project.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  members = ["serviceAccount:${google_project_service_identity.pubsub_identity_agent.email}"]
}

resource "google_cloud_run_service" "service_account_keys_rotator" {
  name     = var.service_name
  location = var.region

  template {
    spec {
      service_account_name = google_service_account.service_account_keys_rotator.email
      containers {
        image = var.service_account_keys_rotator_image
        env {
          name  = "COMPONENT_NAME"
          value = "service-account-keys-rotator"
        }
        env {
          name  = "APPLICATION_NAME"
          value = var.application_name
        }
        env {
          name  = "LISTEN_PORT"
          value = var.cloud_run_service_listen_port
        }
      }
    }
  }
}

resource "google_pubsub_subscription" "service_account_keys_rotator" {
  name                 = format("%s-%s", var.application_name, var.service_name)
  topic                = var.secret_manager_notifications_topic
  ack_deadline_seconds = 20

  labels = {
    application_name = var.application_name
  }

  filter = "attributes.eventType = \"SECRET_ROTATE\""

  push_config {
    push_endpoint = google_cloud_run_service.service_account_keys_rotator.status[0].url
    oidc_token {
      service_account_email = var.secrets_rotator_sa_email
    }
  }

  expiration_policy {
    ttl = "31556952s" // 1 year
  }

  retry_policy {
    minimum_backoff = "300s"
    maximum_backoff = "600s"
  }

  dead_letter_policy {
    dead_letter_topic     = var.service_account_keys_rotator_dead_letter_topic_uri
    max_delivery_attempts = 15
  }
}

# Reference to an existing notification channel
data "google_monitoring_notification_channel" "kyma_tooling" {
  display_name = "Alerting channel for Kyma tooling components."
}

# Log-based alerting policy for service-account-keys-rotator
resource "google_monitoring_alert_policy" "service_account_keys_rotator_error_alert" {
  display_name = "Error detected in service-account-keys-rotator"
  severity     = "ERROR" # Severity level of the alert

  # Define the condition to match logs with errors
  conditions {
    display_name = "Error in service-account-keys-rotator logs"

    condition_matched_log {
      # Filter to match logs from Cloud Run revisions with severity >= ERROR
      filter = <<-EOT
        resource.type="cloud_run_revision"
        AND resource.labels.service_name="service-account-keys-rotator"
        AND severity>=ERROR
      EOT
    }
  }

  # Documentation included in the alert notification
  documentation {
    mime_type = "text/markdown"
    content   = <<-EOT
### Alert: Error detected in service-account-keys-rotator

#### Summary:
A new error has been detected in the Cloud Run service *service-account-keys-rotator*.

#### Error details:
- **Service Name:** $${resource.labels.service_name}
- **Project ID:** $${resource.project}

#### Quick Links:
- [View Logs](https://console.cloud.google.com/logs/query;query=resource.labels.service_name="service-account-keys-rotator")
    EOT
  }

  # Alert strategy configuration
  alert_strategy {
    notification_rate_limit {
      period = "259200s" # Minimum time between notifications (3 days)
    }
    auto_close = "604800s" # Automatically close incidents after 7 days of no matching logs
  }

  # Use the existing notification channel for alerts
  notification_channels = [data.google_monitoring_notification_channel.kyma_tooling.id]
}
