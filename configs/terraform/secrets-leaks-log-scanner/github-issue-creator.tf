resource "google_service_account" "github_issue_creator" {
  account_id   = "github-issue-creator-cr"
  description = "Identity of cloud run instance running github issue creator service."
}

resource "google_secret_manager_secret_iam_member" "gh_issue_creator_gh_tools_kyma_bot_token_accessor" {
  project = data.google_secret_manager_secret.gh_tools_kyma_bot_token.project
  secret_id = data.google_secret_manager_secret.gh_tools_kyma_bot_token.secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.github_issue_creator.email}"
}

resource "google_cloud_run_service" "github_issue_creator" {
  depends_on = [google_secret_manager_secret_iam_member.gh_issue_creator_gh_tools_kyma_bot_token_accessor]
  name     = "github-issue-creator"
  location = "europe-west3"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      service_account_name = google_service_account.github_issue_creator.email
      containers {
        image = "europe-docker.pkg.dev/kyma-project/dev/test-infra/creategithubissue:PR-6801"
        env {
          name  = "PROJECT_ID"
          value = var.google_project_id
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
          name = "TOOLS_GITHUB_TOKEN_PATH"
          value = "/etc/gh-token/gh-tools-kyma-bot-token"
        }
        volume_mounts {
          mount_path = "/etc/gh-token"
          name       = "gh-tools-kyma-bot-token"
        }
      }
      volumes {
        name = "gh-tools-kyma-bot-token"
        secret {
          secret_name = "gh-tools-kyma-bot-token"
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
