# Create a service account for Eventarc trigger and Workflows
resource "google_service_account" "release_cluster_created" {
  account_id  = "release-cluster-created"
  description = "Identity of release cluster created eventarc."
}

data "google_iam_policy" "run_invoker" {
  binding {
    role    = "roles/run.invoker"
    members = ["serviceAccount:${google_service_account.release_cluster_created.email}"]
  }
}

# # Grant the logWriter role to the service account
# resource "google_project_iam_member" "project_log_writer" {
#   member  = "serviceAccount:${google_service_account.release_cluster_created.email}"
#   project = var.gcp_project_id
#   role    = "roles/logging.logWriter"
# }

# Grant the workflows.invoker role to the service account
resource "google_project_iam_member" "project_workflows_invoker" {
  project = var.gcp_project_id
  role    = "roles/workflows.invoker"
  member  = "serviceAccount:${google_service_account.release_cluster_created.email}"
}


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
      path    = var.release_cluster_created_could_run_path
    }
  }
  service_account = google_service_account.release_cluster_created.id
  labels = {
    application = "release_cluster_created"
  }
  transport {
    pubsub {
      topic = "projects/${var.gcp_project_id}/topics/${var.release_cluster_created_pubsub_topic_name}"
    }
  }
}
