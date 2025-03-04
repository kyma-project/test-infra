resource "google_eventarc_trigger" "release_cluster_created" {
  name     = "release-cluster-created"
  location = google_cloud_run_service.slack_message_sender.location
  matching_criteria {
    attribute = "type"
    value     = "google.cloud.pubsub.topic.v1.messagePublished"
  }
  destination {
    cloud_run_service {
      service = google_cloud_run_service.slack_message_sender.name
      region  = google_cloud_run_service.slack_message_sender.location
      path    = var.release_cluster_created_cloud_run_path
    }
  }
  service_account = google_service_account.slack_message_sender.id
  labels = {
    application = "release_cluster_created"
  }
  transport {
    pubsub {
      topic = "projects/${var.gcp_project_id}/topics/${var.release_cluster_created_pubsub_topic_name}"
    }
  }
}
# (2025-03-04)