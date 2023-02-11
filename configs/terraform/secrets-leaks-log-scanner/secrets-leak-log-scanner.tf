resource "google_service_account" "secrets_leak_log_scanner" {
  account_id  = "secrets-leak-log-scanner-cr"
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

resource "google_monitoring_alert_policy" "secrets_leak_log_scanner" {
  combiner     = "OR"
  display_name = "secrets-leak-log-scanner-error-logged"
  conditions {
    display_name = "error-log-message"
    condition_matched_log {
      filter = "resource.type=cloud_run_revision AND severity>=ERROR AND jsonPayload.component=secrets-leak-log-scanner AND labels.io.kyma.app=secrets-leaks-detector"
    }
  }
  notification_channels = ["projects/${var.google_project_id}/notificationChannels/5909844679104799956"]
  alert_strategy {
    notification_rate_limit {
      period = "6 hr"
    }
    auto_close = "4 days"
  }
  user_labels = {
    component = "secrets-leak-log-scanner"
    app       = "secrets-leak-detector"
  }
}

#resource "google_logging_metric" "secrets_leak_log_scanner" {
#  name = "secrests_leak_log_scanner_errors"
#  filter = "resource.type=cloud_run_revision AND severity>=ERROR AND jsonPayload.component=secrets-leak-log-scanner AND labels.io.kyma.app=secrets-leaks-detector"
#  metric_descriptor {
#    metric_kind = "DELTA"
#    value_type  = "INT64"
#    labels {
#      key = "component"
#      value_type = "STRING"
#      description = ""
#    }
#    labels {
#      key = "app"
#      value_type = "STRING"
#      description = ""
#    }
#  }
#  label_extractors = {
#    "component" = "EXTRACT(jsonPayload.component)"
#    "app" = "EXTRACT(labels.io.kyma.app)"
#  }
#}
