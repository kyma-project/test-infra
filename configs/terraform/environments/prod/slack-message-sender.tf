module "slack_message_sender" {

  providers = {
    google = google
  }
  source         = "../../modules/slack-message-sender"
  gcp_project_id = var.gcp_project_id
}
# (2025-03-04)