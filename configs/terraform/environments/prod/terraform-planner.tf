# Create the terraform planner GCP servcie account.
# Grants the browser permissions to refresh state of the resources.

resource "google_service_account" "terraform_planner" {
  project      = var.terraform_planner_gcp_service_account.project_id
  account_id   = var.terraform_planner_gcp_service_account.id
  display_name = var.terraform_planner_gcp_service_account.id

  description = "Identity of terraform planner"
}

# Grant browser role to terraform planner service account
resource "google_project_iam_member" "terraform_planner_prow_project_browser" {
  project = var.terraform_planner_gcp_service_account.project_id
  role    = "roles/browser"
  member  = "serviceAccount:${google_service_account.terraform_planner.email}"
}

resource "google_service_account_iam_binding" "terraform_planner_workload_identity" {
  members = [
    "principal://iam.googleapis.com/projects/351981214969/locations/global/workloadIdentityPools/github-com-kyma-project/subject/repository_id:147495537:repository_owner_id:39153523:workflow:Pull Plan Prod Terraform"
  ]
  role               = "roles/workloadIdentityUser"
  service_account_id = google_service_account.terraform_planner.name
}

resource "google_project_iam_member" "terraform_planner_workloads_project_browser" {
  project = var.workloads_project_id
  role    = "roles/browser"
  member  = "serviceAccount:${google_service_account.terraform_planner.email}"
}
