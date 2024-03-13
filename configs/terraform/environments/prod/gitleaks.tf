# Gives permissions to gitleaks to read secret with license key
resource "google_service_account" "gitleaks_secret_accesor" {
  project      = var.gitleaks_gcp_service_account.project_id
  account_id   = var.gitleaks_gcp_service_account.id
  display_name = var.gitleaks_gcp_service_account.id
  description  = "Identity of gitleaks. It's used to retrieve secrets from secret manager"
}

data "github_repository" "gitleaks_repository" {
  for_each = var.gitleaks_repositories
  provider = github.kyma_project
  name     = each.value
}

resource "google_service_account_iam_binding" "gitleaks_workload_identity_federation_binding" {
  for_each = data.github_repository.gitleaks_repository
  members = [
    "principal://iam.googleapis.com/projects/${data.google_client_config.gcp.project}/locations/global/workloadIdentityPools/${module.gh_com_kyma_project_workload_identity_federation.provider_name}/subject/repository_id:${each.value.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.gitleaks_workflow_name}"
  ]
  role               = "roles/iam.workloadIdentityUser"
  service_account_id = google_service_account.gitleaks_secret_accesor.name
}
