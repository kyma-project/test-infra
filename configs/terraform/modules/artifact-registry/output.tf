output "artifact_registry" {
  description = "Artifact Registry"
  # Avoid update_time and create_time output contribution
  # It shoudl work according to:
  # See: https://github.com/hashicorp/terraform/issues/28803#issuecomment-1072740861
  # Fixes: https://github.tools.sap/kyma/test-infra/issues/945
  value = merge(local.repository, {
    update_time = null
    create_time = null
  })
}
