data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth" {
  location = google_cloud_run_service.cors_proxy.location
  project  = google_cloud_run_service.cors_proxy.project
  service  = google_cloud_run_service.cors_proxy.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_cloud_run_service" "cors_proxy" {
  name     = "cors-proxy"
  location = "europe-west3"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      containers {
        image = "europe-docker.pkg.dev/kyma-project/prod/test-infra/ko/cors-proxy:v20240918-39d265ca"
        env {
          name  = "COMPONENT_NAME"
          value = "cors-proxy"
        }
        env {
          name  = "APPLICATION_NAME"
          value = "cors-proxy"
        }
        env {
          name  = "LISTEN_PORT"
          value = "8080"
        }
      }
    }
  }
}
