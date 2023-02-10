resource "google_service_account" "github_issue_finder" {
  account_id   = "github-issue-finder-cr"
  description = "Identity of cloud run instance running github issue finder service."
}

resource "google_secret_manager_secret_iam_member" "gh_issue_finder_gh_tools_kyma_bot_token_accessor" {
  project = data.google_secret_manager_secret.gh_tools_kyma_bot_token.project
  secret_id = data.google_secret_manager_secret.gh_tools_kyma_bot_token.secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.github_issue_finder.email}"
}

resource "google_cloud_run_service" "github_issue_finder" {
  depends_on = [google_secret_manager_secret_iam_member.gh_issue_finder_gh_tools_kyma_bot_token_accessor]
  name     = "github-issue-finder"
  location = "europe-west3"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      service_account_name = google_service_account.github_issue_finder.email
      containers {
        image = "europe-docker.pkg.dev/kyma-project/prod/test-infra/searchgithubissue:v20230202-40569193"
        env {
          name  = "PROJECT_ID"
          value = var.gcp_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "github-issue-finder"
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

resource "google_cloud_run_service_iam_policy" "github_issue_finder" {
  location = google_cloud_run_service.github_issue_finder.location
  project  = google_cloud_run_service.github_issue_finder.project
  service  = google_cloud_run_service.github_issue_finder.name

  policy_data = data.google_iam_policy.run_invoker.policy_data
}
