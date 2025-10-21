resource "google_service_account" "kyma-oci-image-builder" {
  account_id   = "kyma-oci-image-builder"
  display_name = "kyma-oci-image-builder"
  description  = "Service Account for retrieving secrets on the oci-image-builder ADO pipeline."

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_service_account" "terraform-planner" {
  account_id   = "terraform-planner"
  display_name = "terraform-planner"
  description  = "Identity of terraform planner"

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_service_account" "terraform-executor" {
  account_id   = "terraform-executor"
  display_name = "terraform-executor"
  description  = "Identity of terraform executor. It's mapped to k8s service account through workload identity."

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_service_account" "sa-kyma-project" {
  account_id   = "sa-kyma-project"
  display_name = "sa-kyma-project"
  description  = "SA to manage PROD Artifact Registry in SAP CX Kyma Project"

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_service_account" "gitleaks-secret-accesor" {
  account_id   = "gitleaks-secret-accesor"
  display_name = "gitleaks-secret-accesor"
  description  = "Identity of gitleaks. It's used to retrieve secrets from secret manager"
}

resource "google_service_account" "sa-secret-update" {
  account_id   = "sa-secret-update"
  display_name = "sa-secret-update"
  description  = "Can update secrets in Secret Manager"
}

resource "google_service_account" "sa-security-dashboard-oauth" {
  account_id   = "sa-security-dashboard-oauth"
  display_name = "sa-security-dashboard-oauth"
  description  = "Used for the Security dashboard cloud run"

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_service_account" "sa-dev-kyma-project" {
  account_id   = "sa-dev-kyma-project"
  display_name = "sa-dev-kyma-project"
  description  = "SA to manage DEV Artifact Registry in SAP CX Kyma Project"

  lifecycle {
    prevent_destroy = true
  }
}


resource "google_service_account" "kyma-security-scanners" {
  account_id   = "kyma-security-scanners"
  display_name = "kyma-security-scanners"
  description  = "Service account for retrieving secrets on the security-scanners and orphan-cleaner Azure pipelines."

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_service_account" "kyma-compliance-pipeline" {
  account_id   = "kyma-compliance-pipeline"
  display_name = "kyma-compliance-pipeline"
  description  = "Service account for retrieving secrets on the compliance Azure pipeline."

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_service_account" "neighbors-conduit-cli-builder" {
  account_id   = "neighbors-conduit-cli-builder"
  display_name = "neighbors-conduit-cli-builder"
  description  = "Service account for retrieving secrets on the conduit-cli build pipeline."
}