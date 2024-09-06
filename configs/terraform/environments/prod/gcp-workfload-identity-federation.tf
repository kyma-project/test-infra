data "github_repository" "test_infra" {
  provider = github.kyma_project
  name     = var.github_test_infra_repository_name
}

module "gh_com_kyma_project_workload_identity_federation" {
  source = "../../modules/gcp-workload-identity-federation"

  project_id  = data.google_client_config.gcp.project
  pool_id     = var.gh_com_kyma_project_wif_pool_id
  provider_id = var.gh_com_kyma_project_wif_provider_id
  issuer_uri  = var.gh_com_kyma_project_wif_issuer_uri
  # attribute_condition = var.gh_com_kyma_project_wif_attribute_condition

  attribute_mapping = {
    "google.subject"                  = "\"repository_id:\" + assertion.repository_id + \":repository_owner_id:\" + assertion.repository_owner_id + \":workflow:\" + assertion.workflow"
    "attribute.actor"                 = "assertion.actor"
    "attribute.aud"                   = "assertion.aud"
    "attribute.repository_id"         = "assertion.repository_id"
    "attribute.repository_owner_id"   = "assertion.repository_owner_id"
    "attribute.workflow"              = "assertion.workflow"
    "attribute.workflow_ref"          = "assertion.workflow_ref"
    "attribute.event_name"            = "assertion.event_name"
    "attribute.reusable_workflow_ref" = "assertion.job_workflow_ref"
    "attribute.reusable_workflow_sha" = "assertion.job_workflow_sha"
    # The reusable_workflow_run attribute is used to identify the run of github workflow that is calling a reusable workflow.
    # It allows to distinguish between runs for pull requests and push events. Usually workflows need different permissions for these events.
    # The repository_owner_id is used to prevent granting access to the oidc tokens issued for other organisations existing in github.com.
    # The reusable workflows should have step to check if the caller owner is allowed to use the reusable workflow.
    # The repository_owner_id is used here as second line of defense and validation at the infrastructure level.
    "attribute.reusable_workflow_run" = "\"event_name:\" + assertion.event_name + \":repository_owner_id:\" + assertion.repository_owner_id + \":reusable_workflow_ref:\" + assertion.job_workflow_ref"
  }
}

resource "github_actions_organization_variable" "gh_com_kyma_project_gcp_workload_identity_federation_provider" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "GH_COM_KYMA_PROJECT_GCP_WORKLOAD_IDENTITY_FEDERATION_PROVIDER"
  value         = module.gh_com_kyma_project_workload_identity_federation.provider_name
}