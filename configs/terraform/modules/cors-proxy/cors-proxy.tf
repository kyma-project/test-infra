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
        image = "europe-docker.pkg.dev/kyma-project/prod/test-infra/ko/github-webhook-gateway:v20230816-1f308fb9"
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
