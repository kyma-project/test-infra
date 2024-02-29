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

variable "github_terraform_plan_workflow_name" {
  type        = string
  default     = "Pull Plan Prod Terraform"
  description = "Workflow name for terraform plan workflow"
}

variable "github_terraform_apply_workflow_name" {
  type        = string
  default     = "Post Apply Prod Terraform"
  description = "Workflow name for terraform apply workflow"
}
