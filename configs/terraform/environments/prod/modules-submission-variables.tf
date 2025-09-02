variable "dev_kyma_modules_repository" {
  type = object({
    name                       = string
    description                = string
    repository_prevent_destroy = bool
  })
  default = {
    name                       = "dev-kyma-modules"
    description                = "Development Kyma modules"
    repository_prevent_destroy = false
  }
}

variable "kyma_modules_repository" {
  type = object({
    name                       = string
    description                = string
    type                       = string
    reader_serviceaccounts     = list(string)
    reader_groups              = list(string)
    repository_prevent_destroy = bool
  })
  default = {
    name        = "kyma-modules"
    description = "Production Kyma modules"
    type        = "production"
    reader_serviceaccounts = [
      "klm-controller-manager@sap-ti-dx-kyma-mps-dev.iam.gserviceaccount.com",
      "klm-controller-manager@sap-ti-dx-kyma-mps-stage.iam.gserviceaccount.com",
      "klm-controller-manager@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com"
    ]
    reader_groups = [
      "cam_dx_kyma_gcp_sre@sap.com"
    ]
    repository_prevent_destroy = true
  }
}
