


resource "google_service_account" "github_webhook_gateway" {
  account_id  = "github-webhook-gateway"
  description = "Identity of cloud run instance running github webhook gateway service."
}


resource "google_secret_manager_secret_iam_member" "gh_issue_finder_gh_tools_kyma_bot_token_accessor" {
  secret_id = data.google_secret_manager_secret.gh_tools_kyma_bot_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.github_webhook_gateway.email}"
}


resource "google_pubsub_topic" "issue_labeled" {
  name = var.pubsub_topic_name
}

resource "google_cloud_run_service" "github_webhook_gateway" {
  depends_on = [google_secret_manager_secret_iam_member.gh_issue_finder_gh_tools_kyma_bot_token_accessor]
  name       = "github-webhook-gateway"
  location   = "europe-west3"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      service_account_name = google_service_account.github_webhook_gateway.email
      containers {
        image = "europe-docker.pkg.dev/kyma-project/prod/test-infra/ko/github-webhook-gateway:v20230808-b281686e"
        env {
          name  = "PROJECT_ID"
          value = var.gcp_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "github-webhook-gateway"
        }
        env {
          name  = "APPLICATION_NAME"
          value = "github-webhook-gateway"
        }
        env {
          name  = "LISTEN_PORT"
          value = "8080"
        }
        env {
          name  = "PUBSUB_TOPIC"
          value = google_pubsub_topic.issue_labeled.id
        }
        env {
          name  = "TOOLS_GITHUB_TOKEN_PATH"
          value = "/etc/gh-token/${data.google_secret_manager_secret.gh_tools_kyma_bot_token.secret_id}"
        }
        volume_mounts {
          mount_path = "/etc/gh-token"
          name       = "gh-tools-kyma-bot-token"
        }
      }
      volumes {
        name = "gh-tools-kyma-bot-token"
        secret {
          secret_name = data.google_secret_manager_secret.gh_tools_kyma_bot_token.secret_id
        }
      }
    }
  }
}
