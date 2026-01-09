resource "google_cloud_identity_group" "restricted_registry_prod_read" {
  provider     = google.kyma_project
  display_name = "Kyma Restricted Registry PROD Read Access"
  description  = "Read-only access to Kyma Restricted Registry PROD environment"

  group_key {
    id = "kyma-restricted-registry-prod-read@sap.com"
  }

  labels = {
    "cloudidentity.googleapis.com/groups.security" = ""
  }

  parent = "customers/${data.google_organization.sap.directory_customer_id}"
}

resource "google_cloud_identity_group" "restricted_registry_prod_write" {
  provider     = google.kyma_project
  display_name = "Kyma Restricted Registry PROD Write Access"
  description  = "Read-write access to Kyma Restricted Registry PROD environment"

  group_key {
    id = "kyma-restricted-registry-prod-write@sap.com"
  }

  labels = {
    "cloudidentity.googleapis.com/groups.security" = ""
  }

  parent = "customers/${data.google_organization.sap.directory_customer_id}"
}

resource "google_cloud_identity_group" "restricted_registry_dev_read" {
  provider     = google.kyma_project
  display_name = "Kyma Restricted Registry DEV Read Access"
  description  = "Read-only access to Kyma Restricted Registry DEV environment"

  group_key {
    id = "kyma-restricted-registry-dev-read@sap.com"
  }

  labels = {
    "cloudidentity.googleapis.com/groups.security" = ""
  }

  parent = "customers/${data.google_organization.sap.directory_customer_id}"
}

resource "google_cloud_identity_group" "restricted_registry_dev_write" {
  provider     = google.kyma_project
  display_name = "Kyma Restricted Registry DEV Write Access"
  description  = "Read-write access to Kyma Restricted Registry DEV environment"

  group_key {
    id = "kyma-restricted-registry-dev-write@sap.com"
  }

  labels = {
    "cloudidentity.googleapis.com/groups.security" = ""
  }

  parent = "customers/${data.google_organization.sap.directory_customer_id}"
}
