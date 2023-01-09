terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-neighbors-dev"
    prefix = "terraform/state"
  }
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.47.0"
    }
  }
}

variable "google_project_id" {
  type    = string
  default = "sap-kyma-neighbors-dev"
}

provider "google" {
  project = var.google_project_id
  region  = "europe-west1"
  zone    = "europe-west1-b"
}

data "google_iam_policy" "noauth" {
  binding {
    role    = "roles/run.invoker"
    members = ["allUsers"]
  }
}

resource "google_cloud_run_service" "scan-logs-for-secrets" {
  name     = "scan-logs-for-secrets"
  location = "europe-west1"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      containers {
        image = "europe-west3-docker.pkg.dev/sap-kyma-neighbors-dev/kyma-neighbors-dev-test/scanlogsforsecrets:0.0.49"
        env {
          name  = "PROJECT_ID"
          value = var.google_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "logs-scanner"
        }
        env {
          name  = "APPLICATION_NAME"
          value = "secrets-leaks-log-scanner"
        }
        env {
          name  = "LISTEN_PORT"
          value = "8080"
        }
        env {
          name  = "GCS_PREFIX"
          value = "gcsweb.build.kyma-project.io/gcs/"
        }
      }
    }
  }
}

resource "google_cloud_run_service_iam_policy" "scan-logs-for-secrets-noauth" {
  location = google_cloud_run_service.scan-logs-for-secrets.location
  project  = google_cloud_run_service.scan-logs-for-secrets.project
  service  = google_cloud_run_service.scan-logs-for-secrets.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_cloud_run_service" "move-gcs-bucket" {
  name     = "move-gcs-bucket"
  location = "europe-west1"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      containers {
        image = "europe-west3-docker.pkg.dev/sap-kyma-neighbors-dev/kyma-neighbors-dev-test/movegcsbucket:0.0.10"
        env {
          name  = "PROJECT_ID"
          value = var.google_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "bucket-mover"
        }
        env {
          name  = "APPLICATION_NAME"
          value = "secrets-leaks-log-scanner"
        }
        env {
          name  = "LISTEN_PORT"
          value = "8080"
        }
        env {
          name  = "DST_BUCKET_NAME"
          value = "dev-prow-logs-secured"
        }
      }
    }
  }
}

resource "google_cloud_run_service_iam_policy" "move-gcs-bucket-noauth" {
  location = google_cloud_run_service.move-gcs-bucket.location
  project  = google_cloud_run_service.move-gcs-bucket.project
  service  = google_cloud_run_service.move-gcs-bucket.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_cloud_run_service" "search-github-issue" {
  name     = "search-github-issue"
  location = "europe-west1"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      containers {
        image = "europe-west3-docker.pkg.dev/sap-kyma-neighbors-dev/kyma-neighbors-dev-test/searchgithubissue:0.0.20"
        env {
          name  = "PROJECT_ID"
          value = var.google_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "issue-finder"
        }
        env {
          name  = "APPLICATION_NAME"
          value = "secrets-leaks-log-scanner"
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
        volume_mounts {
          mount_path = "/etc/github-token"
          name       = "github-token"
        }
      }
      volumes {
        name = "github-token"
        secret {
          secret_name = "github-token"
        }
      }
    }
  }
}

resource "google_cloud_run_service_iam_policy" "search-github-issue-noauth" {
  location = google_cloud_run_service.search-github-issue.location
  project  = google_cloud_run_service.search-github-issue.project
  service  = google_cloud_run_service.search-github-issue.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_cloud_run_service" "create-github-issue" {
  name     = "create-github-issue"
  location = "europe-west1"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      containers {
        image = "europe-west3-docker.pkg.dev/sap-kyma-neighbors-dev/kyma-neighbors-dev-test/creategithubissue:0.0.16"
        env {
          name  = "PROJECT_ID"
          value = var.google_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "issue-creator"
        }
        env {
          name  = "APPLICATION_NAME"
          value = "secrets-leaks-log-scanner"
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
        volume_mounts {
          mount_path = "/etc/github-token"
          name       = "github-token"
        }
      }
      volumes {
        name = "github-token"
        secret {
          secret_name = "github-token"
        }
      }
    }
  }
}

resource "google_cloud_run_service_iam_policy" "create-github-issue-noauth" {
  location = google_cloud_run_service.create-github-issue.location
  project  = google_cloud_run_service.create-github-issue.project
  service  = google_cloud_run_service.create-github-issue.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_cloud_run_service" "send-slack-message" {
  name     = "send-slack-message"
  location = "europe-west1"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      containers {
        image = "europe-west3-docker.pkg.dev/sap-kyma-neighbors-dev/kyma-neighbors-dev-test/slackmessagesender:0.0.16"
        env {
          name  = "PROJECT_ID"
          value = var.google_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "message-sender"
        }
        env {
          name  = "APPLICATION_NAME"
          value = "kyma-slack-bot"
        }
        env {
          name  = "SLACK_CHANNEL_ID"
          value = "C01KSP10MB5"
        }
        env {
          name  = "SLACK_BASE_URL"
          value = "https://slack.com/api"
        }
        volume_mounts {
          mount_path = "/etc/slack-secret"
          name       = "slack-secret"
        }
      }
      volumes {
        name = "slack-secret"
        secret {
          secret_name = "common-slack-bot-token"
        }
      }
    }
  }
}

resource "google_cloud_run_service_iam_policy" "send-slack-message-noauth" {
  location = google_cloud_run_service.send-slack-message.location
  project  = google_cloud_run_service.send-slack-message.project
  service  = google_cloud_run_service.send-slack-message.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

data "template_file" "scan-logs-for-secrets_yaml" {
  template = file("${path.module}/../../../development/gcp/workflows/scan-logs-for-secrets.yaml")
  vars = {
    scan-logs-for-secrets-url = google_cloud_run_service.scan-logs-for-secrets.status[0].url
    move-gcs-bucket-url       = google_cloud_run_service.move-gcs-bucket.status[0].url
    search-github-issue-url   = google_cloud_run_service.search-github-issue.status[0].url
    create-github-issue-url   = google_cloud_run_service.create-github-issue.status[0].url
    send-slack-message-url    = google_cloud_run_service.send-slack-message.status[0].url
  }
}

resource "google_workflows_workflow" "scan-logs-for-secrets" {
  name        = "poc-scan-logs-for-secrets"
  region      = "europe-west4"
  description = "Workflow is triggered on pubsub ..."
  # service_account = google_service_account.workflows_service_account.id
  source_contents = data.template_file.scan-logs-for-secrets_yaml.rendered

}
