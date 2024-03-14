# Allow mapping the permisisons to workflows running gitleaks action via workload identity federation
# it's primary use to retrieve gitleaks related secrets from GCP secret manager.
resource "google_service_account" "gitleaks_secret_accesor" {
  project      = var.gitleaks_gcp_service_account.project_id
  account_id   = var.gitleaks_gcp_service_account.id
  display_name = var.gitleaks_gcp_service_account.id
  description  = "Identity of gitleaks. It's used to retrieve secrets from secret manager"
}

# Retrieve id of kyma project github organization used in subject of workload identity federation.
data "github_organization" "kyma-project" {
  provider = github.kyma_project
  name     = "kyma-project"
}

# Binds gitleaks service account with associated workload identity federation subject, allowing all workflows under kyma-project on github to use that service account.
resource "google_service_account_iam_binding" "gitleaks_workload_identity_federation_binding" {
  members = [
    "principal://iam.googleapis.com/projects/${data.google_client_config.gcp.project}/locations/global/workloadIdentityPools/${module.gh_com_kyma_project_workload_identity_federation.pool_name}/subject/repository_owner_id:${data.github_organization.kyma-project.id}:workflow:${var.gitleaks_workflow_name}"
  ]
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = google_service_account.gitleaks_secret_accesor.name
}

# Github organizational wide variables used in gitleaks workflows on kyma-project organization
# It holds the email address of gitleaks service account
resource "github_actions_organization_variable" "github_gitleaks_secret_accesor_service_account_email" {
  provider      = github.kyma_project
  variable_name = "GCP_GITLEAKS_SECRET_ACCESOR_SERVICE_ACCOUNT_EMAIL"
  visibility    = "all"
  value         = google_service_account.gitleaks_secret_accesor.email
}

# Github organizational wide variables used in gitleaks workflows on kyma-project organization
# It holds the name of the secret containing gitleaks license key created by neighbors team and required by the gitleaks action
resource "github_actions_organization_variable" "github_gitleaks_license_secret_name" {
  provider      = github.kyma_project
  variable_name = "GH_GITLEAKS_LICENSE_SECRET_NAME"
  visibility    = "all"
  value         = "gitleaks-kyma-license-key"
}
