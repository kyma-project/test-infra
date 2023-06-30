# Create the terraform executor Google Cloud and k8s service accounts.
# k8s service accounts are created in the prow workloads clusters.
# GCP and k8s service account are bind together with workload identity.
# It grants owner rights to the Google Cloud service account. The owner role is required to let
# the terraform executor manage all the resources in the Google Cloud project.
# It also grants the terraform executor gcp service account the owner role in the workloads project.

# Create workload identity principal name.
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

# Grant owner role to terraform executor service account in the workloads project.
resource "google_project_iam_member" "terraform_executor_owner" {
  project = var.workloads_project_id
  role    = "roles/owner"
  member  = "serviceAccount:${google_service_account.terraform_executor.email}"
}

resource "google_service_account_iam_binding" "terraform_workload_identity" {
  members            = ["serviceAccount:${local.terraform_workload_identity_gcp_service_account}"]
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = google_service_account.terraform_executor.name
}

resource "kubernetes_service_account" "trusted_workload_terraform_executor" {
  provider = kubernetes.trusted_workload_k8s_cluster

  metadata {
    namespace = var.terraform_executor_k8s_service_account.namespace
    name      = var.terraform_executor_k8s_service_account.name
    annotations = {
      "iam.gke.io/gcp-service-account" = format("%s@%s.iam.gserviceaccount.com", var
      .terraform_executor_gcp_service_account.id, var.terraform_executor_gcp_service_account.project_id)
    }
  }
  automount_service_account_token = true
}

resource "kubernetes_service_account" "untrusted_workload_terraform_executor" {
  provider = kubernetes.untrusted_workload_k8s_cluster

  metadata {
    namespace = var.terraform_executor_k8s_service_account.namespace
    name      = var.terraform_executor_k8s_service_account.name
    annotations = {
      "iam.gke.io/gcp-service-account" = format("%s@%s.iam.gserviceaccount.com", var
      .terraform_executor_gcp_service_account.id, var.terraform_executor_gcp_service_account.project_id)
    }
  }
  automount_service_account_token = true
}
