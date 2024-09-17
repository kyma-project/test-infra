data "google_project" "project" {
  provider = google
}

resource "google_service_account" "service_account_keys_cleaner" {
  account_id  = var.service_account_keys_cleaner_account_id
  description = "Identity of the service account keys rotator service."
}

// Allow the service account to delete service account keys.
resource "google_project_iam_member" "service_account_keys_cleaner_sa_keys_admin" {
  project = data.google_project.project.project_id
  role    = "roles/iam.serviceAccountKeyAdmin"
  member  = "serviceAccount:${google_service_account.service_account_keys_cleaner.email}"
}

// Allow the service account to delete secret versions in the secret manager.
resource "google_project_iam_member" "service_account_keys_cleaner_secrets_versions_manager" {
  project = data.google_project.project.project_id
  role    = "roles/secretmanager.secretVersionManager"
  member  = "serviceAccount:${google_service_account.service_account_keys_cleaner.email}"
}

// roles/secretmanager.viewer is required to be able to access the secret in secret manager and read its metadata
resource "google_project_iam_member" "service_account_keys_cleaner_secret_viewer" {
  project = data.google_project.project.project_id
  role    = "roles/secretmanager.viewer"
  member  = "serviceAccount:${google_service_account.service_account_keys_cleaner.email}"
}

// Allow secrets rotator to call the service account keys cleaner service.
resource "google_cloud_run_service_iam_member" "service_account_keys_cleaner_invoker" {
  location = google_cloud_run_service.service_account_keys_cleaner.location
  service  = google_cloud_run_service.service_account_keys_cleaner.name
  project  = google_cloud_run_service.service_account_keys_cleaner.project
  role     = "roles/run.invoker"
  member   = "serviceAccount:${var.secrets_rotator_sa_email}"
}

resource "google_cloud_run_service" "service_account_keys_cleaner" {
  name     = var.service_name
  location = var.region

  template {
    spec {
      service_account_name = google_service_account.service_account_keys_cleaner.email
      containers {
        image = var.service_account_keys_cleaner_image
        env {
          name  = "COMPONENT_NAME"
          value = "service-account-keys-cleaner"
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

resource "google_cloud_scheduler_job" "service_account_keys_cleaner" {
  name             = var.scheduler_name
  region           = var.scheduler_region
  description      = "Call service account keys cleaner service, to remove old versions of secrets"
  schedule         = var.scheduler_cron_schedule
  time_zone        = "Etc/UTC"
  attempt_deadline = "320s"

  http_target {
    http_method = "GET"
    uri         = format("%s/?project=%s&age=%s", google_cloud_run_service.service_account_keys_cleaner.status[0].url, data.google_project.project.project_id, var.service_account_key_latest_version_min_age)

    oidc_token {
      service_account_email = var.secrets_rotator_sa_email
      audience              = google_cloud_run_service.service_account_keys_cleaner.status[0].url
    }
  }
}
