# Setups IAM for the project administrators in the kyma-project GCP project.
resource "google_project_iam_member" "kyma_developer_admins_editor" {
  provider = google.kyma_project
  project = var.kyma_project_gcp_project_id
  role = "roles/editor"
  member = "group:${var.kyma_developer_admins_email}"
}

# Add roles required to see audit logs in kyma-project GCP project.
resource "google_project_iam_member" "kyma_developer_admins_logging_viewer" {
  provider = google.kyma_project
  project = var.kyma_project_gcp_project_id
  role = "roles/logging.viewer"
  member = "group:${var.kyma_developer_admins_email}"
}