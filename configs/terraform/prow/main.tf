terraform {
  backend "gcs" {
    bucket = "tf-state-kyma-project"
    prefix = "prow"
  }
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.50.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.17.0"
    }
  }
}

variable "gcp_zone" {
  type        = string
  description = ""
}

variable "gcp_region" {
  type        = string
  description = ""
}

variable "gcp_project_id" {
  type        = string
  description = ""
}

variable "k8s_config_path" {
  type        = string
  description = ""
}

variable "k8s_config_context" {
  type        = string
  description = ""
}

variable "prow_terraform_executor_gcp_service_account" {
  type = object({
    id = string
  })
  description = ""
}

provider "kubernetes" {
  config_path    = var.k8s_config_path
  config_context = var.k8s_config_context
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
  zone    = var.gcp_zone
}

locals {
  #prow_terraform_workload_identity_gcp_service_account = format("%s.svc.id.goog[%s/%s]", var.gcp_project_id, var.k8s_terraform_sa.namespace, var.k8s_terraform_sa.name)
  prow_terraform_workload_identity_gcp_service_account = format("%s.svc.id.goog[%s/%s]", var.gcp_project_id, kubernetes_service_account.terraform.metadata[0].namespace, kubernetes_service_account.terraform.metadata[0].name)
}

resource "google_service_account" "prow-terraform-executor" {
  account_id   = var.prow_terraform_executor_gcp_service_account.id
  display_name = var.prow_terraform_executor_gcp_service_account.id
  description  = "Identity of terraform executor running on Prow. It's mapped to k8s service account through workload identity."
}

resource "google_project_iam_member" "project_editor" {
  project = var.gcp_project_id
  role    = "roles/editor"
  member  = "serviceAccount:${google_service_account.prow-terraform-executor.email}"
}

#resource "google_service_account_iam_member" "prow_terraform_workload_identity" {
#  member             = "serviceAccount:${local.prow_terraform_workload_identity_gcp_service_account}"
#  role               = "roles/iam.workloadIdentityUser"
#  service_account_id = google_service_account.prow-terraform-executor.name
#  depends_on = [kubernetes_service_account.terraform]
#}

#data "google_iam_policy" "prow_terraform_workload_identity" {
#  binding {
#    members = ["serviceAccount:${local.prow_terraform_workload_identity_gcp_service_account}"]
#    role    = "roles/iam.workloadIdentityUser"
#  }
#}

resource "google_service_account_iam_binding" "prow_terraform_workload_identity" {
  members            = ["serviceAccount:${local.prow_terraform_workload_identity_gcp_service_account}"]
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = google_service_account.prow-terraform-executor.name
}
