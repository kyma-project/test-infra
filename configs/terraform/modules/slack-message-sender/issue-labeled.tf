resource "google_eventarc_trigger" "issue_labeled" {
  name     = "issue-labeled"
  location = google_cloud_run_service.slack_message_sender.location
  matching_criteria {
    attribute = "type"
    value     = "google.cloud.pubsub.topic.v1.messagePublished"
  }
  destination {
    cloud_run_service {
      service = google_cloud_run_service.slack_message_sender.name
      region  = google_cloud_run_service.slack_message_sender.location
      path    = var.issue_labeled_cloud_run_path
    }
  }
  service_account = google_service_account.slack_message_sender.id
  transport {
    pubsub {
      topic = "projects/${var.gcp_project_id}/topics/${var.issue_labeled_pubsub_topic_name}"
    }
  }
}
