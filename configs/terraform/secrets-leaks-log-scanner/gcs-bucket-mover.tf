resource "google_service_account" "gcs_bucket_mover" {
  account_id   = "gcs-bucket-mover-cr"
  description = "Identity of cloud run instance running gcs bucket mover service."
}

resource "google_storage_bucket_iam_member" "kyma_prow_logs_viewer" {
  bucket = data.google_storage_bucket.kyma_prow_logs.name
  role = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.gcs_bucket_mover.email}"
}

resource "google_storage_bucket_iam_member" "kyma_prow_logs_secured_object_admin" {
  bucket = google_storage_bucket.kyma_prow_logs_secured.name
  role = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.gcs_bucket_mover.email}"
}

resource "google_storage_bucket" "kyma_prow_logs_secured" {
  name          = "kyma-prow-logs-secured"
  location      = "EU"
  force_destroy = true
  storage_class = "STANDARD"
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
        image = "europe-docker.pkg.dev/kyma-project/dev/test-infra/movegcsbucket:PR-6689"
        env {
          name  = "PROJECT_ID"
          value = var.google_project_id
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
          value = "dev-prow-logs-secured"
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
