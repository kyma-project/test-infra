# This file creates the terraform executor Google Cloud service account and k8s service accounts need to run terraform.
# k8s service accounts are created in the prow workloads and tekton clusters.
# It grants required permissions to the Google Cloud service account and setup workload identity.
# It also grants the terraform executor service account the owner role in the workloads project.

module "terraform_executor_gcp_service_account" {
  source = "../../../../development/terraform-executor/terraform/modules/gcp-terraform-executor"

  terraform_executor_gcp_service_account = {
    id         = var.terraform_executor_gcp_service_account.id
    project_id = var.terraform_executor_gcp_service_account.project_id
  }

  terraform_executor_k8s_service_account = {
    name      = var.terraform_executor_k8s_service_account.name,
    namespace = var.terraform_executor_k8s_service_account.namespace
  }
}

# Grant owner role to terraform executor service account in the workloads project. The owner role is required to let
# the terraform executor manage all the resources in the Google Cloud project.
resource "google_project_iam_member" "terraform_executor_owner" {
  project = var.workloads_project_id
  role    = "roles/owner"
  member  = "serviceAccount:${module.terraform_executor_gcp_service_account.terraform_executor_gcp_service_account.email}"
}

module "trusted_workload_terraform_executor_k8s_service_account" {
  providers = {
    google     = google
    kubernetes = kubernetes.trusted_workload_k8s_cluster
  }
  source = "../../../../development/terraform-executor/terraform/modules/k8s-terraform-executor"

  terraform_executor_gcp_service_account = {
    id         = var.terraform_executor_gcp_service_account.id
    project_id = var.terraform_executor_gcp_service_account.project_id
  }
  terraform_executor_k8s_service_account = {
    name      = var.terraform_executor_k8s_service_account.name,
    namespace = var.terraform_executor_k8s_service_account.namespace
  }

  external_secrets_sa = {
    name      = var.external_secrets_sa_trusted_cluster.name,
    namespace = var.external_secrets_sa_trusted_cluster.namespace
  }

}

module "untrusted_workload_terraform_executor_k8s_service_account" {
  providers = {
    google     = google
    kubernetes = kubernetes.untrusted_workload_k8s_cluster
  }
  source = "../../../../development/terraform-executor/terraform/modules/k8s-terraform-executor"

  terraform_executor_gcp_service_account = {
    id         = var.terraform_executor_gcp_service_account.id
    project_id = var.terraform_executor_gcp_service_account.project_id
  }
  terraform_executor_k8s_service_account = {
    name      = var.terraform_executor_k8s_service_account.name,
    namespace = var.terraform_executor_k8s_service_account.namespace
  }

  external_secrets_sa = {
    name      = var.external_secrets_sa_untrusted_cluster.name,
    namespace = var.external_secrets_sa_untrusted_cluster.namespace
  }

}
