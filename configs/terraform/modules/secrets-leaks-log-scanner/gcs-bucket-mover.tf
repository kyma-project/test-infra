resource "google_service_account" "gcs_bucket_mover" {
  account_id  = "gcs-bucket-mover"
  description = "Identity of cloud run instance running gcs bucket mover service."
}

resource "google_storage_bucket_iam_member" "kyma_prow_logs_viewer" {
  bucket = data.google_storage_bucket.kyma_prow_logs.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.gcs_bucket_mover.email}"
}

resource "google_storage_bucket_iam_member" "kyma_prow_logs_object_admin" {
  bucket = data.google_storage_bucket.kyma_prow_logs.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.gcs_bucket_mover.email}"
}

resource "google_storage_bucket_iam_member" "kyma_prow_logs_secured_object_admin" {
  bucket = google_storage_bucket.kyma_prow_logs_secured.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.gcs_bucket_mover.email}"
}

resource "google_storage_bucket" "kyma_prow_logs_secured" {
  name                        = "kyma-prow-logs-secured"
  location                    = "EU"
  force_destroy               = true
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
  custom_placement_config {
    data_locations = ["EUROPE-WEST1", "EUROPE-WEST4"]
  }
  public_access_prevention = "enforced"
}

resource "google_cloud_run_service" "gcs_bucket_mover" {
  name     = "gcs-bucket-mover"
  location = "europe-west3"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      service_account_name = google_service_account.gcs_bucket_mover.email
      containers {
        image = "europe-docker.pkg.dev/kyma-project/prod/test-infra/ko/move-gcs-bucket:v20240522-fa411c64"
        env {
          name  = "PROJECT_ID"
          value = var.gcp_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "gcs-bucket-mover"
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
          name  = "DST_BUCKET_NAME"
          value = google_storage_bucket.kyma_prow_logs_secured.name
        }
        env {
          name  = "DRY_RUN"
          value = "true"
        }
      }
    }
  }
}

resource "google_cloud_run_service_iam_policy" "gcs_bucket_mover" {
  location = google_cloud_run_service.gcs_bucket_mover.location
  project  = google_cloud_run_service.gcs_bucket_mover.project
  service  = google_cloud_run_service.gcs_bucket_mover.name

  policy_data = data.google_iam_policy.run_invoker.policy_data
}

resource "google_monitoring_alert_policy" "gcs_bucket_mover" {
  combiner     = "OR"
  display_name = "gcs-bucket-mover-error-logged"
  conditions {
    display_name = "error-log-message"
    condition_matched_log {
      filter = "resource.type=cloud_run_revision AND severity>=ERROR AND jsonPayload.component=gcs-bucket-mover AND labels.io.kyma.app=secrets-leaks-detector"
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
    component = "gcs-bucket-mover"
    app       = "secrets-leak-detector"
  }
}
