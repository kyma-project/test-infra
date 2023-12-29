data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth" {
  location = google_cloud_run_service.security_dashboard_token.location
  project  = google_cloud_run_service.security_dashboard_token.project
  service  = google_cloud_run_service.security_dashboard_token.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_cloud_run_service" "security_dashboard_token" {
  name     = "security_dashboard_token"
  location = "europe-west1"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      containers {
        image = "europe-west3-docker.pkg.dev/sap-kyma-neighbors-dev/kyma-neighbors-dev-test/security-dashboard-token:0.0.2"
        env {
          name = "CLIENT_SECRET"
          value_from {
            secret_key_ref {
              key = "latest"
              name = "security-dashboard-oauth-client-secret"
            }
          }
        }
        env {
          name = "CLIENT_ID"
          value_from {
            secret_key_ref {
              key = "latest"
              name = "security-dashboard-oauth-client-id"
            }
          }
        }
      }
    }
  }
}
