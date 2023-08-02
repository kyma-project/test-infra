gcp_region = "europe-west4"
artifact_registry_list = [
  {
    name                   = "modules-internal"
    owner                  = "neighbors"
    type                   = "development"
    reader_serviceaccounts = ["klm-controller-manager@sap-ti-dx-kyma-mps-dev.iam.gserviceaccount.com", "klm-controller-manager@sap-ti-dx-kyma-mps-stage.iam.gserviceaccount.com", "klm-controller-manager@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com"]
  },
]