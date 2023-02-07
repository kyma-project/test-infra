resource "google_service_account" "secrets_leak_log_scanner" {
  account_id   = "secrets-leak-log-scanner-cr"
  description = "Identity of cloud run instance running log scanner service."
}

resource "google_storage_bucket_iam_member" "secrets_leak_detector" {
  bucket = data.google_storage_bucket.kyma_prow_logs.name
  role = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.secrets_leak_detector.email}"
}

resource "google_cloud_run_service" "secrets_leak_log_scanner" {
  name     = "secrets-leak-log-scanner"
  location = "europe-west3"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      service_account_name = google_service_account.secrets_leak_log_scanner.email
      containers {
        image = "europe-docker.pkg.dev/kyma-project/dev/test-infra/scanlogsforsecrets:PR-6684"
        env {
          name  = "PROJECT_ID"
          value = var.google_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "secrets-leak-log-scanner"
        }
        env {
          name  = "APPLICATION_NAME"
          value = "secrets-leaks-detector"
        }
        env {
          name  = "LISTEN_PORT"
          value = "8080"
        }
        env {
          name  = "GCS_PREFIX"
          value = "gcsweb.build.kyma-project.io/gcs/"
        }
      }
    }
  }
}

resource "google_cloud_run_service_iam_policy" "secrets_leak_log_scanner" {
  location = google_cloud_run_service.secrets_leak_log_scanner.location
  project  = google_cloud_run_service.secrets_leak_log_scanner.project
  service  = google_cloud_run_service.secrets_leak_log_scanner.name

  policy_data = data.google_iam_policy.run_invoker.policy_data
}
