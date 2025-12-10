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
      version = "2.38.0"
    }
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0.0"
    }
    github = {
      source  = "integrations/github"
      version = "~> 6.8.0"
    }
  }
}


# Provider for public GitHub (github.com) - kyma-project organization
provider "github" {
  alias = "kyma_project"
  owner = var.kyma_project_github_org
}

# ------------------------------------------------------------------------------
# GitHub Enterprise Provider (github.tools.sap)
# ------------------------------------------------------------------------------
# This provider configuration enables Terraform to manage resources in SAP's
# internal GitHub Enterprise instance (github.tools.sap).
#
# Authentication:
# - The token is passed via TF_VAR_github_tools_sap_token environment variable
# - The token is retrieved from GCP Secret Manager during workflow execution
# - For terraform plan: uses the planner token (read-only operations)
# - For terraform apply: uses the executor token (write operations)
# ------------------------------------------------------------------------------
provider "github" {
  alias = "github_tools_sap"
  owner = var.github_tools_sap_organization_name
  # Token is provided via TF_VAR_github_tools_sap_token environment variable from GitHub Actions workflow
  token = var.github_tools_sap_token
  # Base URL is set to github.tools.sap API endpoint for GitHub Enterprise
  base_url = "https://github.tools.sap/api/v3"
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
