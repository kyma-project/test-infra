variable "github_test_infra_repository_name" {
  type        = string
  default     = "test-infra"
  description = "Name of the test-infra repository"
}

variable "github_kyma_project_organization_id" {
  type        = string
  default     = "39153523"
  description = "kyma-project organziaiton id"
}

# variable "github_terraform_plan_workflow_name" {
#   type        = string
#   default     = "Pull Plan Prod Terraform"
#   description = "Workflow name for terraform plan workflow"
# }

# variable "github_terraform_apply_workflow_name" {
#   type        = string
#   default     = "Post Apply Prod Terraform"
#   description = "Workflow name for terraform apply workflow"
# }

variable "gh_com_kyma_project_wif_pool_id" {
  type        = string
  default     = "github-com-kyma-project"
  description = "Google Cloud Platform workflow identity federation pool id used for github.com/kyma-project org identities"
}

variable "gh_com_kyma_project_wif_provider_id" {
  type        = string
  default     = "github-com-kyma-project"
  description = "Google Cloud Platform workflow identity federation provider id used for github.com/kyma-project org identities"
}

variable "gh_com_kyma_project_wif_issuer_uri" {
  type        = string
  default     = "https://token.actions.githubusercontent.com"
  description = "GitHub OIDC provider issuer URI, this URI is used to validated a token signature when authenticating using Workload Identity Federation."
}