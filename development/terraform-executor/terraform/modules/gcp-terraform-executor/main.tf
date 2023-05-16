locals {
  terraform_workload_identity_gcp_service_account = format("%s.svc.id.goog[%s/%s]", var
    .terraform_executor_gcp_service_account.project_id,
  var.terraform_executor_k8s_service_account.namespace, var.terraform_executor_k8s_service_account.name)
}

resource "google_service_account" "terraform_executor" {
  project      = var.terraform_executor_gcp_service_account.project_id
  account_id   = var.terraform_executor_gcp_service_account.id
  display_name = var.terraform_executor_gcp_service_account.id
  description  = "Identity of terraform executor. It's mapped to k8s service account through workload identity."
}

resource "google_project_iam_member" "terraform_executor_owner" {
  project = var.terraform_executor_gcp_service_account.project_id
  role    = "roles/owner"
  member  = "serviceAccount:${google_service_account.terraform_executor.email}"
}

resource "google_service_account_iam_binding" "terraform_workload_identity" {
  members            = ["serviceAccount:${local.terraform_workload_identity_gcp_service_account}"]
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = google_service_account.terraform_executor.name
}
