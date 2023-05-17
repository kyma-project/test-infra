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

# Grant owner role to terraform executor service account in the workloads project.
resource "google_project_iam_member" "terraform_executor_owner" {
  project = var.workloads_project_id
  role    = "roles/owner"
  member  = "serviceAccount:${module.terraform_executor_gcp_service_account.terraform_executor_gcp_service_account.email}"
}

module "tekton_terraform_executor_k8s_service_account" {
  source = "../../../../development/terraform-executor/terraform/modules/k8s-terraform-executor"

  terraform_executor_gcp_service_account = {
    id         = var.terraform_executor_gcp_service_account.id
    project_id = var.terraform_executor_gcp_service_account.project_id
  }
  terraform_executor_k8s_service_account = {
    name      = var.terraform_executor_k8s_service_account.name,
    namespace = var.terraform_executor_k8s_service_account.namespace
  }

  managed_k8s_cluster = var.tekton_k8s_cluster

  gcp_region     = var.gcp_region
  gcp_project_id = var.gcp_project_id
}

module "trusted_workloads_terraform_executor_k8s_service_account" {
  source = "../../../../development/terraform-executor/terraform/modules/k8s-terraform-executor"

  terraform_executor_gcp_service_account = {
    id         = var.terraform_executor_gcp_service_account.id
    project_id = var.terraform_executor_gcp_service_account.project_id
  }
  terraform_executor_k8s_service_account = {
    name      = var.terraform_executor_k8s_service_account.name,
    namespace = var.terraform_executor_k8s_service_account.namespace
  }

  managed_k8s_cluster = var.trusted_workload_k8s_cluster

  gcp_region     = var.gcp_region
  gcp_project_id = var.gcp_project_id
}

module "untrusted_workloads_terraform_executor_k8s_service_account" {
  source = "../../../../development/terraform-executor/terraform/modules/k8s-terraform-executor"

  terraform_executor_gcp_service_account = {
    id         = var.terraform_executor_gcp_service_account.id
    project_id = var.terraform_executor_gcp_service_account.project_id
  }
  terraform_executor_k8s_service_account = {
    name      = var.terraform_executor_k8s_service_account.name,
    namespace = var.terraform_executor_k8s_service_account.namespace
  }

  managed_k8s_cluster = var.untrusted_workload_k8s_cluster

  gcp_region     = var.gcp_region
  gcp_project_id = var.gcp_project_id
}
