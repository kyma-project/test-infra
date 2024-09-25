data "google_project" "project" {
  provider = google
}

resource "google_service_account" "signify_secret_rotator" {
  account_id = var.signify_secret_rotator_account_id
  description = "Identity of the signify secret rotator service."
}

// roles/secretmanager.secretAccessor is required to be able to access the secret version payload in secret manager
resource "google_project_iam_member" "signify_secret_rotator_secret_version_accessor" {
  project = data.google_project.project.project_id
  role = "roles/secretmanager.secretAccessor"

  member = "serviceAccount:${google_service_account.signify_secret_rotator.email}"
}

// roles/secretmanager.secretVersionAdder is required to be able to add new versions to the secret in secret manager
resource "google_project_iam_member" "signify_secret_rotator_secret_version_adder" {
  project = data.google_project.project.project_id
  role = "roles/secretmanager.secretVersionAdder"
  member = "serviceAccount:${google_service_account.signify_secret_rotator.email}"
}

// roles/secretmanager.viewer is required to be able to access the secret in secret manager and read its metadata
resource "google_project_iam_member" "service_account_keys_rotator_secret_version_viewer" {
  project = data.google_project.project.project_id
  role    = "roles/secretmanager.viewer"
  member = "serviceAccount:${google_service_account.signify_secret_rotator.email}"  
}

resource "google_cloud_run_service_iam_member" "signify_secret_rotator_invoker" {
  location = google_cloud_run_service.signify_secret_rotator.location
  service = google_cloud_run_service.signify_secret_rotator.name
  project = google_cloud_run_service.signify_secret_rotator.project
  role = "roles/run.invoker"
  member = "serviceAccount:${var.secrets_rotator_sa_email}"
}

resource "google_cloud_run_service" "signify_secret_rotator" {
  name = var.service_name
  location = var.region

  template {
    spec {
      service_account_name = google_service_account.signify_secret_rotator.email

      containers {
        image = var.signify_secret_rotator_image

        env {
          name = "COMPONENT_NAME"
          value = "signify-secret-rotator"
        }

        env {
          name = "APPLICATION_NAME"
          value = var.application_name
        }

        env {
          name = "LISTEN_PORT"
          value = var.cloud_run_service_listen_port
        }
      }
    }
  }
}

resource "google_pubsub_subscription" "signify_secret_rotator" {
  name = format("%s-%s", var.application_name, var.service_name)

  topic = var.secret_manager_notifications_topic

  ack_deadline_seconds = var.acknowledge_deadline

  labels = {
    application_name = var.application_name
  }

  filter = "attributes.eventType = \"SECRET_ROTATE\""

  push_config {
    push_endpoint = google_cloud_run_service.signify_secret_rotator.status[0].url

    oidc_token {
      service_account_email = var.secrets_rotator_sa_email
    }
  }

  expiration_policy {
    ttl = var.time_to_live
  }

  retry_policy {
    minimum_backoff = var.retry_policy.minimum_backoff
    maximum_backoff = var.retry_policy.maximum_backoff
  }

  dead_letter_policy {
    dead_letter_topic = var.signify_secret_rotator_dead_letter_topic_uri
    max_delivery_attempts = var.dead_letter_maximum_delivery_attempts
  }
}
