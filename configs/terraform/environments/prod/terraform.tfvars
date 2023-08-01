gcp_region                               = "europe-west4"
artifact_registry_module                 = "manifest-repository"
artifact_registry_owner                  = "neighbors"
artifact_registry_type                   = "production"
artifact_registry_multi_region           = true
artifact_registry_reader_serviceaccounts = ["klm-controller-manager@sap-ti-dx-kyma-mps-dev.iam.gserviceaccount.com", "klm-controller-manager@sap-ti-dx-kyma-mps-stage.iam.gserviceaccount.com", "klm-controller-manager@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com"]
artifact_registry_writer_serviceaccount  = ""