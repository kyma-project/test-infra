terraform {
  required_version = ">= 1.6.1"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 6.0.1"
    }

    google-beta = {
      source  = "hashicorp/google-beta"
      version = ">= 6.0.1"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "3.0.1"
    }
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0.0"
    }
    github = {
      source  = "integrations/github"
      version = "~> 6.9.0"
    }
  }
}


# Provider for public GitHub (github.com) - kyma-project organization
provider "github" {
  alias = "kyma_project"
  owner = var.kyma_project_github_org
}

# ------------------------------------------------------------------------------
# Internal GitHub Enterprise Provider
# ------------------------------------------------------------------------------
# This provider configuration enables Terraform to manage resources in SAP's
# internal GitHub Enterprise instance.
#
# Authentication:
# - The token is passed via TF_VAR_internal_github_token environment variable
# - The token is retrieved from GCP Secret Manager during workflow execution
# - For terraform plan: uses the planner token (read-only operations)
# - For terraform apply: uses the executor token (write operations)
# ------------------------------------------------------------------------------
provider "github" {
  alias = "internal_github"
  owner = var.internal_github_organization_name
  # Token is provided via TF_VAR_internal_github_token environment variable from GitHub Actions workflow
  token    = var.internal_github_token
  base_url = var.internal_github_base_url
}

# sap-kyma-prow project provider
provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

provider "google" {
  alias   = "kyma_project"
  project = var.kyma_project_gcp_project_id
  region  = var.kyma_project_gcp_region
}

provider "google-beta" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

# data.google_client_config configures Google Cloud client.
data "google_client_config" "gcp" {
}
