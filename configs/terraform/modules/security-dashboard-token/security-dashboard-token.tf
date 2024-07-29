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
  name     = "security-dashboard-token"
  location = "europe-west1"

  // FIX(long-apply): See https://github.com/hashicorp/terraform-provider-google/issues/5898#issuecomment-605062566
  autogenerate_revision_name = true

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      containers {
        image = "europe-docker.pkg.dev/kyma-project/prod/test-infra/ko/dashboard-token-proxy:v20240729-5b8c0d0f" #gitleaks:allow ignore gitleaks detection
        env {
          name = "CLIENT_SECRET"
          value_from {
            secret_key_ref {
              key  = "latest"
              name = "security-dashboard-oauth-client-secret"
            }
          }
        }
        env {
          name = "CLIENT_ID"
          value_from {
            secret_key_ref {
              key  = "latest"
              name = "security-dashboard-oauth-client-id"
            }
          }
        }

        env {
          name = "GH_BASE_URL"
          value_from {
            secret_key_ref {
              key  = "latest"
              name = "gh-internal-url"
            }
          }
        }
      }
    }
  }
}
