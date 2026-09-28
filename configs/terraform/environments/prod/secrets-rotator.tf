# -----------------------------------------------------------------------------
# Secrets Rotator — migration to tooling-infra
# -----------------------------------------------------------------------------
# The secrets-rotator applications (signify-secret-rotator, service-account-keys
# rotator + cleaner) and their shared infrastructure have been migrated to the
# internal github.tools.sap/kyma/tooling-infra central root, which manages the
# sap-kyma-prow project going forward.
#
# Removal strategy (single-owner rule — a resource is managed by exactly one
# Terraform state at a time):
#
#   * Resources that hold non-recoverable data or have many dependents are
#     RELEASED FROM STATE ONLY here (removed { destroy = false }), so they stay
#     alive and are then IMPORTED by tooling-infra:
#       - google_pubsub_topic.secrets_rotator_dead_letter
#       - google_pubsub_subscription.secrets-rotator-dead-letter
#       - google_storage_bucket.secret-rotator-dead-letters-bucket
#         (force_destroy = true, holds archived dead-letter messages)
#
#   * Recreatable resources (config only, no data) are simply deleted from the
#     configuration so Terraform DESTROYS them; tooling-infra recreates them:
#       - google_service_account.secrets-rotator (orchestration SA — dropped;
#         each migrated app module now owns its own service account)
#       - google_storage_bucket_iam_member.dead-letter-bucket-access
#       - google_monitoring_alert_policy.dead-letter-alert
#       - module.service_account_keys_rotator
#       - module.service_account_keys_cleaner
#       - module.signify_secret_rotator
#
# The kyma-signify-prod secret is NOT touched — it is owned by the image-builder
# service (see image-builder.tf) and stays in this state.
#
# These removed{} blocks are safe to delete after the first successful apply.
# -----------------------------------------------------------------------------

removed {
  from = google_pubsub_topic.secrets_rotator_dead_letter
  lifecycle {
    destroy = false
  }
}

removed {
  from = google_pubsub_subscription.secrets-rotator-dead-letter
  lifecycle {
    destroy = false
  }
}

removed {
  from = google_storage_bucket.secret-rotator-dead-letters-bucket
  lifecycle {
    destroy = false
  }
}
