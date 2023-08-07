gcp_region = "europe-west4"
artifact_registry_list = [
  {
    name                   = "modules-internal"
    owner                  = "neighbors"
    type                   = "production"
    reader_serviceaccounts = ["klm-controller-manager@sap-ti-dx-kyma-mps-dev.iam.gserviceaccount.com", "klm-controller-manager@sap-ti-dx-kyma-mps-stage.iam.gserviceaccount.com", "klm-controller-manager@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com"]
    writer_serviceaccount  = "kyma-submission-pipeline@kyma-project.iam.gserviceaccount.com"
  }
]