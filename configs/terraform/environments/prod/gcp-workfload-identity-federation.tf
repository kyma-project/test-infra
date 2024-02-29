data "github_repository" "test_infra" {
  provider = github.kyma_project
  name     = var.github_test_infra_repository_name
}

module "gh_com_kyma_project_workload_identity_federation" {
  source = "../../modules/gcp-workload-identity-federation"

  project_id  = data.google_client_config.gcp.project
  pool_id     = "github-com-kyma-project"
  provider_id = "github-com-kyma-project"
  issuer_uri  = "https://token.actions.githubusercontent.com"

  attribute_mapping = {
    "google.subject"                = "\"repository_id:\" + assertion.repository_id + \":repository_owner_id:\" + assertion.repository_owner_id + \":workflow:\" + assertion.workflow"
    "attribute.actor"               = "assertion.actor"
    "attribute.aud"                 = "assertion.aud"
    "attribute.repository_id"       = "assertion.repository_id"
    "attribute.repository_owner_id" = "assertion.repository_owner_id"
    "attribute.workflow"            = "assertion.workflow"
    "attribute.workflow_ref"        = "assertion.workflow_ref"
    "attribute.event_name"          = "assertion.event_name"
  }

  sa_mapping = {
    "terraform_planner_pull_prod_plan" = {
      sa_name   = "projects/${data.google_client_config.gcp.project}/serviceAccounts/${google_service_account.terraform_planner.email}"
      attribute = "subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.github_terraform_plan_workflow_name}"
    },
    "terraform_executor_post_prod_apply" = {
      sa_name   = "projects/${data.google_client_config.gcp.project}/serviceAccounts/${google_service_account.terraform_executor.email}"
      attribute = "subject/repository_id:${data.github_repository.test_infra.repo_id}:repository_owner_id:${var.github_kyma_project_organization_id}:workflow:${var.github_terraform_apply_workflow_name}"
    }
  }
}

resource "github_actions_variable" "gcp_terraform_executor_service_account_email" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GCP_TERRAFORM_EXECUTOR_SERVICE_ACCOUNT_EMAIL"
  value         = google_service_account.terraform_executor.email
}

resource "github_actions_variable" "gcp_terraform_planner_service_account_email" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GCP_TERRAFORM_PLANNER_SERVICE_ACCOUNT_EMAIL"
  value         = google_service_account.terraform_planner.email
}

resource "github_actions_variable" "gh_com_kyma_project_gcp_workload_identity_federation_provider" {
  provider      = github.kyma_project
  repository    = "test-infra"
  variable_name = "GH_COM_KYMA_PROJECT_GCP_WORKLOAD_IDENTITY_FEDERATION_PROVIDER"
  value         = module.gh_com_kyma_project_workload_identity_federation.provider_name
}

