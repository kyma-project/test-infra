resource "google_service_account" "slack_message_sender" {
  account_id  = "slack-message-sender"
  description = "Identity of cloud run instance running slack message sender service."
}

resource "google_project_iam_member" "project_run_invoker" {
  project = var.gcp_project_id
  role    = "roles/run.invoker"
  member  = "serviceAccount:${google_service_account.slack_message_sender.email}"
}

data "google_iam_policy" "run_invoker" {
  binding {
    role    = "roles/run.invoker"
    members = ["serviceAccount:${google_service_account.slack_message_sender.email}"]
  }
}


resource "google_secret_manager_secret_iam_member" "slack_msg_sender_common_slack_bot_token_accessor" {
  secret_id = data.google_secret_manager_secret.common_slack_bot_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.slack_message_sender.email}"
}

resource "google_cloud_run_service" "slack_message_sender" {
  depends_on = [google_secret_manager_secret_iam_member.slack_msg_sender_common_slack_bot_token_accessor]
  name       = "slack-message-sender"
  location   = "europe-west3"

  metadata {
    annotations = {
      "run.googleapis.com/client-name" = "terraform"
    }
  }

  template {
    spec {
      service_account_name = google_service_account.slack_message_sender.email
      containers {
        image = "europe-docker.pkg.dev/kyma-project/prod/test-infra/slackmessagesender:v20240423-031a88bf"
        env {
          name  = "PROJECT_ID"
          value = var.gcp_project_id
        }
        env {
          name  = "COMPONENT_NAME"
          value = "slack-message-sender"
        }
        env {
          name  = "APPLICATION_NAME"
          value = "slack-kyma-bot"
        }
        env {
          name  = "PROW_DEV_NULL_SLACK_CHANNEL_ID"
          value = "C01KSP10MB5" # kyma-prow-dev-null
        }
        env {
          name  = "RELEASE_SLACK_CHANNEL_ID"
          value = "C01KKPXCPK8" # kyma-skr-release
        }
        env {
          name  = "KYMA_TEAM_SLACK_CHANNEL_ID"
          value = "C01LGCBS196" # kyma-team
        }
        env {
          name  = "SLACK_BASE_URL"
          value = "https://slack.com/api"
        }
        env {
          name  = "SLACK_TOKEN_PATH"
          value = "/etc/slack-secret/${data.google_secret_manager_secret.common_slack_bot_token.secret_id}"
        }
        volume_mounts {
          # TODO: change mount path to slack-token after updating a slackmessagesender cloud run.
          mount_path = "/etc/slack-secret"
          name       = "slack-token"
        }
      }
      volumes {
        name = "slack-token"
        secret {
          secret_name = data.google_secret_manager_secret.common_slack_bot_token.secret_id
        }
      }
    }
  }
}

resource "google_cloud_run_service_iam_policy" "slack_message_sender" {
  location = google_cloud_run_service.slack_message_sender.location
  project  = google_cloud_run_service.slack_message_sender.project
  service  = google_cloud_run_service.slack_message_sender.name

  policy_data = data.google_iam_policy.run_invoker.policy_data
}

resource "google_monitoring_alert_policy" "slack_message_sender" {
  combiner     = "OR"
  display_name = "slack-message-sender-error-logged"
  conditions {
    display_name = "error-log-message"
    condition_matched_log {
      filter = "resource.type=cloud_run_revision AND severity>=ERROR AND jsonPayload.component=slack-message-sender"
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
    component = "slack-message-sender"
    app       = "slack-messagesender"
  }
}
