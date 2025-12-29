variable "sec-scanner-cfg-processor-service-account" {
  type = object({
    service-account-name        = string
    service-account-description = string
  })
  default = {
    service-account-name        = "sec-scanner-cfg-processor"
    service-account-description = "Identity of sec-scanner-cfg-processor"
  }
}

resource "google_service_account" "sec-scanner-cfg-processor" {
  account_id   = var.sec-scanner-cfg-processor-service-account.service-account-name
  display_name = var.sec-scanner-cfg-processor-service-account.service-account-name
  description  = var.sec-scanner-cfg-processor-service-account.service-account-description
}

resource "google_artifact_registry_repository_iam_member" "sec_scanner_cfg_processor_kyma_modules_reader" {
  project    = module.kyma_modules.artifact_registry.project
  repository = module.kyma_modules.artifact_registry.name
  location   = module.kyma_modules.artifact_registry.location
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.sec-scanner-cfg-processor.email}"
}

resource "google_secret_manager_secret" "sec-scanner-cfg-processor-gcp-sa-key" {
  secret_id = "sec-scanner-cfg-gcp-sa-key"

  # Enable automatic replication across Google-managed regions.
  replication {
    auto {}
  }

  # Configure rotation to occur every 85 days (85 * 24 * 3600 = 7,344,000 seconds).
  rotation {
    rotation_period    = "7344000s"
    # Next rotation time must be set when rotation period is specified.
    # Set it to now + rotation period. We ignore future drifts via lifecycle below.
    next_rotation_time = timeadd(timestamp(), "7344000s")
  }

  # Publish Secret Manager events (create, delete, etc.) to the given Pub/Sub topic.
  topics {
    name = "projects/sap-kyma-prow/topics/secret-manager-notifications"
  }

  labels = {
    type = "service-account-key"
  }

  # Avoid perpetual diffs when next_rotation_time advances automatically on Google side
  lifecycle {
    ignore_changes = [
      rotation[0].next_rotation_time,
    ]
  }
}

# Grant read access to the secret for kyma-security-scanners service account.
resource "google_secret_manager_secret_iam_member" "sec_scanner_cfg_processor_secret_accessor" {
  project   = data.google_client_config.gcp.project
  secret_id = google_secret_manager_secret.sec-scanner-cfg-processor-gcp-sa-key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.kyma-security-scanners.email}"
}