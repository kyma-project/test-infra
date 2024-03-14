# Gives permissions to gitleaks to read secret with license key
resource "google_service_account" "gitleaks_secret_accesor" {
  project      = var.gitleaks_gcp_service_account.project_id
  account_id   = var.gitleaks_gcp_service_account.id
  display_name = var.gitleaks_gcp_service_account.id
  description  = "Identity of gitleaks. It's used to retrieve secrets from secret manager"
}

resource "google_service_account_iam_binding" "gitleaks_workload_identity_federation_binding" {
  members = [
    "principal://iam.googleapis.com/projects/${data.google_client_config.gcp.project}/locations/global/workloadIdentityPools/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.gitleaks_workflow_name}"
  ]
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = google_service_account.gitleaks_secret_accesor.name
}

resource "github_actions_organization_variable" "github_gitleaks_secret_accesor_service_account_email" {
  provider      = github.kyma_project
  variable_name = "GCP_GITLEAKS_SECRET_ACCESOR_SERVICE_ACCOUNT_EMAIL"
  visibility    = "all"
  value         = google_service_account.gitleaks_secret_accesor.email
}

resource "github_actions_organization_variable" "github_gitleaks_license_secret_name" {
  provider      = github.kyma_project
  variable_name = "GH_GITLEAKS_LICENSE_SECRET_NAME"
  visibility    = "all"
  value         = "gitleaks-kyma-license-key"
}
