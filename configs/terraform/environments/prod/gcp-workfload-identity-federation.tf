data "github_repository" "test_infra" {
  provider = github.kyma_project
  name     = var.github_test_infra_repository_name
}

# TODO(dekiel): Repalce pool_id and provider_id with variables.
module "gh_com_kyma_project_workload_identity_federation" {
  source = "../../modules/gcp-workload-identity-federation"

  project_id  = data.google_client_config.gcp.project
  pool_id     = var.gh_com_kyma_project_wif_pool_id
  provider_id = var.gh_com_kyma_project_wif_provider_id
  issuer_uri  = var.gh_com_kyma_project_wif_issuer_uri

  attribute_mapping = {
    "google.subject"                = "\"repository_id:\" + assertion.repository_id + \":repository_owner_id:\" + assertion.repository_owner_id + \":workflow:\" + assertion.workflow"
    "attribute.actor"               = "assertion.actor"
    "attribute.aud"                 = "assertion.aud"
    "attribute.repository_id"       = "assertion.repository_id"
    "attribute.repository_owner_id" = "assertion.repository_owner_id"
    "attribute.workflow"            = "assertion.workflow"
    "attribute.workflow_ref"        = "assertion.workflow_ref"
    "attribute.event_name"          = "assertion.event_name"
    "attribute.reusable_workflow_ref" = "assertion.job_workflow_ref"
    "attribute.reusable_workflow_sha" = "assertion.job_workflow_sha"
  }
}

resource "github_actions_organization_variable" "gh_com_kyma_project_gcp_workload_identity_federation_provider" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "GH_COM_KYMA_PROJECT_GCP_WORKLOAD_IDENTITY_FEDERATION_PROVIDER"
  value         = module.gh_com_kyma_project_workload_identity_federation.provider_name
}