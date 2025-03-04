artifact_registry_gcp_project_id = "smart-tractor-389208"
artifact_registry_gcp_region     = "europe-west4"
artifact_registry_collection = {
  modules-internal = {
    name                   = "modules-internal"
    owner                  = "neighbors"
    type                   = "development"
    reader_serviceaccounts = ["klm-controller-manager@sap-ti-dx-kyma-mps-dev.iam.gserviceaccount.com", "klm-controller-manager@sap-ti-dx-kyma-mps-stage.iam.gserviceaccount.com", "klm-controller-manager@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com"]
    writer_serviceaccounts = ["kyma-submission-pipeline@kyma-project.iam.gserviceaccount.com"]
  },
}# (2025-03-04)