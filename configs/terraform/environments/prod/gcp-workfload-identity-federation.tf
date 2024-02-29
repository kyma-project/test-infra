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
    "terraform_executor_pull_prod_plan" = {
      sa_name   = "projects/${data.google_client_config.gcp.project}/serviceAccounts/${google_service_account.terraform_executor.email}"
      attribute = "subject/repository_id:147495537:repository_owner_id:39153523:workflow:Pull Plan Prod Terraform"
    },
    "terraform_executor_post_prod_apply" = {
      sa_name   = "projects/${data.google_client_config.gcp.project}/serviceAccounts/${google_service_account.terraform_executor.email}"
      attribute = "subject/repository_id:147495537:repository_owner_id:39153523:workflow:Post Apply Prod Terraform"
    }
  }
}

