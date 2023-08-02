# gcp-artifact-registry-terraform

This is the GCP private image registry. Use the registry to publish modules that should be accessible by internal SAP teams.

## Usage

Configure variables in the `terraform.tfvars` file.

These are the default values:

```terraform
artifact_registry_owner          = "neighbors"
artifact_registry_type           = "development"
artifact_registry_multi_region   = true
artifact_registry_primary_area  = "europe"
```

When you use the GCP private image registry, consider the following:

- The solution is prepared for the GCP Service Account related execution.
- The `roles/artifactregistry.writer` role binding is part of the solution. To learn more, read [Artifact Registry Repository Access Control](https://cloud.google.com/artifact-registry/docs/access-control). If this variable is empty, the solution won't add any service account with the `writer` permission.
- The `roles/artifactregistry.reader` role binding is required for lifecycle-manager service accounts. To learn more, read [Artifact Registry Repository Access Control](https://cloud.google.com/artifact-registry/docs/access-control).
- You can make your repository public if you use the `artifact_registry_public = true` in the `terraform.tfvars`.
- Vulnerability scanning is enabled by default.