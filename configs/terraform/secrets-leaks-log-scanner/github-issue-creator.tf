resource "google_service_account" "github_issue_creator" {
  account_id   = "github-issue-creator-cr"
  description = ""
}

resource "google_cloud_run_service" "github_issue_creator" {
  name     = "github-issue-creator"
  location = "europe-west3"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      containers {
        image = "europe-docker.pkg.dev/kyma-project/dev/test-infra/creategithubissue:PR-6676"
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
          name = "TOOLS_SAP_TOKEN_PATH"
          value = "/etc/gh-tools-kyma-bot-token"
        }
        volume_mounts {
          mount_path = "/etc/gh-tools-kyma-bot-token"
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
