resource "google_iam_workload_identity_pool" "main" {
  project                   = var.project_id
  workload_identity_pool_id = var.pool_id
  disabled                  = false
}

# TODO (dekiel): Add setting attributeCondition value. https://cloud.google.com/iam/docs/reference/rest/v1/projects.locations.workloadIdentityPools.providers
#  The attributeCondition let us control which external identities issued by provider are allowed to use the pool.
resource "google_iam_workload_identity_pool_provider" "main" {
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.main.workload_identity_pool_id
  workload_identity_pool_provider_id = var.provider_id
  disabled                           = false
  attribute_mapping                  = var.attribute_mapping
  attribute_condition = var.attribute_condition

  oidc {
    issuer_uri        = var.issuer_uri
    allowed_audiences = var.allowed_audiences
  }
}

resource "google_service_account_iam_member" "service_account" {
  for_each           = var.sa_mapping
  service_account_id = each.value.sa_name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principal://iam.googleapis.com/${google_iam_workload_identity_pool.main.name}/${each.value.attribute}"
}