resource "google_service_account" "secrets_leak_log_scanner" {
  account_id  = "secrets-leak-log-scanner"
  description = "Identity of cloud run instance running log scanner service."
}

resource "google_storage_bucket_iam_member" "secrets_leak_detector" {
  bucket = data.google_storage_bucket.kyma_prow_logs.name
  role   = "roles/storage.objectViewer"
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
        image = "europe-docker.pkg.dev/kyma-project/prod/test-infra/ko/scan-logs-for-secrets:v20240318-7c1da83c" #nosec
        env {
          name  = "PROJECT_ID"
          value = var.gcp_project_id
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

resource "google_monitoring_alert_policy" "secrets_leak_log_scanner" {
  combiner     = "OR"
  display_name = "secrets-leak-log-scanner-error-logged"
  conditions {
    display_name = "error-log-message"
    condition_matched_log {
      filter = "resource.type=cloud_run_revision AND severity>=ERROR AND jsonPayload.component=secrets-leak-log-scanner AND labels.io.kyma.app=secrets-leaks-detector"
    }
  }
  notification_channels = ["projects/${var.gcp_project_id}/notificationChannels/5909844679104799956"]
  alert_strategy {
    notification_rate_limit {
      period = "21600s"
    }
    auto_close = "345600s"
  }
  user_labels = {
    component = "secrets-leak-log-scanner"
    app       = "secrets-leak-detector"
  }
}
