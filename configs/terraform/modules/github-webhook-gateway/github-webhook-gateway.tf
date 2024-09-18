


resource "google_service_account" "github_webhook_gateway" {
  account_id  = "github-webhook-gateway"
  description = "Identity of cloud run instance running github webhook gateway service."
}


resource "google_secret_manager_secret_iam_member" "gh_tools_kyma_bot_token_accessor" {
  secret_id = data.google_secret_manager_secret.gh_tools_kyma_bot_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.github_webhook_gateway.email}"
}

resource "google_secret_manager_secret_iam_member" "webhook_token_accessor" {
  secret_id = data.google_secret_manager_secret.webhook_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.github_webhook_gateway.email}"
}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth" {
  location = google_cloud_run_service.github_webhook_gateway.location
  project  = google_cloud_run_service.github_webhook_gateway.project
  service  = google_cloud_run_service.github_webhook_gateway.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_pubsub_topic" "issue_labeled" {
  name = var.pubsub_topic_name
}

resource "google_pubsub_topic_iam_binding" "issue_labeled" {
  project = google_pubsub_topic.issue_labeled.project
  topic   = google_pubsub_topic.issue_labeled.name
  role    = "roles/pubsub.publisher"
  members = [
    "serviceAccount:${google_service_account.github_webhook_gateway.email}",
  ]
}

resource "google_cloud_run_service" "github_webhook_gateway" {
  depends_on = [
    google_secret_manager_secret_iam_member.gh_tools_kyma_bot_token_accessor,
    google_secret_manager_secret_iam_member.webhook_token_accessor,
  ]
  name     = "github-webhook-gateway"
  location = "europe-west3"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      service_account_name = google_service_account.github_webhook_gateway.email
      containers {
        image = "europe-docker.pkg.dev/kyma-project/prod/test-infra/ko/github-webhook-gateway:v20240918-39d265ca"
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
          value = google_pubsub_topic.issue_labeled.name
        }
        env {
          name  = "TOOLS_GITHUB_TOKEN_PATH"
          value = "/etc/gh-token/${data.google_secret_manager_secret.gh_tools_kyma_bot_token.secret_id}"
        }
        env {
          name  = "WEBHOOK_TOKEN_PATH"
          value = "/etc/webhook-token/${data.google_secret_manager_secret.webhook_token.secret_id}"
        }
        volume_mounts {
          mount_path = "/etc/gh-token"
          name       = "gh-tools-kyma-bot-token"
        }
        volume_mounts {
          mount_path = "/etc/webhook-token"
          name       = "webhook-token"
        }
      }
      volumes {
        name = "gh-tools-kyma-bot-token"
        secret {
          secret_name = data.google_secret_manager_secret.gh_tools_kyma_bot_token.secret_id
        }
      }
      volumes {
        name = "webhook-token"
        secret {
          secret_name = data.google_secret_manager_secret.webhook_token.secret_id
        }
      }
    }
  }
}
