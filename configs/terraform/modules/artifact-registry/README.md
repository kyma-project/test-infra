# Artifact Registry Module

Artifact Registry module for GCP is used to create and manage Artifact Registry in GCP. 
It is designed to maintain a standarised and reusable way of creating Artifact Registry in GCP.
Module ensures that all the necessary resources are created and configured correctly, including identities and IAM roles.

## Usage

Configure single Artifact Registry per module call directly in `.tf` file.
Refer to `variables.tf` for all the variables that can be configured and their default values.
Refer to `outputs.tf` for all the outputs that are available after the module is created.

> **CAUTION:** Carefully review the planned changes before applying them, especially when using the module to configure production environments.

When you use the GCP private image registry, consider the following:

- The solution is prepared for the GCP Service Account related execution.
- The **roles/artifactregistry.repoAdmin** role binding is part of the solution. To learn more, read [Artifact Registry Repository Access Control](https://cloud.google.com/artifact-registry/docs/access-control).
- The **roles/artifactregistry.reader** role binding is required for lifecycle-manager service accounts. To learn more, read [Artifact Registry Repository Access Control](https://cloud.google.com/artifact-registry/docs/access-control).
- You can make your repository public if you use the `public = true` in the module call.
- Vulnerability scanning is enabled by default.

## Example

```hcl
module "docker_repository" {
  source = "../../modules/artifact-registry"

  providers = {
    google = google.kyma_project
  }

  repository_name        = var.docker_repository.name
  description            = var.docker_repository.description
  location               = var.docker_repository.location
  immutable_tags         = var.docker_repository.immutable_tags
  format                 = var.docker_repository.format
  cleanup_policies       = var.docker_repository.cleanup_policies
  cleanup_policy_dry_run = var.docker_repository.cleanup_policy_dry_run
}
```