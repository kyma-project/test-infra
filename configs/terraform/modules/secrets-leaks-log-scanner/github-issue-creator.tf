resource "google_service_account" "github_issue_creator" {
  account_id  = "github-issue-creator"
  description = "Identity of cloud run instance running github issue creator service."
}

resource "google_secret_manager_secret_iam_member" "gh_issue_creator_gh_tools_kyma_bot_token_accessor" {
  secret_id = data.google_secret_manager_secret.gh_tools_kyma_bot_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.github_issue_creator.email}"
}

resource "google_cloud_run_service" "github_issue_creator" {
  depends_on = [google_secret_manager_secret_iam_member.gh_issue_creator_gh_tools_kyma_bot_token_accessor]
  name       = "github-issue-creator"
  location   = "europe-west3"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      service_account_name = google_service_account.github_issue_creator.email
      containers {
        image = "europe-docker.pkg.dev/kyma-project/prod/create-github-issue:v20250416-974a22b1"
        env {
          name  = "PROJECT_ID"
          value = var.gcp_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "github-issue-creator"
        }
        env {
          name  = "APPLICATION_NAME"
          value = "github-kyma-bot"
        }
        env {
          name  = "LISTEN_PORT"
          value = "8080"
        }
        env {
          name  = "GITHUB_ORG"
          value = "neighbors-team"
        }
        env {
          name  = "GITHUB_REPO"
          value = "leaks-test"
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

resource "google_cloud_run_service_iam_policy" "github_issue_creator" {
  location = google_cloud_run_service.github_issue_creator.location
  project  = google_cloud_run_service.github_issue_creator.project
  service  = google_cloud_run_service.github_issue_creator.name

  policy_data = data.google_iam_policy.run_invoker.policy_data
}

resource "google_monitoring_alert_policy" "github_issue_creator" {
  combiner     = "OR"
  display_name = "github-issue-creator-error-logged"
  conditions {
    display_name = "error-log-message"
    condition_matched_log {
      filter = "resource.type=cloud_run_revision AND severity>=ERROR AND jsonPayload.component=github-issue-creator AND labels.io.kyma.app=secrets-leaks-detector"
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
    component = "github-issue-creator"
    app       = "secrets-leak-detector"
  }
}
