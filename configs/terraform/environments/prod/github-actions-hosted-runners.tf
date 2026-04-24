# ------------------------------------------------------------------------------
# GitHub Actions Hosted Runners for telemetry-manager
# ------------------------------------------------------------------------------
# This configuration provisions a dedicated runner group and GitHub-hosted
# larger runner for the telemetry-manager repository's CI workloads.
#
# The runner group is scoped exclusively to the telemetry-manager repository
# via visibility = "selected", ensuring isolation from other org repos.
# This addresses CI flakiness caused by shared runner pool contention during
# business hours (see: https://github.tools.sap/kyma/test-infra/issues/1308).
# ------------------------------------------------------------------------------

data "github_repository" "telemetry_manager" {
  provider = github.kyma_project
  name     = "telemetry-manager"
}

resource "github_actions_runner_group" "telemetry_manager" {
  provider                   = github.kyma_project
  name                       = var.telemetry_manager_runner_group_name
  visibility                 = "selected"
  allows_public_repositories = true

  selected_repository_ids = [
    data.github_repository.telemetry_manager.repo_id,
  ]

  restricted_to_workflows = true
  selected_workflows = [
    "kyma-project/telemetry-manager/.github/workflows/pr-integration.yml@refs/heads/main",
  ]
}

resource "github_actions_hosted_runner" "telemetry_manager" {
  provider = github.kyma_project

  name = var.telemetry_manager_hosted_runner.name

  image {
    id     = var.telemetry_manager_hosted_runner.image_id
    source = var.telemetry_manager_hosted_runner.image_source
  }

  size            = var.telemetry_manager_hosted_runner.size
  runner_group_id = github_actions_runner_group.telemetry_manager.id
  maximum_runners = var.telemetry_manager_hosted_runner.max_runners
}
