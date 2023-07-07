# gcp-artifact-registry-terraform

This is the GCP private image registry. Use the registry to publish modules that should be accessible by internal SAP teams.

## Usage

Configure variables in `terraform.tfvars` (configs/terraform/environments/prod/terraform.tfvars) file

These are the default values:

```terraform
artifact_registry_owner          = "neighbors"
artifact_registry_type           = "development"
immutable_artifact_registry      = false
artifact_registry_multi_region   = true
artifact_registry_primary_area  = "europe"
```

When you use the GCP private image registry, consider the following: 

- The solution is prepared for GCP Service Account related execution
- `roles/artifactregistry.write` role binding is part of the solution ([Artifact Registry Repository Access Control](https://cloud.google.com/artifact-registry/docs/access-control))
- Vulnerability scanning is enabled by default.
- You must define `artifact_registry_serviceaccount` in the .tfvars file.
