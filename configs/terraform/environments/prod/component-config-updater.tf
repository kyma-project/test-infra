# ==============================================================================
# component-config-updater workflow — Secret Manager access via WIF (github.com)
# ==============================================================================
# Grants the kyma-project/test-infra reusable-component-config-updater.yml
# workflow (running on github.com) read access to two PAT secrets via the
# external github.com Workload Identity Federation pool:
#
#   - kyma-prow bot PAT (opens PRs)            → GH_TOOLS_KYMA_PROW_BOT_TOKEN_SECRET_NAME
#   - kyma bot auto-approver PAT (approves PRs) → GH_COM_KYMA_BOT_AUTO_APPROVER_TOKEN_SECRET_NAME
#
# Two distinct identities are required because GitHub forbids self-approval.
#
# The principal is scoped to the reusable_workflow_ref of the workflow living
# in kyma-project/test-infra@main on github.com, using the existing
# gh_com_kyma_project WIF pool (pool id: github-com-kyma-project).
# Mirrors the internal pattern in kyma/tooling-infra:component-config-updater.tf.
# ==============================================================================

variable "com_component_config_updater_bumper_token_secret_id" {
  description = "GCP Secret Manager secret ID for the kyma-prow bot PAT used by the github.com reusable-component-config-updater workflow to open PRs. Must match the kyma-project org variable GH_TOOLS_KYMA_PROW_BOT_TOKEN_SECRET_NAME on github.com. Empty disables the IAM binding."
  type        = string
  default     = ""
}

variable "com_component_config_updater_approver_token_secret_id" {
  description = "GCP Secret Manager secret ID for the kyma bot auto-approver PAT used by the github.com reusable-component-config-updater workflow to auto-approve PRs. Must match the kyma-project org variable GH_COM_KYMA_BOT_AUTO_APPROVER_TOKEN_SECRET_NAME on github.com. Empty disables the IAM binding."
  type        = string
  default     = ""
}

variable "com_component_config_updater_reusable_workflow_ref" {
  description = "GitHub reference for the reusable workflow on github.com used by component-config-updater callers. Used to scope the principalSet binding."
  type        = string
  default     = "kyma-project/test-infra/.github/workflows/reusable-component-config-updater.yml@refs/heads/main"
}

locals {
  com_component_config_updater_supported_events = [
    "push",
    "schedule",
    "workflow_dispatch",
  ]

  # WIF pool full resource name for the external github.com pool.
  com_github_wif_pool_name = "projects/${data.google_project.current.number}/locations/global/workloadIdentityPools/${var.gh_com_kyma_project_wif_pool_id}"

  com_component_config_updater_principals = {
    for event in local.com_component_config_updater_supported_events :
    event => "principalSet://iam.googleapis.com/${local.com_github_wif_pool_name}/attribute.reusable_workflow_run/event_name:${event}:repository_owner_id:${var.github_kyma_project_organization_id}:reusable_workflow_ref:${var.com_component_config_updater_reusable_workflow_ref}"
  }
}

resource "google_secret_manager_secret_iam_member" "com_component_config_updater_bumper_token_reader" {
  for_each = var.com_component_config_updater_bumper_token_secret_id != "" ? toset(local.com_component_config_updater_supported_events) : toset([])

  project   = data.google_project.current.project_id
  secret_id = var.com_component_config_updater_bumper_token_secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = local.com_component_config_updater_principals[each.value]
}

resource "google_secret_manager_secret_iam_member" "com_component_config_updater_approver_token_reader" {
  for_each = var.com_component_config_updater_approver_token_secret_id != "" ? toset(local.com_component_config_updater_supported_events) : toset([])

  project   = data.google_project.current.project_id
  secret_id = var.com_component_config_updater_approver_token_secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = local.com_component_config_updater_principals[each.value]
}
