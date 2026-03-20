variable "terraform_executor_gcp_service_account" {
  type = object({
    id         = string
    project_id = string
  })

  default = {
    id         = "terraform-executor"
    project_id = "sap-kyma-prow"
  }

  description = "Details of terraform executor gcp service account."
}

variable "terraform_planner_gcp_service_account" {
  type = object({
    id         = string
    project_id = string
  })

  default = {
    id         = "terraform-planner"
    project_id = "sap-kyma-prow"
  }

  description = "Details of terraform planner gcp service account"
}

variable "github_terraform_apply_workflow_name" {
  type        = string
  default     = "Post Apply Prod Terraform"
  description = "Workflow name for terraform apply workflow"
}

variable "github_terraform_plan_workflow_name" {
  type        = string
  default     = "Pull Plan Prod Terraform"
  description = "Workflow name for terraform plan workflow"
}

# ------------------------------------------------------------------------------
# Internal GitHub Enterprise WIF Binding Variables
# ------------------------------------------------------------------------------
# Variables for configuring WIF IAM bindings that allow internal GitHub
# workflows to impersonate the terraform planner and executor service accounts.
# ------------------------------------------------------------------------------

variable "internal_github_terraform_plan_reusable_workflow_ref" {
  type        = string
  default     = "kyma/test-infra/.github/workflows/iac-plan.yml@refs/heads/main"
  description = "Value of the GitHub OIDC job_workflow_ref claim for the terraform plan workflow on internal GitHub. Used to match the reusable_workflow_ref attribute in the github-tools-sap WIF pool."
}

variable "internal_github_terraform_deploy_identity" {
  type        = string
  default     = "kyma/test-infra/.github/workflows/iac-deploy.yml:main"
  description = "Value of the deploy_identity attribute for the terraform deploy workflow on internal GitHub. The deploy_identity attribute is derived from job_workflow_ref and only populated when the caller ref is refs/heads/main or a v-tag. Format: org/repo/.github/workflows/workflow.yml:main"
}
