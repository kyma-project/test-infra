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
  }

  attribute_condition = var.gh_com_kyma_project_wif_attribute_condition

  #   sa_mapping = {
  #     "terraform_planner_pull_prod_plan" = {
  #       sa_name   = "projects/${data.google_client_config.gcp.project}/serviceAccounts/${google_service_account.terraform_planner.email}"
  #       attribute = "subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.github_terraform_plan_workflow_name}"
  #     },
  #     "terraform_executor_post_prod_apply" = {
  #       sa_name   = "projects/${data.google_client_config.gcp.project}/serviceAccounts/${google_service_account.terraform_executor.email}"
  #       attribute = "subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.github_terraform_apply_workflow_name}"
  #     }
  #    }
}

# TODO(dekiel): Another GitHub variables related to workload identity federation are defined in github-com.tf file.
# resource "github_actions_variable" "gcp_terraform_executor_service_account_email" {
#   provider      = github.kyma_project
#   repository    = "test-infra"
#   variable_name = "GCP_TERRAFORM_EXECUTOR_SERVICE_ACCOUNT_EMAIL"
#   value         = google_service_account.terraform_executor.email
# }
#
# resource "github_actions_variable" "gcp_terraform_planner_service_account_email" {
#   provider      = github.kyma_project
#   repository    = "test-infra"
#   variable_name = "GCP_TERRAFORM_PLANNER_SERVICE_ACCOUNT_EMAIL"
#   value         = google_service_account.terraform_planner.email
# }

resource "github_actions_organization_variable" "gh_com_kyma_project_gcp_workload_identity_federation_provider" {
  provider      = github.kyma_project
  visibility    = "all"
  variable_name = "GH_COM_KYMA_PROJECT_GCP_WORKLOAD_IDENTITY_FEDERATION_PROVIDER"
  value         = module.gh_com_kyma_project_workload_identity_federation.provider_name
}