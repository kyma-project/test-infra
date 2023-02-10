variable "k8s_terraform_sa" {
  type = object({
    name      = string
    namespace = string
  })
  description = ""
}

resource "kubernetes_secret" "terraform_sa" {
  metadata {
    name = "terraform-service-account"
  }
}


resource "kubernetes_service_account" "terraform" {
  metadata {
    namespace = var.k8s_terraform_sa.namespace
    name      = "terraform"
    annotations = {
      "iam.gke.io/gcp-service-account" = "prow-terraform-executor@sap-kyma-prow.iam.gserviceaccount.com"
    }
  }
  secret {
    name = kubernetes_secret.terraform_sa.metadata[0].name
  }
  automount_service_account_token = true
}
