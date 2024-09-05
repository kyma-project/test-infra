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